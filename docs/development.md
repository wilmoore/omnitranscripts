# Development Guide

This guide covers everything you need to know for developing and contributing to OmniTranscripts.

## Quick Start

### Prerequisites

1. **Go 1.23+** - [Install Go](https://golang.org/doc/install)
2. **FFmpeg** - Audio processing
3. **OpenAI Whisper** - AI transcription engine (Python package)

### Installation

#### macOS (using Homebrew)
```bash
# Install dependencies
brew install go ffmpeg
pip install openai-whisper
```

#### Ubuntu/Debian
```bash
# Install Go (if not available via package manager)
wget https://go.dev/dl/go1.23.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install other dependencies
sudo apt update
sudo apt install -y ffmpeg python3-pip
pip install openai-whisper
```

### Project Setup

```bash
# Clone the repository
git clone https://github.com/wilmoore/OmniTranscripts.git
cd OmniTranscripts

# Install Go dependencies
go mod tidy

# Copy environment configuration
cp .env.example .env
vim .env  # Edit configuration

# Run the development server
make dev
```

## Development Workflow

### Using the Makefile

The project includes a comprehensive Makefile with organized commands:

```bash
# Show all available commands
make help

# Development workflow
make setup         # Install development dependencies
make dev           # Run with hot reload
make test          # Run all tests
make lint          # Run code quality checks
make build         # Build for current platform

# Quality assurance
make check         # Run all quality checks (fmt + lint + vet + test)
make test-coverage # Generate coverage report
make benchmark     # Run performance benchmarks
```

### Development Server Options

#### Option 1: Standard Go Server (Fiber)
```bash
# Using Makefile (recommended)
make dev

# Or manually
go run main.go
```

#### Option 2: Encore.dev Framework
```bash
# Install Encore CLI
curl -L https://encore.dev/install.sh | bash

# Run Encore development server
encore run

# The server will be available at http://localhost:4000
```

#### Option 3: Web Dashboard (Development UI)
```bash
# Run the development dashboard
go run web-dashboard.go

# Dashboard available at http://localhost:8765+
```

### Environment Configuration

#### Development `.env` file
```env
# Server Configuration
PORT=3000

# Authentication
API_KEY=dev-api-key-12345

# Processing Configuration
WORK_DIR=/tmp/videotranscript
MAX_VIDEO_LENGTH=1800
FREE_JOB_LIMIT=5

# Development Features
DEBUG=1
LOG_LEVEL=debug

# Database (for Encore.dev)
# DATABASE_URL=postgresql://user:pass@localhost:5432/videotranscript_dev

# External Tool Paths (if not in PATH)
# FFMPEG_PATH=/usr/local/bin/ffmpeg
```

## Code Structure

### Project Layout
```
OmniTranscripts/
├── main.go                 # Fiber server entry point
├── web-dashboard.go        # Development dashboard
├── config/                 # Configuration management
├── handlers/               # HTTP request handlers
├── jobs/                   # Job management system
├── lib/                    # Core business logic
├── models/                 # Data structures
├── transcribe/             # Encore.dev service
│   ├── service.go         # Main service file
│   ├── db.go             # Database operations
│   └── migrations/       # SQL migrations
├── scripts/               # Development scripts
├── docs/                  # Documentation
└── tests/                 # Test files
```

### Package Dependencies

#### Core Dependencies
```go
// HTTP Framework
github.com/gofiber/fiber/v2

// Native Go Libraries
github.com/lrstanley/go-ytdlp   // YouTube downloading (native Go)
github.com/u2takey/ffmpeg-go   // FFmpeg wrapper for audio processing

// Utilities
github.com/google/uuid          // Job ID generation
github.com/joho/godotenv       // Environment loading

// Encore.dev Framework
encore.dev/beta/auth           // Authentication
encore.dev/beta/errs           // Error handling
encore.dev/beta/pubsub         // Pub/Sub messaging
encore.dev/config              // Configuration
encore.dev/rlog                // Logging
```

#### Development Dependencies
```go
// Testing
github.com/stretchr/testify

// Code Quality
github.com/golangci/golangci-lint
golang.org/x/tools/cmd/goimports
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run quick tests (skip load tests)
make test-short

# Run with coverage
make test-coverage

# Run specific test files
go test -v ./handlers/
go test -run TestSpecificFunction ./lib/

# Run benchmarks
make benchmark
```

### Test Structure

#### Unit Tests
```go
// handlers/transcribe_test.go
func TestPostTranscribe_ValidationErrors(t *testing.T) {
    tests := []struct {
        name           string
        requestBody    string
        expectedStatus int
        expectedError  string
    }{
        {
            name:           "Invalid JSON",
            requestBody:    `{"invalid": json}`,
            expectedStatus: 400,
            expectedError:  "Invalid request format",
        },
        {
            name:           "Missing URL",
            requestBody:    `{}`,
            expectedStatus: 400,
            expectedError:  "URL is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### Integration Tests
```go
// lib/transcription_test.go
func TestTranscriptionPipeline(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Test complete pipeline with real external tools
    url := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
    transcript, segments, err := ProcessTranscription(url, "test_job")

    assert.NoError(t, err)
    assert.NotEmpty(t, transcript)
    assert.NotEmpty(t, segments)
}
```

#### Performance Tests
```bash
# Run comprehensive performance tests
./scripts/run_perf_tests.sh

# Or using Make
make perf
```

### Test Data

#### Mock Responses
```go
// tests/mocks/youtube_mock.go
type MockYouTubeResponse struct {
    Title    string `json:"title"`
    Duration int    `json:"duration"`
    URL      string `json:"url"`
}

var MockShortVideo = MockYouTubeResponse{
    Title:    "Test Short Video",
    Duration: 30,
    URL:      "https://www.youtube.com/watch?v=test123",
}
```

## Development Tools

### Code Quality

#### Linting Configuration
```yaml
# .golangci.yml
linters-settings:
  gocyclo:
    min-complexity: 15
  goconst:
    min-len: 2
    min-occurrences: 2
  misspell:
    locale: US

linters:
  enable:
    - gocyclo
    - goconst
    - gofmt
    - goimports
    - golint
    - gosec
    - ineffassign
    - misspell
    - vet

run:
  timeout: 5m
```

#### Code Formatting
```bash
# Format code
make fmt

# Or manually
go fmt ./...
goimports -w .
```

### Debugging

#### Debug Mode
```bash
# Enable debug logging
DEBUG=1 go run main.go

# Or set in .env file
echo "DEBUG=1" >> .env
```

#### Profiling
```go
// Add to main.go for CPU profiling
import _ "net/http/pprof"

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    // Rest of application...
}
```

```bash
# Profile the application
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap
```

### Database Development

#### Encore.dev Migrations
```bash
# Create new migration
encore db migrate create --name add_new_feature

