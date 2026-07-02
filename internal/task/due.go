package task

import (
	"strings"
	"time"
)

// ParseDue parses natural language and ISO date strings into a time.Time.
// Supported: today, tomorrow, next week, day names (monday…sunday), YYYY-MM-DD.
// Returns nil if the input is empty or unrecognised.
func ParseDue(s string) *time.Time {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return nil
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, now.Location())

	switch s {
	case "today":
		return &today
	case "tomorrow":
		t := today.AddDate(0, 0, 1)
		return &t
	case "next week":
		t := today.AddDate(0, 0, 7)
		return &t
	}

	// day names
	days := map[string]time.Weekday{
		"sunday": time.Sunday, "monday": time.Monday, "tuesday": time.Tuesday,
		"wednesday": time.Wednesday, "thursday": time.Thursday,
		"friday": time.Friday, "saturday": time.Saturday,
	}
	if wd, ok := days[s]; ok {
		t := nextWeekday(today, wd)
		return &t
	}

	// ISO date YYYY-MM-DD
	if t, err := time.ParseInLocation("2006-01-02", s, now.Location()); err == nil {
		t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 0, 0, now.Location())
		return &t
	}

	return nil
}

func nextWeekday(from time.Time, wd time.Weekday) time.Time {
	diff := int(wd) - int(from.Weekday())
	if diff <= 0 {
		diff += 7
	}
	return from.AddDate(0, 0, diff)
}

// FormatDue returns a human-readable due date label.
func FormatDue(t *Task) string {
	if t.DueDate == nil {
		return ""
	}
	now := time.Now()
	diff := t.DueDate.Sub(now)
	switch {
	case diff < 0:
		return "overdue"
	case diff < 24*time.Hour:
		return "today"
	case diff < 48*time.Hour:
		return "tomorrow"
	case diff < 7*24*time.Hour:
		return t.DueDate.Format("Mon")
	default:
		return t.DueDate.Format("Jan 2")
	}
}
