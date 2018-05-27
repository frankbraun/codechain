// +build !windows
// +build !appengine

package colorable

import (
	"io"
	"os"

	_ "github.com/mattn/go-isatty"
)

// NewColorableStdout return new instance of Writer which handle escape sequence for stdout.
func NewColorableStdout() io.Writer {
	return os.Stdout
}

// NewColorableStderr return new instance of Writer which handle escape sequence for stderr.
func NewColorableStderr() io.Writer {
	return os.Stderr
}
