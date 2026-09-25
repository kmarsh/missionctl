// Package duration reads durations the way people type them.
package duration

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var units = regexp.MustCompile(`^(?:(\d+(?:\.\d+)?)\s*(?:h|hr|hrs|hour|hours))?\s*(?:(\d+)\s*(?:m|min|mins|minute|minutes))?$`)

// Minutes reads "1:30", ":45", a bare number of minutes ("45"), or hours and
// minutes with units: "45m", "10 min", "1.5h", "2 hrs", "1h30m", "1 hour 15 minutes".
// It reports false when the text is unreadable or zero.
func Minutes(text string) (int, bool) {
	text = strings.ToLower(strings.TrimSpace(text))
	var minutes int
	var ok bool
	if hours, rest, found := strings.Cut(text, ":"); found {
		minutes, ok = clock(hours, rest)
	} else if bare, err := strconv.Atoi(text); err == nil {
		minutes, ok = bare, true
	} else {
		minutes, ok = withUnits(text)
	}
	return minutes, ok && minutes > 0
}

// Format gives "1:30" for 90 minutes.
func Format(minutes int) string {
	return fmt.Sprintf("%d:%02d", minutes/60, minutes%60)
}

func clock(hoursText, minutesText string) (int, bool) {
	hours := 0
	if hoursText != "" {
		var err error
		if hours, err = strconv.Atoi(hoursText); err != nil || hours < 0 {
			return 0, false
		}
	}
	minutes, err := strconv.Atoi(minutesText)
	if err != nil || minutes < 0 || minutes >= 60 {
		return 0, false
	}
	return hours*60 + minutes, true
}

func withUnits(text string) (int, bool) {
	match := units.FindStringSubmatch(text)
	if text == "" || match == nil {
		return 0, false
	}
	hours, minutes := 0.0, 0
	if match[1] != "" {
		hours, _ = strconv.ParseFloat(match[1], 64)
	}
	if match[2] != "" {
		minutes, _ = strconv.Atoi(match[2])
	}
	return int(hours*60+0.5) + minutes, true
}
