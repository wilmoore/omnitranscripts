# API Marketplace Tracking

Track the status of OmniTranscripts listings across API marketplaces.

## Marketplaces (Priority Order)

| Priority | Marketplace | Reach | Status | URL | Notes |
|----------|-------------|-------|--------|-----|-------|
| 1 | **RapidAPI** | 4M+ developers | Not Started | - | Largest marketplace, easiest integration |
| 2 | **API Layer** | Enterprise focus | Not Started | - | Premium positioning possible |
| 3 | **Postman API Network** | Millions via Postman | Not Started | - | Great for developer discovery |
| 4 | **AWS Marketplace** | Enterprise buyers | Not Started | - | Requires more setup, higher ticket |
| 5 | **Azure Marketplace** | Enterprise buyers | Not Started | - | Similar to AWS, different audience |
| 6 | **Mashape (Rakuten)** | International reach | Not Started | - | Good for expansion |

## Status Legend

- **Not Started** - Marketplace listing not yet created
- **In Progress** - Currently working on listing
- **Pending Review** - Submitted, awaiting marketplace approval
- **Live** - Listing is active and accepting traffic
- **Paused** - Listing temporarily disabled

---

## RapidAPI Integration

### Requirements
- [x] OpenAPI 3.0 specification (`docs/swagger.yaml`)
- [ ] RapidAPI account created
- [ ] API listing submitted
- [ ] Pricing tiers configured
- [ ] API testing verified

### Pricing Tiers (Proposed)

| Tier | Price | Requests/Month | Features |
|------|-------|----------------|----------|
| **Free** | $0 | 50 | Basic transcription, 5 min max video |
| **Basic** | $9.99 | 500 | 15 min max video, SRT/VTT subtitles |
| **Pro** | $29.99 | 2,000 | 30 min max video, priority queue |
| **Enterprise** | $99.99 | 10,000 | Unlimited video length, webhooks |

### API Listing Description

**Title**: OmniTranscripts - Universal Media Transcription API

**Short Description**:
Transcribe audio and video from 1000+ platforms including YouTube, Vimeo, TikTok, SoundCloud, and more. Get accurate transcripts with timestamps in seconds.

**Long Description**:
OmniTranscripts is a powerful transcription API that converts audio and video from any URL into accurate text transcripts with precise timestamps.

**Key Features:**
- **Universal Platform Support**: Works with YouTube, Vimeo, TikTok, Twitter, SoundCloud, Instagram, and 1000+ more platforms via yt-dlp
- **Audio-First Design**: Direct audio file URLs are first-class citizens
- **Timestamp Precision**: Get word-level timestamps for subtitles
- **Multiple Output Formats**: JSON, SRT, VTT subtitle formats
- **Async Processing**: Long videos process in background with job status polling
- **Webhook Support**: Get notified when transcription completes

**Use Cases:**
- Content creators needing video subtitles
- Podcast transcription services
- Meeting recording analysis
- Social media content archival
- Accessibility compliance

---

## API Layer Integration

### Requirements
- [ ] API Layer account created
- [ ] Premium listing applied
- [ ] Enterprise pricing configured
- [ ] SLA documentation prepared

---

## Postman API Network

### Requirements
- [ ] Postman collection created
- [ ] Collection published to API Network
- [ ] Documentation verified
- [ ] Run in Postman button added to README

---

## Progress Log

| Date | Marketplace | Action | Status |
|------|-------------|--------|--------|
| 2026-01-10 | All | Created tracking document | In Progress |

---

## Next Steps

1. **RapidAPI** (Priority 1):
   - Create RapidAPI account
   - Import OpenAPI spec
   - Configure pricing tiers
   - Submit for review

2. **Postman** (Priority 3):
   - Create Postman collection from swagger.yaml
   - Publish to API Network
   - Add "Run in Postman" button to README

3. **API Layer** (Priority 2):
   - Apply for premium listing
   - Prepare enterprise documentation
