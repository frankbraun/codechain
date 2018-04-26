// Package file implements file related utility functions.
package file

import (
	"os"
)

// Exists checks if filename exists already.
func Exists(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, err
}
