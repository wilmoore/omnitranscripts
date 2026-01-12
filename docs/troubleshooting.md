# Troubleshooting Guide

This guide covers common issues and solutions when developing, deploying, or using OmniTranscripts.

## Common Issues

### 1. External Tool Dependencies

#### "whisper.cpp: command not found"

**Problem**: The whisper.cpp binary is not found in the system PATH.

**Solutions**:

1. **Install whisper.cpp properly**:
   ```bash
   git clone https://github.com/ggerganov/whisper.cpp.git
   cd whisper.cpp
   make
   sudo cp main /usr/local/bin/whisper.cpp
   ```

2. **Verify installation**:
   ```bash
   which whisper.cpp
   whisper.cpp --help
   ```

3. **Alternative: Update PATH**:
   ```bash
   export PATH=$PATH:/path/to/whisper.cpp/directory
   echo 'export PATH=$PATH:/path/to/whisper.cpp' >> ~/.bashrc
   ```

4. **Set custom path in environment**:
   ```env
   WHISPER_PATH=/custom/path/to/whisper.cpp
   ```

#### "ggml-base.en.bin: no such file or directory"

**Problem**: The whisper.cpp model file is missing.

**Solutions**:

1. **Download the model**:
   ```bash
   cd whisper.cpp
   bash ./models/download-ggml-model.sh base.en
   ```

2. **Copy to standard location**:
   ```bash
   sudo mkdir -p /usr/local/share/whisper
   sudo cp models/ggml-base.en.bin /usr/local/share/whisper/
   ```

3. **Verify model path**:
   ```bash
   ls -la /usr/local/share/whisper/ggml-base.en.bin
   ```

4. **Alternative models**:
   ```bash
   # Download different model sizes
   bash ./models/download-ggml-model.sh tiny.en    # Fastest, least accurate
   bash ./models/download-ggml-model.sh small.en   # Good balance
   bash ./models/download-ggml-model.sh medium.en  # Better accuracy
   bash ./models/download-ggml-model.sh large      # Best accuracy, slowest
   ```

#### "yt-dlp failed" or "youtube-dl not found"

**Problem**: YouTube downloader is missing or outdated.

**Solutions**:

1. **Install/update yt-dlp**:
   ```bash
   # Using pip (recommended)
   pip install --upgrade yt-dlp

   # Using package manager
   brew install yt-dlp  # macOS
   sudo apt install yt-dlp  # Ubuntu/Debian
   ```

2. **Verify installation**:
   ```bash
   yt-dlp --version
   yt-dlp --help
   ```

3. **Test with a video**:
   ```bash
   yt-dlp --get-title "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
   ```

4. **Common yt-dlp issues**:
   ```bash
   # Update to latest version (YouTube changes frequently)
   pip install --upgrade yt-dlp

   # Clear cache if downloads fail
   yt-dlp --rm-cache-dir

   # Use different extractor if needed
   yt-dlp --extractor-args "youtube:player_client=android"
   ```

#### "ffmpeg: command not found"

**Problem**: FFmpeg is not installed or not in PATH.

**Solutions**:

1. **Install FFmpeg**:
   ```bash
   # macOS
   brew install ffmpeg

   # Ubuntu/Debian
   sudo apt update && sudo apt install ffmpeg

   # CentOS/RHEL
   sudo yum install ffmpeg

   # Windows (using Chocolatey)
   choco install ffmpeg
   ```

2. **Verify installation**:
   ```bash
   ffmpeg -version
   ```

3. **Test audio processing**:
   ```bash
   ffmpeg -i input.mp4 -ar 16000 -ac 1 output.wav
   ```

### 2. Application Startup Issues

#### "Port already in use"

**Problem**: The specified port is already occupied.

**Solutions**:

1. **Find process using the port**:
   ```bash
   # Find process on port 3000
   lsof -i :3000
   netstat -tulpn | grep :3000
   ```

2. **Kill the process**:
   ```bash
   sudo kill -9 <PID>
   ```

3. **Use a different port**:
   ```env
   PORT=3001
   ```

4. **Auto-find available port** (web-dashboard.go does this automatically):
   ```go
   func findAvailablePort(startPort int) int {
       for port := startPort; port < startPort+100; port++ {
           // Try to bind to port
       }
   }
   ```

#### "Database connection failed"

**Problem**: Cannot connect to PostgreSQL database (Encore.dev deployments).

**Solutions**:

1. **Check database status**:
   ```bash
   # Local PostgreSQL
   sudo systemctl status postgresql
   pg_isready -h localhost -p 5432

   # Encore.dev
   encore db status
   ```

2. **Verify connection string**:
   ```bash
   # Test connection
   psql "postgresql://user:pass@host:5432/dbname"
   ```

3. **Reset Encore.dev database**:
   ```bash
   encore db reset
   encore db migrate
   ```

4. **Check firewall/network**:
   ```bash
   telnet database-host 5432
   nmap -p 5432 database-host
   ```

