// Package git contains wrappers around some Git commands.
package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// DiffPager calls `git diff no-index` on the two directory trees rooted at a
// and b and shows the result on stdout, possibly using a pager.
func DiffPager(a, b string) error {
	cmd := exec.Command("git", "diff", "--no-index", a, b)
	cmd.Stdout = os.Stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 1 {
					return nil
				}
				// ignore SIGPIPE
				if status.Signaled() && status.Signal() == syscall.SIGPIPE {
					return nil
				}
			}
			return fmt.Errorf("%s: %s", exiterr, strings.TrimSpace(stderr.String()))
		}
		return err
	}
	return nil
}
