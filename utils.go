package main

import (
	"strings"
	"time"
)

// Now returns the current time, it's a helper function for easier testing.
func Now() time.Time {
	return time.Now()
}

// NowString returns the current time in string format.
func NowString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// Trim spaces etc from the start and end of a string.
func Trim(s string) string {
	return strings.TrimSpace(s)
}

// SplitCSV splits a comma-separated string into a slice of strings, trimming spaces and removes empty items.
func SplitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
