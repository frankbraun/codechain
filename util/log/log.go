// Package log implements a minimal logging framework based on stdlib's log.
package log

import (
	"fmt"
	"io"
	"log"
)

// Std is the standard logger. The default is nil (nothing is logged).
var Std *log.Logger

// NewStd returns a new logger with standard flags (log.LstdFlags) and no
// prefix.
func NewStd(out io.Writer) *log.Logger {
	return log.New(out, "", log.LstdFlags)
}

// Printf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...interface{}) {
	if Std == nil {
		return
	}
	Std.Output(2, fmt.Sprintf(format, v...))
}

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Println(v ...interface{}) {
	if Std == nil {
		return
	}
	Std.Output(2, fmt.Sprintln(v...))
}
