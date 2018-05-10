package patchfile

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/frankbraun/codechain/internal/ascii85"
)

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
