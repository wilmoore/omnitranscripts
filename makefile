# OmniTranscripts Makefile
# Universal Media Transcription Engine

# Variables
BINARY_NAME=omnitranscripts
GO_VERSION=1.23
MAIN_PACKAGE=.
BUILD_DIR=build
DIST_DIR=dist
COVERAGE_DIR=coverage
DOCKER_IMAGE=omnitranscripts
DOCKER_TAG=latest

# Get git info for versioning
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME) -X main.GitBranch=$(GIT_BRANCH)"

# CGO is disabled by default (whisper.cpp native requires manual setup)
CGO_ENABLED ?= 0

# Colors for output
BLUE=\033[0;34m
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: help
help: ## Show this help message
	@echo "$(BLUE)OmniTranscripts - Available Commands$(NC)"
	@echo "======================================"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make $(GREEN)<target>$(NC)\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(BLUE)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
.PHONY: setup
setup: ## Install development dependencies
	@echo "$(BLUE)Setting up development environment...$(NC)"
	@go mod download
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest 2>/dev/null || true
	@mkdir -p $(BUILD_DIR) $(DIST_DIR) $(COVERAGE_DIR)
	@echo "$(GREEN)Development environment ready!$(NC)"

.PHONY: dev
dev: ## Run the application in development mode
	@echo "$(BLUE)Starting development server...$(NC)"
	@CGO_ENABLED=$(CGO_ENABLED) go run .

.PHONY: dev-watch
dev-watch: ## Run with file watching (requires air: go install github.com/air-verse/air@latest)
	@echo "$(BLUE)Starting development server with hot reload...$(NC)"
	@air || (echo "$(YELLOW)Install air for hot reload: go install github.com/air-verse/air@latest$(NC)" && CGO_ENABLED=$(CGO_ENABLED) go run .)

.PHONY: run
run: ## Run the application
	@echo "$(BLUE)Running application...$(NC)"
	@CGO_ENABLED=$(CGO_ENABLED) go run .

##@ CLI
.PHONY: download
download: ## Download media from URL (make download URL="..." [OUT=path])
ifndef URL
	@echo "$(RED)Error: URL is required$(NC)"
	@echo "Usage: make download URL=\"https://...\""
	@echo ""
	@echo "Options:"
	@echo "  OUT=path    Override output location (default: ~/Downloads/)"
	@echo ""
	@echo "Examples:"
	@echo "  make download URL=\"https://music.youtube.com/watch?v=...\""
	@echo "  make download URL=\"https://www.youtube.com/watch?v=...\" OUT=./video.mp4"
	@exit 1
endif
	@echo "$(BLUE)Downloading: $(URL)$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build -o $(BUILD_DIR)/download examples/download/main.go
	@$(BUILD_DIR)/download "$(URL)" "$(OUT)"

.PHONY: transcribe
transcribe: ## Transcribe a URL or file (make transcribe URL="https://...")
ifndef URL
	@echo "$(RED)Error: URL is required$(NC)" >&2
	@echo "Usage: make transcribe URL=\"https://youtube.com/watch?v=...\"" >&2
	@echo "" >&2
	@echo "Examples:" >&2
	@echo "  make transcribe URL=\"https://www.youtube.com/watch?v=dQw4w9WgXcQ\"" >&2
	@echo "  make transcribe URL=\"https://www.instagram.com/reel/ABC123/\"" >&2
	@echo "  make transcribe URL=\"/path/to/video.mp4\"" >&2
	@exit 1
endif
	@echo "$(BLUE)Transcribing: $(URL)$(NC)" >&2
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=1 go build -o $(BUILD_DIR)/transcribe examples/transcribe/main.go
	@WHISPER_MODEL_PATH=models/ggml-base.en.bin $(BUILD_DIR)/transcribe "$(URL)"

##@ Building
.PHONY: build
build: ## Build the application
	@echo "$(BLUE)Building application...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

.PHONY: build-cgo
build-cgo: ## Build with CGO enabled (for native whisper.cpp)
	@echo "$(BLUE)Building with CGO (native whisper.cpp)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=1 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

.PHONY: build-all
build-all: clean ## Build for all platforms
	@echo "$(BLUE)Building for all platforms...$(NC)"
	@mkdir -p $(DIST_DIR)

	# Linux AMD64
	@echo "Building for Linux AMD64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)

	# Linux ARM64
	@echo "Building for Linux ARM64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)

	# macOS AMD64
	@echo "Building for macOS AMD64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)

	# macOS ARM64
	@echo "Building for macOS ARM64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)

	# Windows AMD64
	@echo "Building for Windows AMD64..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)

	@echo "$(GREEN)Multi-platform build complete! Check $(DIST_DIR)/$(NC)"

##@ Testing
.PHONY: test
test: ## Run all tests
	@echo "$(BLUE)Running tests...$(NC)"
	@CGO_ENABLED=$(CGO_ENABLED) go test ./...

.PHONY: test-short
test-short: ## Run tests in short mode (skip long-running tests)
	@echo "$(BLUE)Running short tests...$(NC)"
	@CGO_ENABLED=$(CGO_ENABLED) go test -short ./...

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	@echo "$(BLUE)Running tests (verbose)...$(NC)"
	@CGO_ENABLED=$(CGO_ENABLED) go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) go test -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	@go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@go tool cover -func=$(COVERAGE_DIR)/coverage.out | grep total
	@echo "$(GREEN)Coverage report: $(COVERAGE_DIR)/coverage.html$(NC)"

