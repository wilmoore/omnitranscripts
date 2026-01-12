# API Documentation

OmniTranscripts provides a simple REST API for transcribing YouTube videos with support for both synchronous and asynchronous processing.

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

Submit a YouTube video for transcription. Returns immediate results for short videos (≤2 min) or a job ID for longer videos.

**Request Body:**
```json
{
  "url": "https://www.youtube.com/watch?v=VIDEO_ID"
}
```

**Response (Short Videos - Immediate):**
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

**Response (Long Videos - Async):**
```json
{
  "job_id": "job_1234567890"
}
```

**Error Response:**
```json
{
  "error": "Invalid YouTube URL"
}
```

**Example:**
```bash
curl -X POST http://localhost:3000/transcribe \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ"}'
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
  "created_at": "2024-01-01T12:00:00Z",
  "completed_at": "2024-01-01T12:02:30Z",
  "subtitle_files": {
    "srt_url": "https://your-domain.com/transcripts/job_1234567890/subtitles.srt",
    "vtt_url": "https://your-domain.com/transcripts/job_1234567890/subtitles.vtt"
  }
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

## Supported Video Formats

The API supports any YouTube video that can be downloaded by yt-dlp:
- Standard YouTube videos
- YouTube Shorts
- Live streams (after they end)
- Age-restricted videos (with appropriate access)

**Limitations:**
- Maximum video length: 30 minutes (configurable)
- Private videos: Not supported
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