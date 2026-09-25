package cli

import (
	"testing"
)

// These fail before reaching the network, so they need no API key or server.
func TestExitCodes(t *testing.T) {
	t.Setenv("MISSIONCTL_API_KEY", "unused")
	for name, test := range map[string]struct {
		args []string
		want int
	}{
		"help":             {[]string{"--help"}, 0},
		"unknown command":  {[]string{"bogus"}, codeUsage},
		"unknown flag":     {[]string{"projects", "--bogus"}, codeUsage},
		"missing duration": {[]string{"time", "log", "Standup"}, codeUsage},
		"bad duration":     {[]string{"time", "log", "Standup", "-d", "soon"}, codeUsage},
		"bad date":         {[]string{"time", "log", "Standup", "-d", "15", "--date", "2026-13-01"}, codeUsage},
		"nothing to edit":  {[]string{"time", "edit", "some-id"}, codeUsage},
		"both list modes":  {[]string{"projects", "--all", "--disabled"}, codeUsage},
		"delete unasked":   {[]string{"time", "delete", "some-id"}, codeUsage},
	} {
		if got := Run("test", test.args); got != test.want {
			t.Errorf("%s: Run(%q) = %d; want %d", name, test.args, got, test.want)
		}
	}
}

func TestNoAPIKey(t *testing.T) {
	t.Setenv("MISSIONCTL_API_KEY", "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // where Windows looks
	if got := Run("test", []string{"projects"}); got != codeConfig {
		t.Errorf("Run without a key = %d; want %d", got, codeConfig)
	}
}

func TestSummary(t *testing.T) {
	if got := summary("First line\nsecond", 60); got != "First line" {
		t.Errorf("summary kept %q", got)
	}
	if got := summary("abcdef", 4); got != "abc…" {
		t.Errorf("summary cut to %q", got)
	}
}
