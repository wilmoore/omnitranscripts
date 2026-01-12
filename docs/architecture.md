# Architecture Documentation

OmniTranscripts is designed as a high-performance, scalable universal media transcription engine built with Go. It supports 1000+ platforms via yt-dlp with audio-first workflows as first-class citizens.

## Overview

The application follows a microservices-inspired architecture with clear separation of concerns, supporting both traditional HTTP servers and the Encore.dev framework for production deployments.

## System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client Apps   │    │  Load Balancer  │    │   API Gateway   │
│                 │    │                 │    │                 │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
         ┌─────────────────────────────────────────────────────┐
         │              OmniTranscripts API                      │
         └─────────────────────┬───────────────────────────────┘
                               │
    ┌──────────────────────────┼──────────────────────────────┐
    │                          │                              │
┌───▼────┐  ┌─────────┐  ┌─────▼────┐  ┌──────────┐  ┌──────▼───┐
│ Fiber  │  │ Encore  │  │ Job Queue│  │Database  │  │ File     │
│ Server │  │ Service │  │(In-Memory│  │(PostgreSQL│  │ Storage  │
│        │  │         │  │/Redis)   │  │)         │  │          │
└────────┘  └─────────┘  └──────────┘  └──────────┘  └──────────┘
                               │
         ┌─────────────────────────────────────────────────────┐
         │           Processing Pipeline                        │
         └─────────────────────┬───────────────────────────────┘
                               │
    ┌──────────────────────────┼──────────────────────────────┐
    │                          │                              │
┌───▼────┐     ┌─────────┐     ┌─────▼────┐     ┌──────────┐
│ yt-dlp │     │ FFmpeg  │     │ whisper  │     │ Subtitle │
│Download│────▶│ Audio   │────▶│ .cpp     │────▶│Generator │
│        │     │Extract  │     │Transcribe│     │          │
└────────┘     └─────────┘     └──────────┘     └──────────┘
```

## Core Components

### 1. HTTP Server Layer

#### Fiber Server (`main.go`)
- **Purpose**: Alternative HTTP implementation using Fiber framework
- **Features**: Fast HTTP routing, middleware support, CORS handling
- **Use Case**: Development, lightweight deployments, custom hosting

#### Encore.dev Service (`transcribe/service.go`)
- **Purpose**: Production-ready API with built-in infrastructure
- **Features**: Auto-scaling, monitoring, database management, pub/sub
- **Use Case**: Production deployments, enterprise features

### 2. Job Processing System

#### Job Management (`jobs/`)
```go
type Job struct {
    ID          string    `json:"id"`
    URL         string    `json:"url"`
    Status      Status    `json:"status"`
    Progress    int       `json:"progress"`
    CreatedAt   time.Time `json:"created_at"`
    CompletedAt *time.Time`json:"completed_at,omitempty"`
    Transcript  string    `json:"transcript,omitempty"`
    Segments    []Segment `json:"segments,omitempty"`
    Error       string    `json:"error,omitempty"`
}
```

#### Processing States
1. **Pending**: Job created, waiting for processing
2. **Running**: Currently being processed
3. **Complete**: Successfully processed with results
4. **Error**: Failed with error message

#### Queue Implementation
- **In-Memory**: Thread-safe map with mutex protection (development)
- **Redis**: Distributed queue for production scaling
- **Pub/Sub**: Encore.dev topic-based messaging for async processing

### 3. Processing Pipeline

#### Stage 1: Video Download
```go
// lib/transcription.go
func downloadVideo(url, workDir string) (string, error) {
    cmd := exec.Command("yt-dlp",
        "--extract-audio",
        "--audio-format", "wav",
        "--output", filepath.Join(workDir, "%(title)s.%(ext)s"),
        url)
    return cmd.Output()
}
```

**Technology**: `yt-dlp` (YouTube download tool)
**Output**: Raw audio file in various formats

#### Stage 2: Audio Normalization
```go
func normalizeAudio(inputPath, outputPath string) error {
    return ffmpeg.Input(inputPath).
        Filter("aresample", ffmpeg.Args{"16000"}).
        Filter("ac", ffmpeg.Args{"1"}).
        Output(outputPath, ffmpeg.KwArgs{"f": "wav"}).
        Run()
}
```

**Technology**: `FFmpeg` via `ffmpeg-go` wrapper
**Output**: 16kHz mono WAV file optimized for whisper.cpp

#### Stage 3: AI Transcription
```go
func transcribeAudio(audioPath string) ([]Segment, error) {
    cmd := exec.Command("whisper.cpp",
        "--model", "models/ggml-base.en.bin",
        "--output-json",
        "--output-srt",
        "--output-vtt",
        audioPath)
    return parseWhisperOutput(cmd.Output())
}
```

**Technology**: `whisper.cpp` (OpenAI Whisper C++ implementation)
**Output**: Timestamped transcript segments

#### Stage 4: Subtitle Generation
```go
func generateSubtitles(segments []Segment, outputDir string) error {
    // Generate SRT format
    generateSRT(segments, filepath.Join(outputDir, "subtitles.srt"))

    // Generate VTT format
    generateVTT(segments, filepath.Join(outputDir, "subtitles.vtt"))

    // Generate JSON format
    generateJSON(segments, filepath.Join(outputDir, "transcript.json"))

    return nil
}
```

**Output Formats**:
- SRT: Standard subtitle format
- VTT: WebVTT for web players
- JSON: Structured data with timestamps
- TSV: Tab-separated for analysis

### 4. Data Layer

#### Database Schema (PostgreSQL)

**Jobs Table**:
```sql
CREATE TABLE jobs (
    id TEXT PRIMARY KEY,
    url TEXT NOT NULL,
    video_id TEXT,
    title TEXT,
    status TEXT NOT NULL,
    progress INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    transcript TEXT,
    segments JSONB,
    error TEXT,
    duration TEXT,
    file_count INTEGER DEFAULT 0,
    file_size TEXT,
    log_file TEXT,
    output_dir TEXT
);
```

**Metrics Tables**:
- `system_metrics`: Real-time performance data
- `job_performance`: Processing time tracking
- `api_usage`: Request analytics
- `business_metrics`: Revenue and usage statistics

#### File Storage
```
transcripts/
├── {video_id}/
│   ├── metadata.json
│   ├── transcript.txt
│   ├── transcript.json
│   ├── subtitles.srt
│   ├── subtitles.vtt
│   └── subtitles.tsv
```

## Scalability Patterns

### 1. Horizontal Scaling

#### Load Balancing
- Multiple API server instances
- Database connection pooling
- Shared Redis queue for job distribution

#### Worker Scaling
```go
// Configure worker pool size
const MaxWorkers = 10

