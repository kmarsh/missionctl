package duration

import "testing"

func TestMinutes(t *testing.T) {
	readable := map[string]int{
		"1:30": 90, ":45": 45, " 45 ": 45, "1": 1, "45m": 45, "10 min": 10, "30 minutes": 30,
		"1.5h": 90, "2H": 120, "2 hrs": 120, "1 hour": 60, "1h30m": 90, "1h 30m": 90, "1 hour 20 min": 80,
	}
	for text, want := range readable {
		if got, ok := Minutes(text); !ok || got != want {
			t.Errorf("Minutes(%q) = %d, %v; want %d", text, got, ok, want)
		}
	}
	for _, text := range []string{"", "0:00", "0m", "1:75", "h", "soon", "30m 1h", "-5"} {
		if got, ok := Minutes(text); ok {
			t.Errorf("Minutes(%q) = %d; want unreadable", text, got)
		}
	}
}

func TestFormat(t *testing.T) {
	if got := Format(90); got != "1:30" {
		t.Errorf("Format(90) = %q", got)
	}
}
