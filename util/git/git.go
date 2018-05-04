// Package git contains wrappers around some Git commands.
package git

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/log"
)

func diff(a, b string, capture bool) ([]byte, error) {
	var buf bytes.Buffer
	cmd := exec.Command("git", "diff", "--no-index", a, b)
	if capture {
		cmd.Stdout = &buf
	} else {
		cmd.Stdout = os.Stdout
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 1 {
					return buf.Bytes(), nil
				} else if !capture {
					// ignore SIGPIPE for non-caputuring output
					if status.Signaled() && status.Signal() == syscall.SIGPIPE {
						return nil, nil
					}
				}
			}
			return nil, fmt.Errorf("%s: %s", exiterr, strings.TrimSpace(stderr.String()))
		}
		return nil, err
	}
	return buf.Bytes(), nil
}

// Diff calls `git diff --no-index` on the two directory trees rooted at a and
// b and returns the resulting patch.
func Diff(a, b string) ([]byte, error) {
	return diff(a, b, true)
}

// DiffPager calls `git diff no-index` on the two directory trees rooted at a
// and b and shows the result on stdout, possibly using a pager.
func DiffPager(a, b string) error {
	_, err := diff(a, b, false)
	return err
}

// Apply calls `git apply` with the given patch in directory dir.
// Set p > 1 to remove more than 1 leading slashes from traditional diff paths.
// Use reverse to enable option -R.
func Apply(patch io.Reader, p int, dir string, dirOpt, reverse bool) error {
	args := []string{"apply"}
	if dirOpt && dir != "." {
		args = append(args, "--directory", dir)
	}
	if p > 1 {
		args = append(args, "-p", strconv.Itoa(p))
	}
	if reverse {
		args = append(args, "-R")
	}
	log.Printf("git " + strings.Join(args, " "))
	cmd := exec.Command("git", args...)
	// TODO: check for "."?
	cmd.Dir = dir
	cmd.Stdin = patch
	cmd.Stdout = os.Stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("%s: %s", exiterr, strings.TrimSpace(stderr.String()))
		}
		return err
	}
	return nil
}

// EnsureRootGitDir ensures that dir is either the root of a Git repository or
// not in a Git repository at all.
func EnsureRootGitDir(dir string) error {
	exists, err := file.Exists(filepath.Join(dir, ".git"))
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	d, _ := filepath.Split(strings.TrimSuffix(dir, string(filepath.Separator)))
	for d != "" {
		exists, err := file.Exists(filepath.Join(d, ".git"))
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("'%s' is not root of Git repo '%s'",
				dir, filepath.Join(d, ".git"))
		}
		d, _ = filepath.Split(strings.TrimSuffix(d, string(filepath.Separator)))
	}
	return nil
}
