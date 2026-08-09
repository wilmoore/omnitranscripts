# Harden Dashboard Identifiers and Subprocess Inputs

**Backlog:** 2
**Branch:** `fix/security-dashboard-input-hardening`
**Status:** Completed
**Started:** 2026-08-09T22:30:08Z

## Reconciliation

The original audit finding called this command injection. The dashboard uses `exec.Command` directly, not a shell, so shell metacharacters are not evaluated. The remaining security issue is valid: media URLs are unvalidated, persisted video identifiers reach filesystem paths and a regular expression, non-YouTube URLs collapse to `unknown`, and subprocess option parsing is not explicitly terminated.

No existing ADR governs this input boundary. The change is defensive validation without a new architectural decision.

## Requirements

- Accept only absolute HTTP(S) media URLs with a hostname and no embedded credentials.
- Preserve valid YouTube IDs; derive a deterministic, collision-resistant, identifier-safe value for other URLs.
- Bound identifiers to ASCII letters, digits, `_`, and `-` before filesystem or process-pattern use.
- Reject unsafe identifiers loaded from persisted job data rather than normalizing them into a different path.
- Quote identifiers embedded in the `pgrep` regular expression.
- Pass `--` before the media URL supplied to `yt-dlp`.
- Return a stable `400` response for invalid URL submissions.

## Implementation

1. Add focused validation/identifier helpers in a normal internal package so the ignored dashboard entrypoint remains testable.
2. Apply URL validation and safe identifier derivation in job creation.
3. Guard transcript directory access during status and metric updates.
4. Quote process-match input and terminate downloader option parsing.
5. Add table-driven tests for URL schemes, credentials, YouTube extraction, deterministic fallback IDs, path traversal, length bounds, and regex metacharacters.

## Verification

- [x] `go test ./internal/dashboardsecurity -count=1`
- [x] `go vet ./internal/dashboardsecurity`
- [x] `go build ./...`
- [x] `go vet ./...`
- [x] `go build -o /tmp/omnitranscripts-web-dashboard web-dashboard.go`
- [x] `go vet web-dashboard.go`
- [x] Backlog structural validation and `git diff --check`

The full suite is intentionally deferred to backlog item 4, which already records the reproducible hanging-test condition and is dependency-ordered after duration parsing.

## Result

- Invalid, relative, credential-bearing, and non-HTTP(S) URLs return `400` before job persistence or subprocess use.
- YouTube watch, short, Shorts, embed, and live URLs retain their native 11-character ID.
- Other providers receive a deterministic SHA-256-derived identifier instead of sharing `unknown`.
- Persisted identifiers are rejected before transcript or log filesystem access.
- Dynamic process matching is regular-expression quoted and downloader option parsing ends before the URL.

## Risks

- Existing persisted jobs with unsafe identifiers will no longer have their transcript directories inspected; they remain visible but cannot trigger filesystem traversal.
- Deterministic IDs for non-YouTube sources change new dashboard output directory names from the shared `unknown` directory.
