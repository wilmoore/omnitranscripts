//go:build encore

// Package transcribe provides YouTube video transcription services.
// This package requires Encore.dev and is only built when the "encore" build tag is set.
package transcribe

import (
	"context"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/config"
	"encore.dev/pubsub"
	"encore.dev/rlog"

	"omnitranscripts/lib"
	"omnitranscripts/models"
)

// Config holds the service configuration.
var cfg = config.Load[Config]()

type Config struct {
	APIKey         string               `json:"api_key"`
	WorkDir        string               `json:"work_dir"`
	MaxVideoLength int                  `json:"max_video_length"`
	FreeJobLimit   int                  `json:"free_job_limit"`
	WebhookURL     string               `json:"webhook_url"`
	WebhookSecret  string               `json:"webhook_secret"`
	WebhookEvents  []string             `json:"webhook_events"`
}

// TranscribeRequest represents a transcription request.
type TranscribeRequest struct {
	URL string `json:"url"`
}

// TranscribeResponse represents the response from a transcription request.
type TranscribeResponse struct {
	JobID      string         `json:"job_id,omitempty"`
	Transcript string         `json:"transcript,omitempty"`
	Segments   []models.Segment `json:"segments,omitempty"`
}

// JobStatusResponse represents the response for job status queries.
type JobStatusResponse struct {
	ID            string           `json:"id"`
	Status        string           `json:"status"`
	Transcript    string           `json:"transcript,omitempty"`
	Segments      []models.Segment `json:"segments,omitempty"`
	Error         string           `json:"error,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	CompletedAt   *time.Time       `json:"completed_at,omitempty"`
	SubtitleFiles *SubtitleFiles   `json:"subtitle_files,omitempty"`
}

type SubtitleFiles struct {
	SRTURL string `json:"srt_url,omitempty"`
	VTTURL string `json:"vtt_url,omitempty"`
}

// Health endpoint that doesn't require authentication.
//
//encore:api public method=GET path=/health
func Health(ctx context.Context) (*HealthResponse, error) {
	return &HealthResponse{
		Status:  "ok",
		Message: "OmniTranscripts API is running",
	}, nil
}

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Transcribe transcribes a YouTube video.
//
//encore:api auth method=POST path=/transcribe
func Transcribe(ctx context.Context, req *TranscribeRequest) (*TranscribeResponse, error) {
	rlog.Info("transcribe request", "url", req.URL)

	// Validate URL
	if !models.ValidateURL(req.URL) {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Invalid YouTube URL",
		}
	}

	// Get video duration to determine processing strategy
	duration, err := lib.GetVideoDuration(req.URL)
	if err != nil {
		rlog.Error("failed to get video duration", "error", err, "url", req.URL)
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Failed to get video information",
		}
	}

	// Create job
	job := models.NewJob(req.URL)

	// For short videos (≤2 min), process synchronously
	if duration <= 120 {
		rlog.Info("processing video synchronously", "duration", duration, "job_id", job.ID)

		transcript, segments, err := lib.ProcessTranscription(req.URL, job.ID)
		if err != nil {
			rlog.Error("transcription failed", "error", err, "job_id", job.ID)
			return nil, &errs.Error{
				Code:    errs.Internal,
				Message: "Transcription failed",
			}
		}

		return &TranscribeResponse{
			Transcript: transcript,
			Segments:   segments,
		}, nil
	}

	// For longer videos, queue for async processing
	rlog.Info("queueing video for async processing", "duration", duration, "job_id", job.ID)

	// Store job in database and publish to queue
	if err := storeJob(ctx, job); err != nil {
		return nil, err
	}

	// Publish job to processing queue
	if err := publishJob(ctx, job); err != nil {
		return nil, err
	}

	return &TranscribeResponse{
		JobID: job.ID,
	}, nil
}

// GetJob retrieves the status and result of a transcription job.
//
//encore:api auth method=GET path=/transcribe/:id
func GetJob(ctx context.Context, id string) (*JobStatusResponse, error) {
	job, err := getJob(ctx, id)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Job not found",
		}
	}

	response := &JobStatusResponse{
		ID:        job.ID,
		Status:    string(job.Status),
		CreatedAt: job.CreatedAt,
	}

	if job.Status == models.StatusComplete {
		response.Transcript = job.Transcript
		response.Segments = job.Segments
		response.CompletedAt = job.CompletedAt
	} else if job.Status == models.StatusError {
		response.Error = job.Error
		response.CompletedAt = job.CompletedAt
	}

	return response, nil
}

// AuthHandler validates API key authentication.
//
//encore:authhandler
func AuthHandler(ctx context.Context, token string) (auth.UID, error) {
	if token != cfg.APIKey {
		return "", &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "Invalid API key",
		}
	}
	return auth.UID("authenticated"), nil
}

// Topic for job processing
var jobTopic = pubsub.NewTopic[*models.Job]("job-processing", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

// publishJob publishes a job to the processing queue.
func publishJob(ctx context.Context, job *models.Job) error {
	_, err := jobTopic.Publish(ctx, job)
	return err
}

// Subscribe to job processing
var _ = pubsub.NewSubscription(jobTopic, "process-jobs", pubsub.SubscriptionConfig[*models.Job]{
	Handler: processJobAsync,
})

// processJobAsync processes a job asynchronously.
func processJobAsync(ctx context.Context, job *models.Job) error {
	startTime := time.Now()
	rlog.Info("processing job async", "job_id", job.ID, "url", job.URL)

	// Initialize webhook manager if configured
	var webhookManager *lib.WebhookManager
	if cfg.WebhookURL != "" {
		webhookConfig := lib.WebhookConfig{
			URL:     cfg.WebhookURL,
			Events:  cfg.WebhookEvents,
			Timeout: 10 * time.Second,
			Retries: 3,
		}
		if cfg.WebhookSecret != "" {
			webhookConfig.Headers = map[string]string{
				"X-Webhook-Secret": cfg.WebhookSecret,
			}
		}
		webhookManager = lib.NewWebhookManager(webhookConfig)

		// Send job started webhook
		webhookManager.SendJobStarted(ctx, job)
	}

	// Mark job as running
	job.MarkRunning()
	if err := updateJob(ctx, job); err != nil {
		return err
	}

	// Process transcription
	transcript, segments, err := lib.ProcessTranscription(job.URL, job.ID)
	if err != nil {
		processingTime := time.Since(startTime)
		rlog.Error("async transcription failed", "error", err, "job_id", job.ID)
		job.MarkError(err)
		updateJob(ctx, job)

		// Send failure webhook
		if webhookManager != nil {
			webhookManager.SendJobFailed(ctx, job, err.Error(), processingTime)
		}
		return err
	}

	// Generate subtitle files
	var srtPath, vttPath string
	if len(segments) > 0 {
		outputDir := cfg.WorkDir
		srtPath, vttPath, err = lib.GenerateSubtitles(segments, outputDir, job.ID)
		if err != nil {
			rlog.Error("subtitle generation failed", "error", err, "job_id", job.ID)
			// Don't fail the job for subtitle errors, just log
		} else {
			rlog.Info("subtitles generated", "job_id", job.ID, "srt", srtPath, "vtt", vttPath)
		}
	}

	// Mark job as complete
	job.MarkComplete(transcript, segments)
	if err := updateJob(ctx, job); err != nil {
		return err
	}

	// Send completion webhook
	if webhookManager != nil {
		processingTime := time.Since(startTime)
		webhookManager.SendJobCompleted(ctx, job, srtPath, vttPath, processingTime)
	}

	rlog.Info("job completed successfully", "job_id", job.ID, "processing_time", time.Since(startTime))
	return nil
}