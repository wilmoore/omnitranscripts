# API Documentation

OmniTranscripts provides a simple REST API for transcribing audio and video from local files, direct media URLs, or 1000+ streaming platforms. Supports both synchronous and asynchronous processing.

## Base URL

```
https://your-domain.com
# or for local development
http://localhost:3000
```

## Authentication

All API endpoints (except `/health`) require authentication using Bearer tokens:

```bash
Authorization: Bearer YOUR_API_KEY
```

## Endpoints

### Health Check

#### `GET /health`

Check API health status. No authentication required.

**Response:**
```json
{
  "status": "ok",
  "message": "OmniTranscripts API is running"
}
```

**Example:**
```bash
curl http://localhost:3000/health
```

---

### Start Transcription

#### `POST /transcribe`

Submit media for transcription. Accepts either a URL (JSON) or a file upload (multipart/form-data). Returns immediate results for short media (≤2 min) or a job ID for longer media.

#### Option 1: URL-based (JSON)

**Request Body:**
```json
{
  "url": "https://www.youtube.com/watch?v=VIDEO_ID"
}
```

**Example:**
```bash
curl -X POST http://localhost:3000/transcribe \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ"}'
```

#### Option 2: File Upload (multipart/form-data)

Upload local audio or video files directly using multipart/form-data.

**Supported File Types:**
- Audio: `.mp3`, `.wav`, `.m4a`, `.flac`, `.ogg`, `.aac`
- Video: `.mp4`, `.mkv`, `.webm`, `.avi`, `.mov`

**Max File Size:** 500MB (configurable via `MAX_UPLOAD_SIZE` environment variable)

**Example:**
```bash
# Upload a local file
curl -X POST http://localhost:3000/transcribe \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -F "file=@/path/to/recording.mp4"

# Upload with explicit filename
curl -X POST http://localhost:3000/transcribe \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -F "file=@./podcast.mp3"
```

**Response (File Upload Error - Invalid Type):**
```json
{
  "error": "Unsupported file type",
  "supported_audio": [".mp3", ".wav", ".m4a", ".flac", ".ogg", ".aac"],
  "supported_video": [".mp4", ".mkv", ".webm", ".avi", ".mov"]
}
```

**Response (File Upload Error - Too Large):**
```json
{
  "error": "File too large",
  "max_size": 524288000
}
```

#### Responses

**Response (Short Media - Immediate):**
```json
{
  "transcript": "Complete transcript text...",
  "segments": [
    {
      "start": 0.0,
      "end": 3.5,
      "text": "First segment text"
    },
    {
      "start": 3.5,
      "end": 7.2,
      "text": "Second segment text"
    }
  ]
}
```

**Response (Long Media / File Upload - Async):**
```json
{
  "job_id": "job_1234567890"
}
```

**Error Response:**
```json
{
  "error": "Invalid URL. Must be a valid HTTP/HTTPS URL"
}
```

---

### Get Job Status

#### `GET /transcribe/{job_id}`

Retrieve the status and results of a transcription job.

**Parameters:**
- `job_id` (string): The job ID returned by the transcribe endpoint

**Response (Pending/Running):**
```json
{
  "id": "job_1234567890",
  "status": "running",
  "meta": {
    "source_type": "file",
    "input_format": "mp4"
  },
  "created_at": "2024-01-01T12:00:00Z"
}
```

**Response (Completed):**
```json
{
  "id": "job_1234567890",
  "status": "complete",
  "transcript": "Complete transcript text...",
  "segments": [
    {
      "start": 0.0,
      "end": 3.5,
      "text": "First segment text"
    }
  ],
  "meta": {
    "source_type": "url",
    "input_format": "mp3",
    "processing_time_ms": 42123
  },
  "created_at": "2024-01-01T12:00:00Z",
  "completed_at": "2024-01-01T12:02:30Z"
}
```

**Response (Failed):**
```json
{
  "id": "job_1234567890",
  "status": "error",
  "error": "Video download failed: Video unavailable",
  "created_at": "2024-01-01T12:00:00Z",
  "completed_at": "2024-01-01T12:01:15Z"
}
```