#### "Invalid API key" errors

**Problem**: Authentication failures.

**Solutions**:

1. **Verify environment variable**:
   ```bash
   echo $API_KEY
   grep API_KEY .env
   ```

2. **Check API key format**:
   ```bash
   # Should be a non-empty string
   curl -H "Authorization: Bearer your-api-key" http://localhost:3000/health
   ```

3. **Test with curl**:
   ```bash
   curl -X POST http://localhost:3000/transcribe \
     -H "Authorization: Bearer dev-api-key-12345" \
     -H "Content-Type: application/json" \
     -d '{"url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ"}'
   ```

### 3. Processing Issues

#### Jobs stuck in "pending" status

**Problem**: Jobs are not being processed.

**Solutions**:

1. **Check worker status**:
   ```bash
   # Look for processing logs
   tail -f /var/log/videotranscript.log
   ```

2. **Verify external tools**:
   ```bash
   which yt-dlp ffmpeg whisper.cpp
   ```

3. **Check disk space**:
   ```bash
   df -h $WORK_DIR
   du -sh $WORK_DIR/*
   ```

4. **Restart job processing**:
   ```bash
   # For systemd service
   sudo systemctl restart videotranscript

   # For Docker
   docker restart videotranscript-container
   ```

5. **Clear job queue** (development only):
   ```bash
   # Delete jobs.json file
   rm jobs.json

   # Or reset database
   encore db reset
   ```

#### "Video download failed"

**Problem**: Cannot download videos from YouTube.

**Solutions**:

1. **Check video accessibility**:
   ```bash
   yt-dlp --get-title "https://www.youtube.com/watch?v=VIDEO_ID"
   ```

2. **Common video issues**:
   - **Private videos**: Not supported
   - **Age-restricted**: May require authentication
   - **Geo-blocked**: Use VPN or different region
   - **Copyright-protected**: May be blocked
   - **Live streams**: Only supported after stream ends

3. **Update yt-dlp**:
   ```bash
   pip install --upgrade yt-dlp
   ```

4. **Use alternative extractors**:
   ```bash
   yt-dlp --extractor-args "youtube:player_client=android" URL
   ```

5. **Check network connectivity**:
   ```bash
   curl -I https://www.youtube.com/
   ping youtube.com
   ```

#### "Audio extraction failed"

**Problem**: FFmpeg cannot process the downloaded audio.

**Solutions**:

1. **Check audio file**:
   ```bash
   ffprobe downloaded_file.webm
   file downloaded_file.webm
   ```

2. **Test FFmpeg manually**:
   ```bash
   ffmpeg -i input.webm -ar 16000 -ac 1 output.wav
   ```

3. **Try different audio format**:
   ```bash
   yt-dlp --audio-format wav --extract-audio URL
   ```

4. **Check available codecs**:
   ```bash
   ffmpeg -codecs | grep -i audio
   ```

#### "Transcription failed"

**Problem**: whisper.cpp cannot process the audio file.

**Solutions**:

1. **Check audio file format**:
   ```bash
   ffprobe -v quiet -print_format json -show_format audio.wav
   ```

2. **Test whisper.cpp manually**:
   ```bash
   whisper.cpp -m models/ggml-base.en.bin -f audio.wav
   ```

3. **Try different model**:
   ```bash
   # Download and try tiny model (faster, less accurate)
   bash ./models/download-ggml-model.sh tiny.en
   whisper.cpp -m models/ggml-tiny.en.bin -f audio.wav
   ```

4. **Check audio duration**:
   ```bash
   # Very long files may timeout
   ffprobe -v quiet -show_entries format=duration -of csv=p=0 audio.wav
   ```

### 4. Performance Issues

#### Slow transcription processing

**Problem**: Transcription takes too long.

**Solutions**:

1. **Use faster model**:
   ```bash
   # tiny.en: ~32x faster than large, less accurate
   # small.en: ~6x faster than large, good accuracy
   # base.en: ~4x faster than large, good accuracy
   ```

2. **Optimize server resources**:
   ```bash
   # Check CPU usage
   top
   htop

   # Check memory usage
   free -h
   ```

3. **Enable CPU optimizations**:
   ```bash
   # Build whisper.cpp with optimizations
   cd whisper.cpp
   make clean
   make CFLAGS="-O3 -march=native"
   ```

4. **Parallel processing**:
   ```go
   // Increase worker count for multiple jobs
   const MaxWorkers = 4  // Adjust based on CPU cores
   ```

#### High memory usage

**Problem**: Application consumes too much memory.

**Solutions**:

1. **Monitor memory usage**:
   ```bash
   # Check process memory
   ps aux | grep videotranscript
   top -p $(pgrep videotranscript)
   ```

2. **Optimize Go garbage collection**:
   ```bash
   export GOGC=100  # Default GC target
   export GOMEMLIMIT=2GiB  # Set memory limit
   ```