# Apply migrations
encore db migrate

# Reset database (development only)
encore db reset
```

#### Manual Database Setup
```sql
-- Create development database
CREATE DATABASE videotranscript_dev;

-- Run migrations manually
psql -d videotranscript_dev -f transcribe/migrations/1_create_jobs.up.sql
psql -d videotranscript_dev -f transcribe/migrations/2_expand_jobs_table.up.sql
psql -d videotranscript_dev -f transcribe/migrations/3_create_metrics_tables.up.sql
```

## API Development

### Adding New Endpoints

#### Fiber Server Implementation
```go
// handlers/new_feature.go
func NewFeatureHandler(c *fiber.Ctx) error {
    // Parse request
    var req NewFeatureRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request format",
        })
    }

    // Validate request
    if err := validateNewFeatureRequest(req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    // Process request
    result, err := processNewFeature(req)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Processing failed",
        })
    }

    return c.JSON(result)
}

// Register in main.go
app.Post("/new-feature", middleware.Auth(), handlers.NewFeatureHandler)
```

#### Encore.dev Implementation
```go
// transcribe/service.go

// NewFeatureRequest represents the request for new feature
type NewFeatureRequest struct {
    Parameter string `json:"parameter"`
}

// NewFeatureResponse represents the response
type NewFeatureResponse struct {
    Result string `json:"result"`
}

