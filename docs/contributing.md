# Contributing to OmniTranscripts

Thank you for your interest in contributing to OmniTranscripts! This guide will help you get started with contributing to this YouTube video transcription API.

## Quick Start for Contributors

### 1. Fork and Clone
```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/OmniTranscripts.git
cd OmniTranscripts

# Add upstream remote
git remote add upstream https://github.com/wilmoore/OmniTranscripts.git
```

### 2. Set Up Development Environment
```bash
# Install dependencies (see docs/development.md for details)
make setup

# Copy environment configuration
cp .env.example .env

# Install Go dependencies
go mod tidy

# Run tests to ensure everything works
make test-short
```

### 3. Create a Feature Branch
```bash
# Sync with upstream
git fetch upstream
git checkout main
git merge upstream/main

# Create feature branch
git checkout -b feature/your-feature-name
```

## Types of Contributions

### 🐛 Bug Reports
Help us identify and fix issues:

1. **Search existing issues** first to avoid duplicates
2. **Use the bug report template** when creating new issues
3. **Include detailed information**:
   - Go version and OS
   - Steps to reproduce
   - Expected vs actual behavior
   - Relevant log outputs
   - Minimal reproduction case

**Example Bug Report**:
```markdown
## Bug Description
Transcription fails for videos longer than 10 minutes with "context deadline exceeded" error.

## Steps to Reproduce
1. Submit video URL: https://www.youtube.com/watch?v=LONG_VIDEO_ID
2. Wait for processing
3. Check job status

## Expected Behavior
Video should be transcribed successfully

## Actual Behavior
Job fails with timeout error after 5 minutes

## Environment
- Go version: 1.23.0
- OS: Ubuntu 22.04
- OmniTranscripts version: main branch
```

### ✨ Feature Requests
Suggest new functionality:

1. **Check the roadmap** and existing feature requests
2. **Describe the use case** and problem it solves
3. **Provide implementation suggestions** if you have ideas
4. **Consider backwards compatibility**

**Example Feature Request**:
```markdown
## Feature Description
Add support for batch processing multiple videos in a single API call

## Use Case
Users often need to transcribe multiple related videos (e.g., a video series) and would benefit from submitting them all at once.

## Proposed API
```json
{
  "urls": [
    "https://www.youtube.com/watch?v=video1",
    "https://www.youtube.com/watch?v=video2"
  ],
  "batch_options": {
    "priority": "normal",
    "webhook_url": "https://example.com/webhook"
  }
}
```

## Implementation Ideas
- Add new `/transcribe/batch` endpoint
- Return batch job ID for tracking all videos
- Support batch-level webhooks
```

### 🔧 Code Contributions

#### Areas Needing Help
- **Performance optimizations**: Improve transcription speed
- **Error handling**: Better error messages and recovery
- **Testing**: Increase test coverage
- **Documentation**: API docs, guides, examples
- **Integrations**: SDK libraries for different languages
- **Monitoring**: Metrics and observability features

#### Before You Start
1. **Discuss large changes** in an issue first
2. **Follow the coding standards** (see below)
3. **Write tests** for new functionality
4. **Update documentation** as needed

## Coding Standards

### Go Code Style

#### General Guidelines
```go
// Good: Clear, descriptive names
func processTranscriptionJob(job *models.Job) error {
    return nil
}

// Bad: Unclear abbreviations
func procTxJob(j *models.Job) error {
    return nil
}
```

#### Error Handling
```go
// Good: Descriptive error with context
if err := validateURL(req.URL); err != nil {
    return fmt.Errorf("invalid YouTube URL %q: %w", req.URL, err)
}

// Good: Structured error response
return &errs.Error{
    Code:    errs.InvalidArgument,
    Message: "URL validation failed",
    Details: map[string]interface{}{
        "url": req.URL,
        "reason": "not a valid YouTube URL",
    },
}
```

#### Function Documentation
```go
// ProcessTranscription orchestrates the complete video transcription pipeline.
// It downloads the video, extracts audio, and generates timestamped transcripts.
//
// The function handles both short videos (processed synchronously) and long videos
// (processed asynchronously via job queue).
//
// Parameters:
//   - url: YouTube video URL to transcribe
//   - jobID: Unique identifier for tracking processing progress
//
// Returns:
//   - transcript: Complete text transcription
//   - segments: Timestamped transcript segments
//   - error: Processing error or nil on success
func ProcessTranscription(url, jobID string) (string, []models.Segment, error) {
    // Implementation...
}
```

#### Testing Standards
```go
func TestProcessTranscription(t *testing.T) {
    tests := []struct {
        name        string
        url         string
        jobID       string
        wantErr     bool
        wantSegments int
    }{
        {
            name:         "valid short video",
            url:          "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
            jobID:        "test_job_123",
            wantErr:      false,
            wantSegments: 10, // Approximate expected segments
        },
        {
            name:    "invalid URL",
            url:     "not-a-url",
            jobID:   "test_job_456",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            transcript, segments, err := ProcessTranscription(tt.url, tt.jobID)

            if (err != nil) != tt.wantErr {
                t.Errorf("ProcessTranscription() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !tt.wantErr {
                assert.NotEmpty(t, transcript)
                assert.Len(t, segments, tt.wantSegments)
            }
        })
    }
}
```

### API Design Guidelines

