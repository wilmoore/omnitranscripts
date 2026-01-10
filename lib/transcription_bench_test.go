package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"omnitranscripts/jobs"
	"omnitranscripts/models"
)

func BenchmarkGetVideoDuration(b *testing.B) {
	testURL := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = GetVideoDuration(testURL)
	}
}

func BenchmarkValidateURL(b *testing.B) {
	testURLs := []string{
		// YouTube
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		"https://youtu.be/dQw4w9WgXcQ",
		// Other supported platforms
		"https://vimeo.com/123456789",
		"https://www.dailymotion.com/video/x123abc",
		"https://twitter.com/user/status/123456789",
		"https://www.tiktok.com/@user/video/123456789",
		"https://www.twitch.tv/videos/123456789",
		// Invalid URLs
		"not-a-url",
		"ftp://example.com/file",
		"",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, url := range testURLs {
			models.ValidateURL(url)
		}
	}
}

func BenchmarkJobCreation(b *testing.B) {
	testURL := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		job := jobs.NewJob(testURL)
		_ = job
	}
}

func BenchmarkJobStatusTransitions(b *testing.B) {
	testURL := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	segments := []models.Segment{
		{Start: 0.0, End: 5.0, Text: "Test segment 1"},
		{Start: 5.0, End: 10.0, Text: "Test segment 2"},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		job := jobs.NewJob(testURL)
		job.MarkRunning()
		job.MarkComplete("Test transcript", segments)
	}
}

func BenchmarkTranscriptParsing(b *testing.B) {
	// Create a temporary test transcript file
	tempDir := b.TempDir()
	transcriptFile := filepath.Join(tempDir, "test_transcript.txt")

	testContent := `[00:00:00.000 --> 00:00:05.000] Hello, this is a test transcript
[00:00:05.000 --> 00:00:10.000] This is the second segment
[00:00:10.000 --> 00:00:15.000] And this is the third segment
[00:00:15.000 --> 00:00:20.000] Final segment for testing
`

	err := os.WriteFile(transcriptFile, []byte(testContent), 0644)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _ = models.LoadTranscript(transcriptFile)
	}
}

func BenchmarkConcurrentJobProcessing(b *testing.B) {
	jobs.Initialize()
	queue := jobs.GetQueue()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			job := jobs.NewJob("https://www.youtube.com/watch?v=dQw4w9WgXcQ")
			queue.AddJob(job)

			// Simulate processing time
			job.MarkRunning()
			queue.UpdateJob(job)

			// Simulate completion
			job.MarkComplete("Benchmark transcript", []models.Segment{
				{Start: 0.0, End: 5.0, Text: "Benchmark test"},
			})
			queue.UpdateJob(job)
		}
	})
}

// Memory benchmarks
func BenchmarkMemoryUsage_JobQueue(b *testing.B) {
	jobs.Initialize()
	queue := jobs.GetQueue()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		job := jobs.NewJob("https://www.youtube.com/watch?v=dQw4w9WgXcQ")
		queue.AddJob(job)
	}
}

func BenchmarkMemoryUsage_LargeTranscript(b *testing.B) {
	// Create a large transcript with many segments
	var segments []models.Segment
	for i := 0; i < 1000; i++ {
		segments = append(segments, models.Segment{
			Start: float64(i * 5),
			End:   float64((i + 1) * 5),
			Text:  "This is a test segment with some text content that simulates a real transcript",
		})
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		job := jobs.NewJob("https://www.youtube.com/watch?v=dQw4w9WgXcQ")
		job.MarkComplete("Large transcript content", segments)
	}
}

// Context and timeout benchmarks
func BenchmarkContextOperations(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		select {
		case <-ctx.Done():
			b.Error("Context should not be done")
		default:
			// Context is still active
		}
		cancel()
	}
}

func BenchmarkFileOperations(b *testing.B) {
	tempDir := b.TempDir()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate temporary file creation and cleanup
		audioFile := filepath.Join(tempDir, "test_audio.wav")
		normalizedAudio := filepath.Join(tempDir, "test_normalized.wav")
		transcriptFile := filepath.Join(tempDir, "test_transcript.txt")

		// Create dummy files
		os.WriteFile(audioFile, []byte("dummy audio content"), 0644)
		os.WriteFile(normalizedAudio, []byte("dummy normalized content"), 0644)
		os.WriteFile(transcriptFile, []byte("dummy transcript content"), 0644)

		// Clean up
		os.Remove(audioFile)
		os.Remove(normalizedAudio)
		os.Remove(transcriptFile)
	}
}

// Performance comparison benchmarks
func BenchmarkStringOperations_Transcript(b *testing.B) {
	segments := []models.Segment{
		{Start: 0.0, End: 5.0, Text: "First segment"},
		{Start: 5.0, End: 10.0, Text: "Second segment"},
		{Start: 10.0, End: 15.0, Text: "Third segment"},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var transcript string
		for _, segment := range segments {
			transcript += segment.Text + " "
		}
	}
}

func BenchmarkStringBuilder_Transcript(b *testing.B) {
	segments := []models.Segment{
		{Start: 0.0, End: 5.0, Text: "First segment"},
		{Start: 5.0, End: 10.0, Text: "Second segment"},
		{Start: 10.0, End: 15.0, Text: "Third segment"},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var builder strings.Builder
		for _, segment := range segments {
			builder.WriteString(segment.Text + " ")
		}
		_ = builder.String()
	}
}

func BenchmarkJSONSerialization_Job(b *testing.B) {
	job := jobs.NewJob("https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	job.MarkComplete("Test transcript", []models.Segment{
		{Start: 0.0, End: 5.0, Text: "Test segment"},
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(job)
	}
}

// Resource cleanup benchmark
func BenchmarkResourceCleanup(b *testing.B) {
	tempDir := b.TempDir()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create multiple temporary files
		files := make([]string, 3)
		for j := 0; j < 3; j++ {
			files[j] = filepath.Join(tempDir, fmt.Sprintf("temp_file_%d_%d.tmp", i, j))
			os.WriteFile(files[j], []byte("temporary content"), 0644)
		}

		// Cleanup all files
		for _, file := range files {
			os.Remove(file)
		}
	}
}