**Example:**
```bash
curl -X GET http://localhost:3000/transcribe/job_1234567890 \
  -H "Authorization: Bearer YOUR_API_KEY"
```

## Job Status Values

| Status | Description |
|--------|-------------|
| `pending` | Job created and queued for processing |
| `running` | Job is currently being processed |
| `complete` | Job completed successfully |
| `error` | Job failed with an error |

## Response Metadata

All job responses include a `meta` object with processing details:

| Field | Type | Description |
|-------|------|-------------|
| `source_type` | string | `"file"` or `"url"` - indicates input source |
| `input_format` | string | File extension (e.g., `"mp4"`, `"mp3"`) |
| `processing_time_ms` | int | Processing duration in milliseconds (only on completion) |

## Rate Limits

- **Free Tier**: 5 jobs per API key
- **Production**: Configurable limits based on your plan

## Error Codes

| HTTP Status | Description |
|-------------|-------------|
| `200` | Success |
| `400` | Bad Request (invalid URL, missing parameters) |
| `401` | Unauthorized (invalid or missing API key) |
| `404` | Not Found (job ID not found) |
| `429` | Too Many Requests (rate limit exceeded) |
| `500` | Internal Server Error |

## Supported Input Sources

### Local File Uploads
Upload audio or video files directly via `multipart/form-data`:
- **Audio**: `.mp3`, `.wav`, `.m4a`, `.flac`, `.ogg`, `.aac`
- **Video**: `.mp4`, `.mkv`, `.webm`, `.avi`, `.mov`
- **Max size**: 500MB (configurable via `MAX_UPLOAD_SIZE`)

### Direct Media URLs
Any publicly accessible audio or video URL:
- `https://example.com/podcast.mp3`
- `https://cdn.example.com/video.mp4`

### Streaming Platforms (via yt-dlp)
1000+ platforms supported including:
- YouTube (videos, Shorts, live streams after they end)
- Vimeo
- SoundCloud
- Twitter/X
- TikTok
- And many more...

**Limitations:**
- Maximum media length: 30 minutes (configurable via `MAX_VIDEO_LENGTH`)
- Private/paywalled content: Not supported
- Copyright-protected content: May fail depending on restrictions

## Response Formats

### Transcript Segments

Each segment in the `segments` array contains:

```json
{
  "start": 0.0,        // Start time in seconds
  "end": 3.5,          // End time in seconds
  "text": "Spoken text" // Transcribed text for this segment
}
```

### Subtitle Files

When transcription completes, subtitle files are automatically generated:
- **SRT format**: Standard subtitle format for video players
- **VTT format**: WebVTT format for web players
- **JSON format**: Raw transcript data with timestamps
- **TSV format**: Tab-separated values for data analysis

## SDK Examples

### JavaScript/Node.js

```javascript
class OmniTranscriptsAPI {
  constructor(apiKey, baseURL = 'http://localhost:3000') {
    this.apiKey = apiKey;
    this.baseURL = baseURL;
  }

  async transcribe(url) {
    const response = await fetch(`${this.baseURL}/transcribe`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ url })
    });

    return await response.json();
  }

  async getJobStatus(jobId) {
    const response = await fetch(`${this.baseURL}/transcribe/${jobId}`, {
      headers: {
        'Authorization': `Bearer ${this.apiKey}`
      }
    });

    return await response.json();
  }

  async waitForCompletion(jobId, pollInterval = 5000) {
    while (true) {
      const result = await this.getJobStatus(jobId);

      if (result.status === 'complete' || result.status === 'error') {
        return result;
      }

      await new Promise(resolve => setTimeout(resolve, pollInterval));
    }
  }
}

// Usage
const api = new OmniTranscriptsAPI('your-api-key');
const result = await api.transcribe('https://www.youtube.com/watch?v=dQw4w9WgXcQ');

if (result.job_id) {
  const finalResult = await api.waitForCompletion(result.job_id);
  console.log(finalResult.transcript);
} else {
  console.log(result.transcript); // Short video, immediate result
}
```

