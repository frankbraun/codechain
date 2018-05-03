// Package util contains utility functions.
package util

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
