// Package file implements file related utility functions.
package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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

// Copy source file src to destination file dst.
// The destination file must not exist already.
func Copy(src, dst string) error {
	if dst == "." {
		dst = filepath.Base(src)
	}
	if dst != "." {
		// make sure destination file does not exist already
		exists, err := Exists(dst)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("destination file '%s' exists already", dst)
		}
	}
	// open source file
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	// get mode of source file
	fi, err := s.Stat()
	if err != nil {
		return err
	}
	mode := fi.Mode() & os.ModePerm // only keep standard UNIX permission bits
	// create destination file
	d, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer d.Close()
	// copy content
	if _, err := io.Copy(d, s); err != nil {
		return err
	}
	return nil
}
