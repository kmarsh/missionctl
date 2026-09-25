// Package dates handles the API's YYYY-MM-DD dates in the local time zone.
package dates

import (
	"fmt"
	"strings"
	"time"
)

const layout = "2006-01-02"

// Resolve turns "today", "yesterday", or a YYYY-MM-DD date into YYYY-MM-DD.
func Resolve(text string, now time.Time) (string, error) {
	switch strings.ToLower(text) {
	case "today":
		return now.Format(layout), nil
	case "yesterday":
		return now.AddDate(0, 0, -1).Format(layout), nil
	}
	if date, err := time.Parse(layout, text); err == nil && date.Format(layout) == text {
		return text, nil
	}
	return "", fmt.Errorf("invalid date %q; use YYYY-MM-DD, today, or yesterday", text)
}

// Humanize gives "Thu, Sep 24" for an API date, with the year when it isn't now's.
func Humanize(text string, now time.Time) string {
	date, err := time.Parse(layout, text)
	if err != nil {
		return text
	}
	if date.Year() != now.Year() {
		return date.Format("Mon, Jan 2, 2006")
	}
	return date.Format("Mon, Jan 2")
}

// StartOfWeek is the Monday of the week containing now.
func StartOfWeek(now time.Time) string {
	daysSinceMonday := (int(now.Weekday()) + 6) % 7
	return now.AddDate(0, 0, -daysSinceMonday).Format(layout)
}