func startWorkerPool() {
    for i := 0; i < MaxWorkers; i++ {
        go func(workerID int) {
            for job := range jobQueue {
                processJob(job, workerID)
            }
        }(i)
    }
}
```

### 2. Async Processing Strategy

#### Sync vs Async Decision
```go
func determineProcessingMode(duration int) ProcessingMode {
    if duration <= 120 { // 2 minutes
        return SynchronousMode
    }
    return AsynchronousMode
}
```

**Synchronous** (≤2 min):
- Immediate processing
- Real-time response
- Better user experience for short content

**Asynchronous** (>2 min):
- Background processing
- Job queue with status polling
- Scales better for long content

### 3. Caching Strategy

#### Video Metadata Caching
```go
type CacheEntry struct {
    VideoID     string        `json:"video_id"`
    Title       string        `json:"title"`
    Duration    time.Duration `json:"duration"`
    Transcript  string        `json:"transcript,omitempty"`
    CachedAt    time.Time     `json:"cached_at"`
    ExpiresAt   time.Time     `json:"expires_at"`
}
```

#### Cache Layers
1. **In-Memory**: Recent transcriptions (LRU cache)
2. **Redis**: Distributed cache for multiple instances
3. **Database**: Persistent storage for completed jobs

## Performance Optimizations

### 1. Processing Pipeline Optimizations

#### Parallel Processing
```go
func processJobParallel(job *Job) error {
    var wg sync.WaitGroup
    errors := make(chan error, 3)

    // Download video
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := downloadVideo(job.URL); err != nil {
            errors <- err
        }
    }()

    // Process in pipeline stages
    wg.Wait()
    close(errors)

    for err := range errors {
        if err != nil {
            return err
        }
    }

    return nil
}
```

#### Resource Management
- Automatic cleanup of temporary files
- Memory-mapped file processing for large audio files
- Goroutine pools to prevent resource exhaustion

### 2. Database Optimizations

#### Indexing Strategy
```sql
-- Performance indexes
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_created_at ON jobs(created_at DESC);
CREATE INDEX idx_jobs_video_id ON jobs(video_id);

