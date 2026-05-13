# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## OmniTranscripts - Universal Media Transcription Engine

> A Go-based transcription engine and API for audio and video from any URL. Supports 1000+ platforms via yt-dlp, with audio-first workflows as first-class citizens. Available as a Go library or HTTP API.

## Documentation Structure

This project has comprehensive documentation organized in the `docs/` directory:

| Document | Purpose | When to Reference |
|----------|---------|-------------------|
| [API Documentation](docs/api.md) | Complete API reference, endpoints, examples | API changes, endpoint questions, integration help |
| [ChatGPT Integration](docs/chatgpt-integration.md) | MCP server setup for ChatGPT | ChatGPT/MCP integration, OpenAI Apps SDK |
| [Architecture](docs/architecture.md) | Technical architecture, design patterns, scaling | Understanding codebase structure, performance questions |
| [Deployment](docs/deployment.md) | Production deployment guides | Deployment issues, configuration questions |
| [Development](docs/development.md) | Development setup, workflows, testing | Setting up environment, development practices |
| [Troubleshooting](docs/troubleshooting.md) | Common issues and solutions | Error resolution, debugging help |
| [Contributing](docs/contributing.md) | Contribution guidelines and standards | Code style, PR process, community guidelines |
| [Changelog](docs/changelog.md) | Version history and changes | Understanding changes, version differences |
| [ADRs](docs/decisions/) | Architecture Decision Records | Understanding design decisions |

**Always reference relevant documentation when helping with specific topics.**

## Development Commands

All development operations use the comprehensive Makefile:

```bash
# Essential commands
make help          # Show all available commands with descriptions
make dev           # Run server in development mode
make test-short    # Run quick tests (skip load tests)
make build         # Build application
make clean         # Clean all build artifacts

# CLI media operations (no server needed)
make download URL="https://music.youtube.com/watch?v=..."     # Download media to ~/Downloads
make download URL="https://..." OUT=./video.mp4               # Download to custom path
make transcribe URL="https://www.youtube.com/watch?v=..."     # Transcribe any URL
make transcribe URL="https://www.instagram.com/reel/ABC123/" # Instagram, TikTok, etc.
make transcribe URL="/path/to/video.mp4"                     # Local file

# Testing & quality
make test          # Run all tests including load tests
make test-coverage # Generate coverage report in coverage/coverage.html
make benchmark     # Run performance benchmarks
make perf-short    # Quick performance tests
make fmt           # Format code (go fmt + goimports if available)
make lint          # Run golangci-lint (requires setup)
make check         # Run fmt + lint + vet + test-short

# Single test execution
go test -run TestSpecificFunction ./package
go test -v -run TestPostTranscribe_ValidationErrors ./handlers
```

**Complete command reference and development workflows:** [docs/development.md](docs/development.md)

## CLI Output and Piping

The `make transcribe` command outputs the transcript text to stdout and progress/diagnostics to stderr. This enables natural Unix piping for flexible workflows:

### Copy transcript to clipboard
```bash
# macOS
make transcribe URL="https://www.youtube.com/watch?v=dQw4w9WgXcQ" | pbcopy

# Linux (xclip)
make transcribe URL="https://www.youtube.com/watch?v=dQw4w9WgXcQ" | xclip -selection clipboard

# Linux (xsel)
make transcribe URL="https://www.youtube.com/watch?v=dQw4w9WgXcQ" | xsel --clipboard --input
```

### Save transcript to file
```bash
make transcribe URL="https://www.youtube.com/watch?v=dQw4w9WgXcQ" > transcript.txt
make transcribe URL="https://www.instagram.com/reel/ABC123/" >> all-transcripts.txt
```

### Suppress progress output
```bash
make transcribe URL="https://www.youtube.com/watch?v=dQw4w9WgXcQ" 2>/dev/null | pbcopy
```

### Pipe to other tools
```bash
# Word count
make transcribe URL="..." 2>/dev/null | wc -w

# Search for keywords
make transcribe URL="..." | grep -i "keyword"

# Get first 20 lines
make transcribe URL="..." | head -20

# Line count
make transcribe URL="..." 2>/dev/null | wc -l
```

**Note:** Progress and diagnostic information appears on stderr, allowing you to see transcription status while safely piping stdout to other commands.

Environment configuration via `.env` file:
```bash
PORT=3000
API_KEY=dev-api-key-12345
WORK_DIR=/tmp/omnitranscripts
MAX_VIDEO_LENGTH=1800
FREE_JOB_LIMIT=5
# MCP Server (ChatGPT integration)
MCP_ENABLED=true
MCP_ENDPOINT=/mcp
```

## Architecture Overview

### Dual Consumption Model
OmniTranscripts supports two consumption modes from one codebase:

1. **Go Library** (`engine/` package): Direct programmatic use via import
2. **HTTP API**: Thin adapter over the engine for external clients

The transcription engine is transport-agnostic. HTTP concerns live at the edge only.

### Core Processing Pipeline
The application implements a three-stage media transcription pipeline:

1. **Download Stage** (`engine/engine.go`): Uses yt-dlp to extract audio from any supported URL (1000+ platforms)
2. **Normalize Stage**: Uses ffmpeg-go to convert audio to 16kHz mono WAV format suitable for whisper.cpp
3. **Transcription Stage**: Transcribes using whisper.cpp (native, server, or demo fallback)