.PHONY: benchmark
benchmark: ## Run benchmark tests
	@echo "$(BLUE)Running benchmark tests...$(NC)"
	@CGO_ENABLED=$(CGO_ENABLED) go test -bench=. -benchmem -benchtime=5s ./...

.PHONY: perf
perf: ## Run comprehensive performance tests
	@echo "$(BLUE)Running performance test suite...$(NC)"
	@chmod +x scripts/run_perf_tests.sh
	@./scripts/run_perf_tests.sh

.PHONY: perf-short
perf-short: ## Run quick performance tests
	@echo "$(BLUE)Running quick performance tests...$(NC)"
	@chmod +x scripts/run_perf_tests.sh
	@./scripts/run_perf_tests.sh short

##@ Code Quality
.PHONY: lint
lint: ## Run linter
	@echo "$(BLUE)Running linter...$(NC)"
	@golangci-lint run ./... || echo "$(YELLOW)Install golangci-lint for linting$(NC)"

.PHONY: fmt
fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	@go fmt ./...
	@goimports -w . 2>/dev/null || true

.PHONY: vet
vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	@go vet ./...

.PHONY: check
check: fmt vet test-short ## Run all quality checks
	@echo "$(GREEN)All quality checks passed!$(NC)"

##@ Dependencies
.PHONY: deps
deps: ## Download dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	@go mod download

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "$(BLUE)Updating dependencies...$(NC)"
	@go get -u ./...
	@go mod tidy

.PHONY: deps-verify
deps-verify: ## Verify dependencies
	@echo "$(BLUE)Verifying dependencies...$(NC)"
	@go mod verify

.PHONY: tidy
tidy: ## Tidy go modules
	@echo "$(BLUE)Tidying go modules...$(NC)"
	@go mod tidy

##@ Docker
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(NC)"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(NC)"

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "$(BLUE)Running Docker container...$(NC)"
	@docker run -p 3000:3000 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-push
docker-push: docker-build ## Build and push Docker image
	@echo "$(BLUE)Pushing Docker image...$(NC)"
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-clean
docker-clean: ## Clean Docker images and containers
	@echo "$(BLUE)Cleaning Docker images and containers...$(NC)"
	@docker system prune -f
	@docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true

##@ Transcription
.PHONY: setup-transcription
setup-transcription: ## Set up transcription services (models, dependencies)
	@echo "$(BLUE)Setting up transcription services...$(NC)"
	@make setup-models
	@make setup-env-example
	@echo "$(GREEN)Transcription setup complete!$(NC)"
	@echo "$(YELLOW)Configure your .env file with API keys to enable transcription$(NC)"

.PHONY: setup-models
setup-models: ## Download whisper models for offline transcription
	@echo "$(BLUE)Setting up whisper models...$(NC)"
	@mkdir -p models
	@./scripts/download-models.sh
	@echo "$(GREEN)Whisper models downloaded$(NC)"

.PHONY: setup-env-example
setup-env-example: ## Create example environment file
	@echo "$(BLUE)Creating .env.example...$(NC)"
	@echo "# OmniTranscripts Configuration" > .env.example
	@echo "PORT=3000" >> .env.example
	@echo "API_KEY=your-api-key-here" >> .env.example
	@echo "" >> .env.example
	@echo "# MCP Server (ChatGPT integration)" >> .env.example
	@echo "MCP_ENABLED=true" >> .env.example
	@echo "MCP_ENDPOINT=/mcp" >> .env.example
	@echo "" >> .env.example
	@echo "# Transcription Services (choose one or both for fallback)" >> .env.example
	@echo "# AssemblyAI - Cloud transcription" >> .env.example
	@echo "ASSEMBLYAI_API_KEY=" >> .env.example
	@echo "" >> .env.example
	@echo "# Whisper Server - Local transcription (requires setup)" >> .env.example
	@echo "WHISPER_SERVER_URL=" >> .env.example
	@echo "WHISPER_MODEL_PATH=models/ggml-base.en.bin" >> .env.example
	@echo "" >> .env.example
	@echo "# Other Settings" >> .env.example
	@echo "WORK_DIR=/tmp/omnitranscripts" >> .env.example
	@echo "MAX_VIDEO_LENGTH=1800" >> .env.example
	@echo "FREE_JOB_LIMIT=5" >> .env.example
	@echo "MAX_UPLOAD_SIZE=524288000" >> .env.example
	@echo "$(GREEN).env.example created$(NC)"