### Python

```python
import requests
import time

class OmniTranscriptsAPI:
    def __init__(self, api_key, base_url='http://localhost:3000'):
        self.api_key = api_key
        self.base_url = base_url
        self.headers = {'Authorization': f'Bearer {api_key}'}

    def transcribe(self, url):
        response = requests.post(
            f'{self.base_url}/transcribe',
            headers={**self.headers, 'Content-Type': 'application/json'},
            json={'url': url}
        )
        return response.json()

    def get_job_status(self, job_id):
        response = requests.get(
            f'{self.base_url}/transcribe/{job_id}',
            headers=self.headers
        )
        return response.json()

    def wait_for_completion(self, job_id, poll_interval=5):
        while True:
            result = self.get_job_status(job_id)

            if result['status'] in ['complete', 'error']:
                return result

            time.sleep(poll_interval)

# Usage
api = OmniTranscriptsAPI('your-api-key')
result = api.transcribe('https://www.youtube.com/watch?v=dQw4w9WgXcQ')

if 'job_id' in result:
    final_result = api.wait_for_completion(result['job_id'])
    print(final_result['transcript'])
else:
    print(result['transcript'])  # Short video, immediate result
```

## MCP Server (ChatGPT Integration)

OmniTranscripts includes an MCP (Model Context Protocol) server for integration with ChatGPT via the OpenAI Apps SDK.

### Endpoint

```
POST/GET/DELETE /mcp
```

### MCP Tools

| Tool | Description |
|------|-------------|
| `transcribe_url` | Start transcription of a video/audio URL, returns job_id |
| `get_transcription` | Check status and retrieve transcript for a job |

### Configuration

```bash
MCP_ENABLED=true   # Enable/disable MCP server (default: true)
MCP_ENDPOINT=/mcp  # Endpoint path (default: /mcp)
```

For detailed setup instructions, see [ChatGPT Integration Guide](chatgpt-integration.md).

---

## Command-Line Interface (CLI)

The `make transcribe` command provides a convenient way to transcribe media without starting the HTTP server.

### Basic Usage

```bash
# Transcribe a URL
make transcribe URL="https://www.youtube.com/watch?v=VIDEO_ID"

# Transcribe a local file
make transcribe URL="/path/to/video.mp4"
```

### Output Handling

The CLI outputs:
- **stdout**: Clean transcript text (suitable for piping to other tools)
- **stderr**: Progress messages and diagnostics

This separation enables composable Unix workflows.

### Copy to Clipboard

```bash
# macOS
make transcribe URL="https://www.youtube.com/watch?v=VIDEO_ID" | pbcopy

# Linux (xclip)
make transcribe URL="https://www.youtube.com/watch?v=VIDEO_ID" | xclip -selection clipboard

# Linux (xsel)
make transcribe URL="https://www.youtube.com/watch?v=VIDEO_ID" | xsel --clipboard --input
```

### Save to File

```bash
# Save transcript
make transcribe URL="https://www.youtube.com/watch?v=VIDEO_ID" > transcript.txt

# Append to file
make transcribe URL="..." >> all-transcripts.txt
```

### Suppress Progress Output

```bash
# Hide progress messages while piping
make transcribe URL="..." 2>/dev/null | pbcopy
```

### Text Processing

```bash
# Word count
make transcribe URL="..." 2>/dev/null | wc -w

# Line count
make transcribe URL="..." 2>/dev/null | wc -l

# Search for keywords
make transcribe URL="..." | grep -i "keyword"

# Get first N lines
make transcribe URL="..." | head -20
```

### Exit Codes

- `0`: Success
- `1`: Error (invalid URL, file not found, transcription failed, etc.)

Check stderr for error details:
```bash
make transcribe URL="..." 2>&1 | head  # See error messages
```

---

## Webhooks (Coming Soon)

Future versions will support webhook notifications for job completion:

```json
{
  "event": "transcription.completed",
  "job_id": "job_1234567890",
  "status": "complete",
  "timestamp": "2024-01-01T12:02:30Z"
}
```