-- Composite indexes for common queries
CREATE INDEX idx_jobs_status_created ON jobs(status, created_at DESC);
```

#### Query Optimization
```go
// Efficient job status queries
func getJobsByStatus(status string, limit int) ([]Job, error) {
    query := `
        SELECT id, status, progress, created_at, updated_at
        FROM jobs
        WHERE status = $1
        ORDER BY created_at DESC
        LIMIT $2`

    return db.Query(query, status, limit)
}
```

## Security Architecture

### 1. Authentication & Authorization

#### API Key Management
```go
type APIKey struct {
    ID          string    `json:"id"`
    Key         string    `json:"key"`
    Name        string    `json:"name"`
    Permissions []string  `json:"permissions"`
    CreatedAt   time.Time `json:"created_at"`
    ExpiresAt   *time.Time`json:"expires_at,omitempty"`
    LastUsed    *time.Time`json:"last_used,omitempty"`
}
```

#### Rate Limiting
```go
type RateLimiter struct {
    requests map[string][]time.Time
    mutex    sync.RWMutex
    limit    int
    window   time.Duration
}

func (rl *RateLimiter) Allow(apiKey string) bool {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()

    now := time.Now()
    windowStart := now.Add(-rl.window)

    // Clean old requests
    requests := rl.requests[apiKey]
    validRequests := []time.Time{}

    for _, req := range requests {
        if req.After(windowStart) {
            validRequests = append(validRequests, req)
        }
    }

    if len(validRequests) >= rl.limit {
        return false
    }

    validRequests = append(validRequests, now)
    rl.requests[apiKey] = validRequests

    return true
}
```

### 2. Input Validation & Sanitization

#### URL Validation
```go
func validateYouTubeURL(url string) error {
    // Regex pattern for YouTube URLs
    pattern := `^https?://(www\.)?(youtube\.com/watch\?v=|youtu\.be/)[a-zA-Z0-9_-]{11}`

    matched, err := regexp.MatchString(pattern, url)
    if err != nil {
        return err
    }

    if !matched {
        return errors.New("invalid YouTube URL format")
    }

    return nil
}
```

#### Content Security
- Sandboxed processing environment
- Resource limits for external tools (yt-dlp, ffmpeg)
- Automatic cleanup of sensitive temporary files

## Monitoring & Observability

### 1. Metrics Collection

#### Application Metrics
```go
type Metrics struct {
    JobsTotal       prometheus.Counter
    JobsCompleted   prometheus.Counter
    JobsFailed      prometheus.Counter
    ProcessingTime  prometheus.Histogram
    QueueSize       prometheus.Gauge
}
```

#### Business Metrics
- Transcription success rate
- Average processing time by video length
- API usage patterns
- Revenue tracking

### 2. Logging Strategy

#### Structured Logging
```go
func logJobProgress(jobID, stage string, progress int) {
    log.WithFields(logrus.Fields{
        "job_id":  jobID,
        "stage":   stage,
        "progress": progress,
        "timestamp": time.Now(),
    }).Info("Job progress update")
}
```

#### Log Levels
- **Error**: Failed jobs, system errors
- **Warn**: Retries, timeouts, degraded performance
- **Info**: Job lifecycle, API requests
- **Debug**: Detailed processing steps (development only)

## Deployment Architecture

### 1. Container Strategy

#### Multi-Stage Docker Build
```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o videotranscript-app

# Runtime stage
FROM alpine:latest
RUN apk add --no-cache ffmpeg python3 py3-pip
RUN pip3 install yt-dlp
# Install whisper.cpp...
COPY --from=builder /app/videotranscript-app .
```

### 2. Encore.dev Production Deployment

#### Infrastructure as Code
- Automatic provisioning of cloud resources
- Built-in load balancing and auto-scaling
- Managed database and pub/sub services
- Zero-downtime deployments

#### Configuration Management
```go
type Config struct {
    APIKey         string `json:"api_key"`
    WorkDir        string `json:"work_dir"`
    MaxVideoLength int    `json:"max_video_length"`
    FreeJobLimit   int    `json:"free_job_limit"`
    DatabaseURL    string `json:"database_url"`
    RedisURL       string `json:"redis_url"`
}
```

This architecture provides a solid foundation for scaling from development to enterprise-level deployments while maintaining performance, security, and reliability.