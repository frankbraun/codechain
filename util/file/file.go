// Package file implements file related utility functions.
package file

import (
	"fmt"
	"io"
	"io/ioutil"
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
// The source file must exist already and be a regular file.
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
	// make sure source file is a regular file
	if !fi.Mode().IsRegular() {
		return fmt.Errorf("source file '%s' is not a regular file", src)
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

// CopyDir recursivly copies the source directory src to destination directory
// dst. The source directory must exist already and only contain regular files
// and directories. The destination directory must not exist already.
func CopyDir(src, dst string) error {
	if dst == "." {
		dst = filepath.Base(src)
	}
	if dst != "." {
		// make sure destination directory does not exist already
		exists, err := Exists(dst)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("destination directory '%s' exists already", dst)
		}
	}
	// make sure source directory exists and is a directory
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("source directory '%s' is not a directory", src)
	}
	// make destination directory
	mode := fi.Mode() & os.ModePerm // only keep standard UNIX permission bits
	if err := os.MkdirAll(dst, mode); err != nil {
		return err
	}
	// process source directory
	fis, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		s := filepath.Join(src, fi.Name())
		d := filepath.Join(dst, fi.Name())
		if fi.IsDir() {
			// recursion
			if err := CopyDir(s, d); err != nil {
				return err
			}
		} else {
			if err := Copy(s, d); err != nil {
				return err
			}
		}
	}
	return nil
}
