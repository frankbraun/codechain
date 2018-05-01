// Package lockfile implements a lock to limit a binary to one process per
// anchor file.
package lockfile

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

const lockSuffix = ".lock"

// Lock is lockfile to limit a binary to one process per anchor file.
type Lock string

// Create a lock for the given binary anchorFile.
// Returns an error if the lock already exists.
func Create(anchorFile string) (Lock, error) {
	filename := anchorFile + lockSuffix
	_, err := os.Stat(filename)
	if err == nil {
		// file exists
		pid, err := ioutil.ReadFile(filename)
		if err != nil {
			return "", fmt.Errorf("lockfile: %s: already exists: %v",
				filename, err)
		}
		return "", fmt.Errorf("lockfile: %s: already exists (PID %s)",
			filename, bytes.TrimSpace(pid))
	}
	pid := os.Getpid()
	fp, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer fp.Close()
	if _, err := io.WriteString(fp, strconv.Itoa(pid)+"\n"); err != nil {
		return "", err
	}
	return Lock(filename), nil
}

// Release the lock.
// The protected process should call this method during shutdown.
func (l *Lock) Release() error {
	s := string(*l)
	if s == "" {
		return nil
	}
	err := os.Remove(s)
	if err != nil {
		*l = ""
	}
	return err
}
