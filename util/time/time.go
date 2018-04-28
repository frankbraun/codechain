// Package time implements time related utility functions.
package time

import (
	"time"
)

// Now returns the current time in UTC as Unix time,
// the number of seconds elapsed since January 1, 1970 UTC.
func Now() int64 {
	return time.Now().UTC().Unix()
}

// Parse the time value as an RFC3339 string and return it as Unix time.
func Parse(value string) (int64, error) {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return 0, err
	}
	return t.UTC().Unix(), nil
}

// Format the given datum as an RFC3339 string in UTC.
func Format(datum int64) string {
	return time.Unix(datum, 0).UTC().Format(time.RFC3339)
}