#### Request/Response Models
```go
// Good: Clear field names and validation tags
type TranscribeRequest struct {
    URL     string            `json:"url" validate:"required,url"`
    Options *TranscribeOptions `json:"options,omitempty"`
}

type TranscribeOptions struct {
    Model    string `json:"model,omitempty" validate:"oneof=tiny small base large"`
    Language string `json:"language,omitempty" validate:"len=2"`
}

// Good: Consistent response structure
type TranscribeResponse struct {
    JobID      string           `json:"job_id,omitempty"`
    Transcript string           `json:"transcript,omitempty"`
    Segments   []models.Segment `json:"segments,omitempty"`
    Status     string           `json:"status"`
    CreatedAt  time.Time        `json:"created_at"`
}
```

#### Error Responses
```go
// Consistent error format
type ErrorResponse struct {
    Error   string                 `json:"error"`
    Code    string                 `json:"code,omitempty"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

## Development Workflow

### 1. Making Changes
```bash
# Make your changes
vim lib/transcription.go

# Run tests frequently
make test-short

# Check code quality
make lint
make fmt
```

### 2. Testing Your Changes
```bash
# Run all tests
make test

# Run specific tests
go test -v ./lib/ -run TestProcessTranscription

# Test with coverage
make test-coverage

# Performance tests (if relevant)
make benchmark
```

### 3. Documentation Updates
- Update relevant `.md` files in `docs/`
- Add inline code comments for new functions
- Update API documentation if endpoints change
- Add examples for new features

### 4. Commit Guidelines

#### Commit Message Format
```
<type>(<scope>): <description>

<body>

<footer>
```

#### Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

#### Examples
```bash
# Good commit messages
git commit -m "feat(api): add batch transcription endpoint

Support processing multiple videos in a single request.
Includes new /transcribe/batch endpoint with job tracking.

Closes #123"

git commit -m "fix(transcription): handle audio extraction timeout

Increase FFmpeg timeout for long videos and add retry logic.
Fixes issue where 20+ minute videos would fail consistently.

Fixes #456"

git commit -m "docs(api): update endpoint documentation

Add examples for new batch processing endpoint and
clarify error response formats."
```

### 5. Pull Request Process

#### Before Submitting
```bash
# Sync with upstream
git fetch upstream
git rebase upstream/main

# Final check
make check  # Runs fmt, lint, vet, and test-short

# Push to your fork
git push origin feature/your-feature-name
```

#### Pull Request Template
When creating a PR, include:

1. **Description**: Clear explanation of changes
2. **Motivation**: Why is this change needed?
3. **Testing**: How was this tested?
4. **Breaking Changes**: Any backwards compatibility issues?
5. **Checklist**: Confirm you've completed all requirements

**Example PR Description**:
```markdown
## Description
Adds support for batch processing multiple YouTube videos in a single API request.

## Motivation
Users frequently need to transcribe multiple related videos and currently have to make separate API calls for each one. This feature allows them to submit multiple URLs at once and track progress via a single batch job ID.

## Changes
- Added `/transcribe/batch` endpoint
- New `BatchTranscribeRequest` and `BatchTranscribeResponse` models
- Batch job tracking and status aggregation
- Updated documentation and examples

## Testing
- Unit tests for batch processing logic
- Integration tests with real YouTube videos
- Load testing with 10+ videos per batch
- Manual testing via curl and Postman

## Breaking Changes
None - this is a new endpoint with no changes to existing APIs.

## Checklist
- [x] Tests pass locally
- [x] Code follows style guidelines
- [x] Documentation updated
- [x] No breaking changes
- [x] Performance impact assessed
```

## Review Process

### What to Expect
1. **Automated checks**: CI will run tests and linting
2. **Code review**: Maintainers will review your code
3. **Feedback**: You may need to make changes
4. **Approval**: Once approved, your PR will be merged

### Responding to Feedback
- **Be responsive**: Reply to comments promptly
- **Ask questions**: If feedback isn't clear, ask for clarification
- **Make requested changes**: Address all feedback before requesting re-review
- **Update tests**: If functionality changes, update tests accordingly

## Community Guidelines

### Code of Conduct
We follow the [Contributor Covenant](https://www.contributor-covenant.org/version/2/1/code_of_conduct/) to ensure a welcoming environment for all contributors.

### Communication
- **Be respectful**: Treat all community members with respect
- **Be constructive**: Provide helpful feedback and suggestions
- **Be patient**: Maintainers are volunteers with limited time
- **Stay on topic**: Keep discussions focused on the project

### Getting Help
- **Documentation**: Check the [docs](../docs/) first
- **GitHub Discussions**: Ask questions and share ideas
- **Issues**: Report bugs and request features
- **Discord** (if available): Real-time chat with other contributors

## Recognition

### Contributors
All contributors are recognized in:
- Repository contributors page
- Release notes for significant contributions
- Special mentions for major features or bug fixes

### Becoming a Maintainer
Regular contributors who demonstrate:
- Consistent high-quality contributions
- Good understanding of the codebase
- Helpful community participation
- Commitment to project goals

May be invited to become maintainers with additional responsibilities.

## Resources

### Documentation
- [Development Guide](development.md) - Development setup and workflows
- [Architecture](architecture.md) - Technical architecture details
- [API Documentation](api.md) - Complete API reference
- [Deployment Guide](deployment.md) - Production deployment options
- [Troubleshooting](troubleshooting.md) - Common issues and solutions

### Tools and Links
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [GitHub Flow](https://guides.github.com/introduction/flow/)
- [Semantic Versioning](https://semver.org/)

Thank you for contributing to OmniTranscripts! Your contributions help make video transcription more accessible and powerful for everyone.