.PHONY: transcription-status
transcription-status: ## Check transcription service status
	@echo "$(BLUE)Checking transcription service status...$(NC)"
	@echo "Environment Variables:"
	@echo "  ASSEMBLYAI_API_KEY: $$([ -n "$$ASSEMBLYAI_API_KEY" ] && echo "✓ Set" || echo "✗ Not set")"
	@echo "  WHISPER_SERVER_URL: $$([ -n "$$WHISPER_SERVER_URL" ] && echo "✓ Set" || echo "✗ Not set")"
	@echo "  WHISPER_MODEL_PATH: $$([ -n "$$WHISPER_MODEL_PATH" ] && echo "✓ Set ($$WHISPER_MODEL_PATH)" || echo "✗ Not set")"
	@echo ""
	@echo "Models Directory:"
	@echo "  models/: $$([ -d models ] && echo "✓ Exists" || echo "✗ Missing")"
	@echo "  ggml-base.en.bin: $$([ -f models/ggml-base.en.bin ] && echo "✓ Downloaded" || echo "✗ Missing")"

.PHONY: clean-models
clean-models: ## Remove downloaded models
	@echo "$(BLUE)Cleaning transcription models...$(NC)"
	@rm -rf models/
	@echo "$(GREEN)Models cleaned$(NC)"

##@ Process Management
.PHONY: ps
ps: ## Show orphaned yt-dlp/ffmpeg processes
	@echo "$(BLUE)Checking for orphaned processes...$(NC)"
	@echo ""
	@echo "yt-dlp processes: $$(pgrep -f yt-dlp 2>/dev/null | wc -l | tr -d ' ')"
	@echo "ffmpeg processes: $$(pgrep -f ffmpeg 2>/dev/null | wc -l | tr -d ' ')"
	@echo ""
	@pgrep -fl yt-dlp 2>/dev/null || echo "$(GREEN)No yt-dlp processes$(NC)"
	@pgrep -fl ffmpeg 2>/dev/null | grep -v "pgrep" || echo "$(GREEN)No ffmpeg processes$(NC)"

.PHONY: ps-kill
ps-kill: ## Kill all orphaned yt-dlp/ffmpeg processes
	@echo "$(YELLOW)Killing orphaned processes...$(NC)"
	@pkill -f yt-dlp 2>/dev/null && echo "$(GREEN)Killed yt-dlp processes$(NC)" || echo "No yt-dlp processes to kill"
	@pkill -f ffmpeg 2>/dev/null && echo "$(GREEN)Killed ffmpeg processes$(NC)" || echo "No ffmpeg processes to kill"
	@echo "$(GREEN)Cleanup complete$(NC)"

##@ Utilities
.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR) $(DIST_DIR) $(COVERAGE_DIR) test_results/
	@rm -f $(BINARY_NAME) coverage.out *.prof
	@go clean -cache -testcache
	@echo "$(GREEN)Clean complete!$(NC)"

.PHONY: install
install: build ## Install the binary to $GOPATH/bin
	@echo "$(BLUE)Installing $(BINARY_NAME)...$(NC)"
	@cp $(BUILD_DIR)/$(BINARY_NAME) $$(go env GOPATH)/bin/
	@echo "$(GREEN)$(BINARY_NAME) installed to $$(go env GOPATH)/bin/$(NC)"

.PHONY: version
version: ## Show version information
	@echo "$(BLUE)Version Information:$(NC)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Git Branch: $(GIT_BRANCH)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $$(go version)"

.PHONY: env
env: ## Show environment information
	@echo "$(BLUE)Environment Information:$(NC)"
	@echo "GOPATH: $$(go env GOPATH)"
	@echo "GOROOT: $$(go env GOROOT)"
	@echo "GOOS: $$(go env GOOS)"
	@echo "GOARCH: $$(go env GOARCH)"
	@echo "CGO_ENABLED: $(CGO_ENABLED)"

##@ CI/CD
.PHONY: ci
ci: deps-verify check test-coverage ## Run CI pipeline
	@echo "$(GREEN)CI pipeline completed successfully!$(NC)"

.PHONY: pre-commit
pre-commit: fmt vet test-short ## Run pre-commit checks
	@echo "$(GREEN)Pre-commit checks passed!$(NC)"

.PHONY: release
release: clean check test-coverage build-all ## Prepare release artifacts
	@echo "$(BLUE)Preparing release...$(NC)"
	@mkdir -p $(DIST_DIR)/checksums
	@cd $(DIST_DIR) && sha256sum * > checksums/sha256sums.txt 2>/dev/null || shasum -a 256 * > checksums/sha256sums.txt
	@echo "$(GREEN)Release artifacts ready in $(DIST_DIR)/$(NC)"

# Default target
.DEFAULT_GOAL := help
