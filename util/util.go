// Package util contains utility functions.
package util

import (
	"fmt"
	"os"
)

// ContainsString returns true, if the the string array sa contains the string s.
// Otherwise, it returns false.
func ContainsString(sa []string, s string) bool {
	for _, v := range sa {
		if v == s {
			return true
		}
	}
	return false
}

// Fatal prints err to stderr (prefixed with os.Args[0]) and exits the process
// with exit code 1.
func Fatal(err error) {
	fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], err)
	os.Exit(1)
}
