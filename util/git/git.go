// Package git contains wrappers around some Git commands.
package git

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

// Diff calls `git diff --no-index` on the two directory trees rooted at a and
// be and returns the resulting patch.
func Diff(a, b string) (string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("git", "diff", "--no-index", a, b)
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 1 {
					return buf.String(), nil
				}
			}
			return "", fmt.Errorf("%s: %s", exiterr, string(exiterr.Stderr))
		}
		return "", err
	}
	return buf.String(), nil
}

// Apply calls `git apply` with the given patch in directory dir.
// Set p > 1 to remove more than 1 leading slashes from traditional diff paths.
// Use reverse to enable option -R.
func Apply(patch io.Reader, p int, dir string, reverse bool) error {
	args := []string{"apply"}
	if p > 1 {
		args = append(args, "-p", strconv.Itoa(p))
	}
	if reverse {
		args = append(args, "-R")
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdin = patch
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("%s: %s", exiterr, string(exiterr.Stderr))
		}
		return err
	}
	return nil
}
