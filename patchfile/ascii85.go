package patchfile

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/frankbraun/codechain/internal/ascii85"
)

// ascii85Diff encodes the file with filename in ascii85 and writes it to w as
// an "ascii85" section.
func ascii85Diff(w io.Writer, filename string) error {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	buf, lines, err := ascii85.Encode(src)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "ascii85 %d\n", lines)
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}

// ascii85Apply decodes the ascii85 encoding in patch and writes it to w.
// patch must not include the "ascii85" section header.
func ascii85Apply(w io.Writer, _ string, patch []byte) error {
	buf, err := ascii85.Decode(patch)
	if err != nil {
		return err
	}
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}
