package dashboardsecurity

import (
	"regexp"
	"strings"
	"testing"
)

func TestParseMediaURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "https", raw: "https://example.com/media.mp3"},
		{name: "http", raw: "http://example.com/video.mp4"},
		{name: "trim whitespace", raw: "  https://example.com/media.mp3  "},
		{name: "missing", raw: "", wantErr: true},
		{name: "relative", raw: "/media.mp3", wantErr: true},
		{name: "file scheme", raw: "file:///etc/passwd", wantErr: true},
		{name: "ftp scheme", raw: "ftp://example.com/media.mp3", wantErr: true},
		{name: "missing hostname", raw: "https:///media.mp3", wantErr: true},
		{name: "embedded credentials", raw: "https://user:pass@example.com/media.mp3", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseMediaURL(test.raw)
			if (err != nil) != test.wantErr {
				t.Fatalf("ParseMediaURL(%q) error = %v, wantErr %v", test.raw, err, test.wantErr)
			}
		})
	}
}

func TestMediaID(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "watch", raw: "https://www.youtube.com/watch?v=dQw4w9WgXcQ", want: "dQw4w9WgXcQ"},
		{name: "short URL", raw: "https://youtu.be/dQw4w9WgXcQ", want: "dQw4w9WgXcQ"},
		{name: "shorts", raw: "https://youtube.com/shorts/dQw4w9WgXcQ", want: "dQw4w9WgXcQ"},
		{name: "embed", raw: "https://youtube.com/embed/dQw4w9WgXcQ", want: "dQw4w9WgXcQ"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := ParseMediaURL(test.raw)
			if err != nil {
				t.Fatal(err)
			}
			if got := MediaID(parsed); got != test.want {
				t.Fatalf("MediaID(%q) = %q, want %q", test.raw, got, test.want)
			}
		})
	}

	first, err := ParseMediaURL("https://example.com/media.mp3")
	if err != nil {
		t.Fatal(err)
	}
	second, err := ParseMediaURL("https://example.com/other.mp3")
	if err != nil {
		t.Fatal(err)
	}

	firstID := MediaID(first)
	if firstID != MediaID(first) {
		t.Fatal("fallback media ID is not deterministic")
	}
	if firstID == MediaID(second) {
		t.Fatal("different URLs produced the same fallback media ID")
	}
	if !strings.HasPrefix(firstID, "media_") {
		t.Fatalf("fallback media ID %q lacks media_ prefix", firstID)
	}
	if err := ValidateIdentifier(firstID); err != nil {
		t.Fatalf("fallback media ID is unsafe: %v", err)
	}
}

func TestValidateIdentifier(t *testing.T) {
	valid := []string{"dQw4w9WgXcQ", "job_123", "media_deadbeef"}
	for _, identifier := range valid {
		if err := ValidateIdentifier(identifier); err != nil {
			t.Errorf("ValidateIdentifier(%q) returned %v", identifier, err)
		}
	}

	invalid := []string{"", "../outside", "nested/path", "line\nbreak", "regex.*", strings.Repeat("a", 129)}
	for _, identifier := range invalid {
		if err := ValidateIdentifier(identifier); err == nil {
			t.Errorf("ValidateIdentifier(%q) succeeded", identifier)
		}
	}
}

func TestSafePathsAndProcessPattern(t *testing.T) {
	dir, err := Directory("transcripts", "safe-id")
	if err != nil || dir != "transcripts/safe-id" {
		t.Fatalf("Directory() = %q, %v", dir, err)
	}
	if _, err := Directory("transcripts", "../outside"); err == nil {
		t.Fatal("Directory accepted traversal identifier")
	}

	file, err := File("logs", "job_123", ".log")
	if err != nil || file != "logs/job_123.log" {
		t.Fatalf("File() = %q, %v", file, err)
	}

	pattern, err := ProcessPattern("safe-id")
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(pattern).MatchString("transcribe --job safe-id") {
		t.Fatalf("pattern %q did not match expected process", pattern)
	}
	if _, err := ProcessPattern("unsafe.*"); err == nil {
		t.Fatal("ProcessPattern accepted regex metacharacters")
	}
}
