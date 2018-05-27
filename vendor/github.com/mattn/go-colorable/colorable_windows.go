// +build windows
// +build !appengine

package colorable

import (
	"io"
	"os"
)

// NewColorableStdout return new instance of Writer which handle escape sequence for stdout.
func NewColorableStdout() io.Writer {
	return NewColorable(os.Stdout)
}

// NewColorableStderr return new instance of Writer which handle escape sequence for stderr.
func NewColorableStderr() io.Writer {
	return NewColorable(os.Stderr)
}