3. **Clean temporary files**:
   ```bash
   # Set up automatic cleanup
   find $WORK_DIR -type f -mtime +1 -delete
   ```

4. **Limit concurrent jobs**:
   ```go
   const MaxConcurrentJobs = 2  // Reduce if memory constrained
   ```

#### Disk space issues

**Problem**: Running out of disk space.

**Solutions**:

1. **Check disk usage**:
   ```bash
   df -h
   du -sh $WORK_DIR
   ```

2. **Clean old files**:
   ```bash
   # Clean files older than 1 day
   find $WORK_DIR -type f -mtime +1 -delete

   # Clean specific file types
   find $WORK_DIR -name "*.wav" -mtime +1 -delete
   find $WORK_DIR -name "*.mp4" -mtime +1 -delete
   ```

3. **Set up automated cleanup**:
   ```bash
   # Add to crontab
   0 2 * * * find /tmp/videotranscript -type f -mtime +1 -delete
   ```

4. **Use different storage location**:
   ```env
   WORK_DIR=/mnt/large-disk/videotranscript
   ```

### 5. Network and Connectivity Issues

#### "Connection timeout" errors

**Problem**: Network timeouts during video download.

**Solutions**:

1. **Increase timeout values**:
   ```go
   const DownloadTimeout = 300 * time.Second  // 5 minutes
   ```

2. **Check network connectivity**:
   ```bash
   ping google.com
   curl -I https://www.youtube.com/
   ```

3. **Test download speed**:
   ```bash
   wget --progress=bar --show-progress https://www.youtube.com/watch?v=dQw4w9WgXcQ
   ```

4. **Use proxy if needed**:
   ```bash
   export HTTP_PROXY=http://proxy:8080
   export HTTPS_PROXY=http://proxy:8080
   ```

#### Rate limiting issues

**Problem**: Too many requests to YouTube.

**Solutions**:

1. **Implement request throttling**:
   ```go
   time.Sleep(1 * time.Second)  // Wait between requests
   ```

2. **Use different IP addresses**:
   ```bash
   # Rotate through multiple servers
   # or use VPN
   ```

3. **Respect YouTube's rate limits**:
   ```go
   const MaxRequestsPerMinute = 30
   ```

### 6. Docker and Deployment Issues

#### Container startup failures

**Problem**: Docker container won't start.

**Solutions**:

1. **Check container logs**:
   ```bash
   docker logs videotranscript-container
   docker logs --follow videotranscript-container
   ```

2. **Verify image build**:
   ```bash
   docker build -t videotranscript-app .
   docker run --rm -it videotranscript-app /bin/sh
   ```

3. **Check resource limits**:
   ```bash
   # Increase memory limit
   docker run -m 2g videotranscript-app
   ```

4. **Test dependencies in container**:
   ```bash
   docker run --rm -it videotranscript-app which yt-dlp ffmpeg whisper.cpp
   ```

#### Permission issues

**Problem**: File permission errors in containers.

**Solutions**:

1. **Check file ownership**:
   ```bash
   ls -la $WORK_DIR
   ```

2. **Fix permissions**:
   ```bash
   sudo chown -R $(whoami):$(whoami) $WORK_DIR
   chmod 755 $WORK_DIR
   ```

3. **Run container as user**:
   ```dockerfile
   USER 1001:1001
   ```

4. **Mount volumes correctly**:
   ```bash
   docker run -v /host/path:/container/path:rw videotranscript-app
   ```

## Debugging Tools

### Enable Debug Mode

```bash
# Environment variable
DEBUG=1 go run main.go

# Or in .env file
echo "DEBUG=1" >> .env
```

### Logging Configuration

```go
// Set log level
log.SetLevel(log.DebugLevel)

// Enable structured logging
log.WithFields(log.Fields{
    "job_id": jobID,
    "stage":  "download",
}).Debug("Starting video download")
```

### Health Checks

```bash
# Basic health check
curl http://localhost:3000/health

# Detailed status (if implemented)
curl http://localhost:3000/status

# Check all dependencies
curl http://localhost:3000/debug/dependencies
```

### Performance Profiling

```bash
# Enable profiling
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# View profiles
go tool pprof -http=:8080 profile.pb.gz
```

## Getting Help

### Log Analysis

1. **Application logs**: Check for error messages and stack traces
2. **System logs**: Check `/var/log/` for system-level issues
3. **Service logs**: `journalctl -u videotranscript` for systemd services

### Community Support

1. **GitHub Issues**: Report bugs and feature requests
2. **Discussions**: Ask questions in GitHub Discussions
3. **Documentation**: Check the [docs](../docs/) folder for detailed guides

### Professional Support

For enterprise deployments or complex issues:
1. **Encore.dev Support**: For production deployments
2. **Custom Consulting**: Available for specific requirements

Remember to include relevant log outputs, error messages, and system information when seeking help.