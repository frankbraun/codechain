// Package ascii85 implements ascii85 encoding related utility functions.
package ascii85

import (
	"bytes"
	"encoding/ascii85"
	"io"
)

type newlineWriteCloser struct {
	buf         bytes.Buffer
	charsInLine int
	lines       int
}

func (n *newlineWriteCloser) Write(p []byte) (int, error) {
	startLen := len(p)
	for len(p)+n.charsInLine >= 80 {
		if _, err := n.buf.Write(p[:80-n.charsInLine]); err != nil {
			return 0, err
		}
		p = p[80-n.charsInLine:]
		if _, err := n.buf.WriteRune('\n'); err != nil {
			return 0, err
		}
		n.charsInLine = 0
		n.lines++
	}
	if len(p) > 0 {
		if _, err := n.buf.Write(p); err != nil {
			return 0, err
		}
		n.charsInLine = len(p)
	}
	return startLen, nil
}

func (n *newlineWriteCloser) Close() error {
	if n.charsInLine > 0 {
		if _, err := n.buf.WriteRune('\n'); err != nil {
			return err
		}
		n.charsInLine = 0
		n.lines++
	}
	return nil
}

// Encode src to ascii85 with a newline every 80 encoded characters and return
// the result and the number of encoded lines.
func Encode(src []byte) ([]byte, int, error) {
	var n newlineWriteCloser
	a := ascii85.NewEncoder(&n)
	if _, err := a.Write(src); err != nil {
		return nil, 0, err
	}
	if err := a.Close(); err != nil {
		return nil, 0, err
	}
	if err := n.Close(); err != nil {
		return nil, 0, err
	}
	return n.buf.Bytes(), n.lines, nil
}

// Decode ascii85 encoded src and return it.
func Decode(src []byte) ([]byte, error) {
	var dst bytes.Buffer
	dec := ascii85.NewDecoder(bytes.NewBuffer(src))
	if _, err := io.Copy(&dst, dec); err != nil {
		return nil, err
	}
	return dst.Bytes(), nil
}