### Structured Errors
The `engine.TranscriptionError` type provides stage-specific error reporting:
- `StageDownload`: Download failures (network, unsupported URL, etc.)
- `StageNormalize`: Audio conversion failures
- `StageTranscribe`: Transcription failures

### Job Processing System
**Sync vs Async Decision**: Media <=2 minutes processes synchronously with real-time response polling. Longer media uses async job queue with job ID for status checking.

**Job States** (`jobs/job.go`): `pending` → `running` → `complete`/`error`

**Job Queue** (`jobs/queue.go`): Thread-safe in-memory map with read/write mutex. Single global instance initialized in main.go.

**Processing Flow** (`handlers/transcribe.go`):
- `PostTranscribe`: Creates job → determines sync/async based on duration → launches goroutine
- `GetTranscribeJob`: Returns job status and results for async jobs
- Background processing calls `lib.ProcessTranscription` which orchestrates the pipeline

**Detailed architecture documentation:** [docs/architecture.md](docs/architecture.md)

### Package Organization
- `main.go`: Fiber server setup, middleware, routing, MCP server mounting
- `engine/`: **Public library package** - core transcription types and functions
- `config/`: Environment variable management with defaults
- `handlers/`: HTTP request handlers and response logic
- `jobs/`: Job data structures and thread-safe queue implementation
- `lib/`: Internal utilities (auth middleware, subtitles, webhooks)
- `mcp/`: MCP (Model Context Protocol) server for ChatGPT integration
- `models/`: Request/response structs and URL validation
- `scripts/`: Performance testing automation

### External Dependencies
**Runtime Requirements**: The application uses external binaries:
- `yt-dlp`: For media downloads (any supported URL)
- `ffmpeg`: For audio processing
- `whisper.cpp`: For transcription (optional, has demo fallback)

**Go Libraries**:
- `github.com/gofiber/fiber/v2`: HTTP framework (Express-like for Go)
- `github.com/lrstanley/go-ytdlp`: Go wrapper for yt-dlp
- `github.com/u2takey/ffmpeg-go`: Go wrapper for FFmpeg
- `github.com/google/uuid`: Job ID generation
- `github.com/mark3labs/mcp-go`: MCP protocol implementation for ChatGPT integration

## Key Implementation Details

### Authentication
Bearer token authentication via `lib.AuthMiddleware()` checks `Authorization: Bearer <token>` against `API_KEY` environment variable. Health endpoint bypasses authentication.

### File Management
Temporary files created in `WORK_DIR` with job ID naming. Automatic cleanup via defer statements in processing functions.

### Error Handling
Structured errors via `engine.TranscriptionError` provide stage-specific context. Errors propagate up through the pipeline with wrapping. Job errors stored in Job.Error field and returned via API.

### Testing Architecture
- Unit tests in `handlers/transcribe_test.go` with Fiber test framework
- Performance benchmarks in `lib/transcription_bench_test.go`
- Load testing with configurable concurrency and performance assertions
- Performance test runner script at `scripts/run_perf_tests.sh`

### Configuration Patterns
Environment variables loaded once at startup via `config.Load()`. No runtime config changes supported.

## API Endpoints

### HTTP API
- `GET /health`: Unauthenticated health check
- `POST /transcribe`: Submit media URL, returns transcript (sync) or job ID (async)
- `GET /transcribe/{job_id}`: Get job status and results

Request/response handled via `models.TranscribeRequest` and `models.TranscribeResponse` structs.

### MCP Server (ChatGPT Integration)
- `POST/GET/DELETE /mcp`: MCP protocol endpoint for ChatGPT Apps SDK
- Tools: `transcribe_url` (start transcription), `get_transcription` (check status/retrieve results)

**Complete API reference with examples:** [docs/api.md](docs/api.md)
**ChatGPT integration guide:** [docs/chatgpt-integration.md](docs/chatgpt-integration.md)

## Development Notes

- Use `make dev` for development server with automatic restarts
- Tests require external dependencies (yt-dlp, ffmpeg, whisper.cpp) to pass fully
- Load tests have strict performance requirements that may fail without external deps
- The `parseDuration` function currently returns hardcoded 120 seconds (TODO: implement actual parsing)
- Job queue is in-memory only - jobs lost on restart (suitable for MVP)

**Complete development setup guide:** [docs/development.md](docs/development.md)

## Troubleshooting Reference

For common issues and solutions:
- **External tool dependencies**: whisper.cpp, yt-dlp, ffmpeg installation issues
- **Processing failures**: Job stuck in pending, download/transcription errors
- **Performance issues**: Memory usage, slow processing, disk space
- **Deployment problems**: Docker, cloud platform, networking issues

**Complete troubleshooting guide:** [docs/troubleshooting.md](docs/troubleshooting.md)

## Production Deployment

For production deployments, refer to comprehensive guides covering:
- **Docker**: Container builds and orchestration
- **Cloud Platforms**: AWS, GCP, Azure deployment options
- **Encore.dev**: Zero-config production deployment
- **Traditional Servers**: SystemD, process management, reverse proxy setup

**Complete deployment guides:** [docs/deployment.md](docs/deployment.md)
