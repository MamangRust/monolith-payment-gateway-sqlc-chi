package adapter

import "time"

// parseTime parses an RFC3339 string and returns the corresponding time.Time.
// Returns a zero time on empty input or parse failure.
func parseTime(ts string) time.Time {
	if ts == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Time{}
	}
	return t
}

// parseTimePtr is like parseTime but returns a pointer; nil for empty/invalid input.
func parseTimePtr(ts string) *time.Time {
	t := parseTime(ts)
	if t.IsZero() {
		return nil
	}
	return &t
}
