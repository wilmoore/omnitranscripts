// CLI tool for transcribing any URL or local file using OmniTranscripts
//
// Supports:
//   - YouTube, Instagram, TikTok, Vimeo, and 1000+ platforms via yt-dlp
//   - Local audio/video files (mp4, mp3, wav, etc.)
//
// Usage:
//
//	go run main.go <url_or_file_path>
//	make transcribe URL="https://youtube.com/watch?v=..."
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"omnitranscripts/engine"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	input := os.Args[1]

	// Determine if input is a URL or local file
	isURL := strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://")

	if !isURL {
		// Check if local file exists
		if _, err := os.Stat(input); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: File not found: %s\n", input)
			os.Exit(1)
		}
	}

	// Print progress to stderr (not part of transcript output)
	fmt.Fprintf(os.Stderr, "Transcribing: %s\n", input)
	if isURL {
		fmt.Fprintf(os.Stderr, "Type: URL (downloading via yt-dlp)\n")
	} else {
		fmt.Fprintf(os.Stderr, "Type: Local file\n")
	}
	fmt.Fprintf(os.Stderr, "\n")

	// Create a context with timeout for the transcription
	// ADR-0003: Context propagation with appropriate timeouts
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	startTime := time.Now()

	// Use engine.Transcribe - the public library interface (ADR-0001)
	// URLs go through: yt-dlp download -> FFmpeg normalize -> Whisper transcribe
	// Local files go through: FFmpeg normalize -> Whisper transcribe
	opts := engine.DefaultOptions()
	opts.CacheDownloads = true // Cache downloads for CLI usage
	result, err := engine.Transcribe(ctx, input, "cli-transcribe", opts)
	if err != nil {
		// Provide stage-specific error context to stderr
		if tErr, ok := err.(*engine.TranscriptionError); ok {
			fmt.Fprintf(os.Stderr, "Transcription failed at stage '%s': %s\n", tErr.Stage, tErr.Message)
			if tErr.Err != nil {
				fmt.Fprintf(os.Stderr, "  Cause: %v\n", tErr.Err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Transcription failed: %v\n", err)
		}
		os.Exit(1)
	}

	elapsed := time.Since(startTime)

	// Output ONLY transcript to stdout (clean for piping)
	fmt.Fprint(os.Stdout, result.Transcript)

	// Print diagnostic information to stderr
	fmt.Fprintf(os.Stderr, "\n--- Summary ---\n")
	fmt.Fprintf(os.Stderr, "Duration: %s\n", elapsed.Round(time.Second))
	fmt.Fprintf(os.Stderr, "Segments: %d\n", len(result.Segments))
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: go run main.go <url_or_file_path>\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  # Transcribe a YouTube video\n")
	fmt.Fprintf(os.Stderr, "  go run main.go https://www.youtube.com/watch?v=dQw4w9WgXcQ\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  # Transcribe an Instagram reel\n")
	fmt.Fprintf(os.Stderr, "  go run main.go https://www.instagram.com/reel/ABC123/\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  # Transcribe a local file\n")
	fmt.Fprintf(os.Stderr, "  go run main.go /path/to/video.mp4\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Or use the Makefile:\n")
	fmt.Fprintf(os.Stderr, "  make transcribe URL=\"https://youtube.com/watch?v=...\"\n")
}
