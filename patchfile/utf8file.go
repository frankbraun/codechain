package patchfile

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

// utf8fileDiff writes the file with filename to w as an "utf8file" section.
func utf8fileDiff(w io.Writer, filename string) error {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "utf8file %d\n", bytes.Count(src, []byte("\n"))+1)
	if _, err := w.Write(src); err != nil {
		return err
	}
	// write newline, we cannot be sure file ends with one
	if _, err := io.WriteString(w, "\n"); err != nil {
		return err
	}
	return nil
}

// utf8fileApply takes utf8file encoded in patch and writes it to w.
// patch must not include the "utf8file" section header.
func utf8fileApply(w io.Writer, _ string, patch []byte) error {
	if _, err := w.Write(patch); err != nil {
		return err
	}
	return nil
}
