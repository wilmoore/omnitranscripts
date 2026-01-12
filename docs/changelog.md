# Changelog

All notable changes to OmniTranscripts will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive documentation structure with detailed guides
- Repository badges for better project visibility
- Enhanced CLAUDE.md with documentation cross-references
- Live reload functionality for development dashboard
- Comprehensive .gitignore covering all use cases

### Changed
- Restructured README.md with better organization and navigation
- Enhanced project documentation with API, architecture, deployment, development, troubleshooting, and contributing guides

### Security
- Added comprehensive .gitignore to prevent accidental commit of sensitive data

## [1.0.0] - 2024-01-15

### Added
- Initial release of OmniTranscripts
- YouTube video transcription API with Go backend
- Dual framework support (Fiber and Encore.dev)
- Three-stage processing pipeline (download, normalize, transcribe)
- Synchronous processing for short videos (≤2 minutes)
- Asynchronous job queue for longer videos
- Real-time job status tracking with progress updates
- Multiple output formats (SRT, VTT, JSON, TSV, plain text)
- PostgreSQL database integration with comprehensive migrations
- Server-Sent Events (SSE) for live dashboard updates
- Web dashboard with real-time metrics and job monitoring
- Comprehensive test suite with unit, integration, and performance tests
- Docker containerization support
- Extensive Makefile with 30+ development commands
- Authentication via Bearer token API keys
- Rate limiting and security features
- Webhook support for job lifecycle notifications
- Multi-platform builds (Linux, macOS, Windows for AMD64/ARM64)
- Complete API documentation with OpenAPI/Swagger spec
- Performance benchmarking and load testing tools

### Technical Features
- **Core Technologies**: Go 1.23+, yt-dlp, FFmpeg, whisper.cpp
- **Frameworks**: Fiber v2 (HTTP), Encore.dev (production)
- **Database**: PostgreSQL with automatic migrations
- **Job Processing**: Thread-safe in-memory queue with pub/sub messaging
- **Audio Processing**: 16kHz mono WAV optimization for whisper.cpp
- **AI Transcription**: OpenAI Whisper C++ implementation
- **File Management**: Automatic cleanup and temporary file handling
- **Monitoring**: Comprehensive metrics collection and business analytics
- **Deployment**: Docker, Kubernetes, cloud platforms, traditional servers

### Architecture Highlights
- **Hybrid Processing**: Smart sync/async decision based on video length
- **Scalable Design**: Horizontal scaling with load balancing support
- **Performance Optimized**: Parallel processing and resource management
- **Production Ready**: Error handling, logging, monitoring, and observability
- **Developer Friendly**: Hot reload, comprehensive testing, detailed documentation

### Dependencies
- **Runtime**: yt-dlp, FFmpeg, whisper.cpp with ggml-base.en.bin model
- **Go Libraries**: fiber/v2, go-ytdlp, ffmpeg-go, uuid, godotenv
- **Database**: PostgreSQL (optional for Encore.dev deployments)
- **Development**: golangci-lint, goimports, testify

### Configuration
- Environment-based configuration with sensible defaults
- Docker and cloud deployment ready
- Comprehensive security settings
- Performance tuning options

---

## Version History

### [1.0.0] - 2024-01-15
- Initial public release
- Complete YouTube video transcription API
- Production-ready features and documentation

---

## Contributing

See [CONTRIBUTING.md](contributing.md) for details on how to contribute to this project.

## Support

- **Documentation**: [docs/](./README.md)
- **Issues**: [GitHub Issues](https://github.com/wilmoore/OmniTranscripts/issues)
- **Discussions**: [GitHub Discussions](https://github.com/wilmoore/OmniTranscripts/discussions)

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.