// NewFeature implements the new feature endpoint
//
//encore:api auth method=POST path=/new-feature
func NewFeature(ctx context.Context, req *NewFeatureRequest) (*NewFeatureResponse, error) {
    // Validate request
    if req.Parameter == "" {
        return nil, &errs.Error{
            Code:    errs.InvalidArgument,
            Message: "Parameter is required",
        }
    }

    // Process request
    result, err := processNewFeature(req.Parameter)
    if err != nil {
        return nil, &errs.Error{
            Code:    errs.Internal,
            Message: "Processing failed",
        }
    }

    return &NewFeatureResponse{
        Result: result,
    }, nil
}
```

### Request/Response Models

```go
// models/new_feature.go
type NewFeatureRequest struct {
    Parameter string `json:"parameter" validate:"required,min=1,max=100"`
}

type NewFeatureResponse struct {
    Result    string    `json:"result"`
    ProcessedAt time.Time `json:"processed_at"`
}

func (r *NewFeatureRequest) Validate() error {
    if r.Parameter == "" {
        return errors.New("parameter is required")
    }
    return nil
}
```

## Performance Development

### Profiling and Optimization

#### CPU Profiling
```go
import (
    "os"
    "runtime/pprof"
)

func enableCPUProfiling() {
    f, err := os.Create("cpu.prof")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    if err := pprof.StartCPUProfile(f); err != nil {
        log.Fatal(err)
    }
    defer pprof.StopCPUProfile()
}
```

#### Memory Optimization
```go
// Use object pooling for frequently allocated objects
var segmentPool = sync.Pool{
    New: func() interface{} {
        return make([]models.Segment, 0, 100)
    },
}

func processSegments(data []byte) []models.Segment {
    segments := segmentPool.Get().([]models.Segment)
    defer segmentPool.Put(segments[:0])

    // Process segments...
    return append([]models.Segment(nil), segments...)
}
```

### Load Testing

#### Custom Load Test Script
```bash
#!/bin/bash
# scripts/load_test.sh

API_URL="http://localhost:3000"
API_KEY="dev-api-key-12345"
CONCURRENT_USERS=10
TEST_DURATION=60

# Generate test URLs
TEST_URLS=(
    "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
    "https://www.youtube.com/watch?v=9bZkp7q19f0"
    # Add more test URLs
)

# Run concurrent requests
for i in $(seq 1 $CONCURRENT_USERS); do
    {
        for url in "${TEST_URLS[@]}"; do
            curl -X POST "$API_URL/transcribe" \
                -H "Authorization: Bearer $API_KEY" \
                -H "Content-Type: application/json" \
                -d "{\"url\": \"$url\"}" \
                --silent --output /dev/null
        done
    } &
done

wait
echo "Load test completed"
```

## Contributing Guidelines

### Code Style

#### Go Conventions
- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions small and focused
- Use Go modules for dependency management

#### Error Handling
```go
// Good: Specific error messages
if err := validateURL(url); err != nil {
    return fmt.Errorf("URL validation failed: %w", err)
}

// Good: Context in errors
if err := downloadVideo(url); err != nil {
    return fmt.Errorf("failed to download video %s: %w", url, err)
}
```

#### Testing Best Practices
```go
// Use table-driven tests
func TestValidateURL(t *testing.T) {
    tests := []struct {
        name    string
        url     string
        wantErr bool
    }{
        {"valid youtube url", "https://www.youtube.com/watch?v=abc123", false},
        {"invalid url", "not-a-url", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateURL(tt.url)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateURL() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Git Workflow

#### Branch Naming
- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation updates
- `refactor/description` - Code refactoring

#### Commit Messages
```bash
# Good commit messages
git commit -m "feat: add webhook support for job completion notifications"
git commit -m "fix: handle edge case in audio duration calculation"
git commit -m "docs: update API documentation with new endpoints"
git commit -m "test: add integration tests for transcription pipeline"
```

### Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the full test suite: `make check`
6. Update documentation if needed
7. Submit a pull request with a clear description

### Release Process

#### Version Tagging
```bash
# Create a new release
git tag -a v1.2.0 -m "Release version 1.2.0"
git push origin v1.2.0
```

#### Changelog Updates
Update `docs/changelog.md` with:
- New features
- Bug fixes
- Breaking changes
- Performance improvements

This development guide provides everything needed to contribute effectively to OmniTranscripts, from initial setup to advanced development workflows.