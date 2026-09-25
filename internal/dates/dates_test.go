package dates

import (
	"testing"
	"time"
)

var now = time.Date(2026, 9, 25, 12, 0, 0, 0, time.Local)

func TestResolve(t *testing.T) {
	for text, want := range map[string]string{"today": "2026-09-25", "Yesterday": "2026-09-24", "2026-02-28": "2026-02-28"} {
		if got, err := Resolve(text, now); err != nil || got != want {
			t.Errorf("Resolve(%q) = %q, %v; want %q", text, got, err, want)
		}
	}
	for _, text := range []string{"2026-13-01", "2026-02-30", "2026-9-5", "next week"} {
		if got, err := Resolve(text, now); err == nil {
			t.Errorf("Resolve(%q) = %q; want an error", text, got)
		}
	}
}

func TestHumanize(t *testing.T) {
	for text, want := range map[string]string{"2026-09-24": "Thu, Sep 24", "2025-12-31": "Wed, Dec 31, 2025", "nope": "nope"} {
		if got := Humanize(text, now); got != want {
			t.Errorf("Humanize(%q) = %q; want %q", text, got, want)
		}
	}
}

func TestStartOfWeek(t *testing.T) {
	if got := StartOfWeek(now); got != "2026-09-21" {
		t.Errorf("StartOfWeek = %q", got)
	}
	sunday := time.Date(2026, 9, 27, 12, 0, 0, 0, time.Local)
	if got := StartOfWeek(sunday); got != "2026-09-21" {
		t.Errorf("StartOfWeek(Sunday) = %q", got)
	}
}
