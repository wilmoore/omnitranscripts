package dashboardsecurity

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

const maxIdentifierLength = 128

var (
	identifierPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	youtubeIDPattern  = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)
)

// ParseMediaURL validates an externally supplied media URL before it reaches a
// downloader subprocess. Only absolute HTTP(S) URLs without embedded
// credentials are accepted.
func ParseMediaURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("media URL is required")
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("parse media URL: %w", err)
	}

	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return nil, errors.New("media URL must use http or https")
	}
	if parsed.Hostname() == "" {
		return nil, errors.New("media URL must include a hostname")
	}
	if parsed.User != nil {
		return nil, errors.New("media URL must not contain credentials")
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	return parsed, nil
}

// MediaID returns the native YouTube ID when one is available. Other sources
// receive a deterministic hash-derived identifier instead of sharing an
// "unknown" directory.
func MediaID(mediaURL *url.URL) string {
	if id := youtubeID(mediaURL); id != "" {
		return id
	}

	sum := sha256.Sum256([]byte(mediaURL.String()))
	return "media_" + hex.EncodeToString(sum[:12])
}

// ValidateIdentifier rejects persisted identifiers before they are used in a
// filesystem path, response filename, or process-match expression.
func ValidateIdentifier(identifier string) error {
	if identifier == "" {
		return errors.New("identifier is required")
	}
	if len(identifier) > maxIdentifierLength {
		return fmt.Errorf("identifier exceeds %d characters", maxIdentifierLength)
	}
	if !identifierPattern.MatchString(identifier) {
		return errors.New("identifier contains unsupported characters")
	}
	return nil
}

// Directory joins a trusted root to a validated identifier.
func Directory(root, identifier string) (string, error) {
	if root == "" {
		return "", errors.New("directory root is required")
	}
	if err := ValidateIdentifier(identifier); err != nil {
		return "", err
	}
	return filepath.Join(root, identifier), nil
}

// File joins a trusted root to a validated identifier and trusted suffix.
func File(root, identifier, suffix string) (string, error) {
	if root == "" {
		return "", errors.New("file root is required")
	}
	if err := ValidateIdentifier(identifier); err != nil {
		return "", err
	}
	return filepath.Join(root, identifier+suffix), nil
}

// ProcessPattern quotes dynamic text even though accepted identifiers cannot
// currently contain regular-expression metacharacters. This keeps the safety
// property explicit if the identifier grammar changes later.
func ProcessPattern(identifier string) (string, error) {
	if err := ValidateIdentifier(identifier); err != nil {
		return "", err
	}
	return `transcribe.*` + regexp.QuoteMeta(identifier), nil
}

func youtubeID(mediaURL *url.URL) string {
	host := strings.TrimSuffix(strings.ToLower(mediaURL.Hostname()), ".")
	var candidate string

	switch host {
	case "youtu.be":
		candidate = firstPathSegment(mediaURL.Path)
	case "youtube.com", "www.youtube.com", "m.youtube.com", "music.youtube.com":
		candidate = mediaURL.Query().Get("v")
		if candidate == "" {
			segments := strings.Split(strings.Trim(mediaURL.Path, "/"), "/")
			if len(segments) >= 2 && (segments[0] == "embed" || segments[0] == "shorts" || segments[0] == "live") {
				candidate = segments[1]
			}
		}
	}

	if youtubeIDPattern.MatchString(candidate) {
		return candidate
	}
	return ""
}

func firstPathSegment(path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) == 0 {
		return ""
	}
	return segments[0]
}
