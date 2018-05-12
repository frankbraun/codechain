// Package file implements file related utility functions.
package file

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// IsBinary returns true if filename is binary and false otherwise.
// A binary file is defined as a file that doesn't consist entirely of valid
// UTF-8-encoded runes
func IsBinary(filename string) (bool, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return false, err
	}
	return !utf8.Valid(b), nil
}

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

func copyDir(src, dst string, excludePaths []string) error {
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
	files, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}
outer:
	for _, fi := range files {
		s := filepath.Join(src, fi.Name())
		d := filepath.Join(dst, fi.Name())
		if excludePaths != nil {
			canonical := s
			if src != "." {
				canonical = strings.TrimPrefix(s, src)
				canonical = strings.TrimPrefix(canonical, string(filepath.Separator))
			}
			canonical = filepath.ToSlash(canonical)
			for _, excludePath := range excludePaths {
				if excludePath == canonical {
					continue outer
				}
			}
		}
		if fi.IsDir() {
			// recursion
			if err := copyDir(s, d, nil); err != nil {
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

// CopyDir recursively copies the source directory src to destination directory
// dst. The source directory must exist already and only contain regular files
// and directories. The destination directory must not exist already.
func CopyDir(src, dst string) error {
	return copyDir(src, dst, nil)
}

// CopyDirExclude recursively copies the source directory src to destination
// directory dst, except for paths contained in excludePath. The source
// directory must exist already and only contain regular files and
// directories. The destination directory must not exist already.
func CopyDirExclude(src, dst string, excludePaths []string) error {
	return copyDir(src, dst, excludePaths)
}

// RemoveAll removes all files and directories in path except the ones given
// in excludePaths.
func RemoveAll(path string, excludePaths []string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
outer:
	for _, fi := range files {
		if excludePaths != nil {
			canonical := filepath.ToSlash(fi.Name())
			for _, excludePath := range excludePaths {
				if excludePath == canonical {
					continue outer
				}
			}
		}
		if err := os.RemoveAll(filepath.Join(path, fi.Name())); err != nil {
			return err
		}
	}
	return nil
}
