package ssot

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/frankbraun/codechain/util/base64"
	utime "github.com/frankbraun/codechain/util/time"
)

func (sh *SignedHead) calculateRotateTime(validity time.Duration) int64 {
	now := utime.Now()
	rest := now - sh.validTo
	if rest < 0 {
		rest = 0
	}
	rotateIn := int64(validity/time.Second) - rest
	rotateIn /= 3
	if rotateIn < 0 {
		rotateIn = 0
	}
	return now + rotateIn
}

// WriteRotateTo writes "rotate to" file to given filename.
func (sh *SignedHead) WriteRotateTo(
	filename string,
	secKeyRotate *[64]byte,
	sigRotate *[64]byte,
	commentRotate []byte,
	validity time.Duration,
) error {
	rotateTime := sh.calculateRotateTime(validity)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s %s", base64.Encode(secKeyRotate[32:]),
		base64.Encode(sigRotate[:]))
	if err != nil {
		return err
	}
	if commentRotate != nil {
		_, err := fmt.Fprintf(f, " %s", commentRotate)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(f, "\n%s\n", utime.Format(rotateTime))
	if err != nil {
		return err
	}
	return nil
}

// ReadRotateTo reads "rotate to" file from given filename and returns the
// public key to rotate to and a bool indicating if the rotation time has been
// reached.
func ReadRotateTo(filename string) (string, bool, error) {
	c, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", false, err
	}
	lines := bytes.SplitN(c, []byte("\n"), 2)
	line0 := strings.SplitN(string(lines[0]), " ", 3)
	line1 := string(bytes.TrimSpace(lines[1]))
	if _, err := base64.Decode(line0[0], 32); err != nil {
		return "", false, err
	}
	rotateTo := line0[0]
	rotateTime, err := utime.Parse(line1)
	if err != nil {
		return "", false, err
	}
	var reached bool
	if rotateTime <= utime.Now() {
		reached = true
	}
	return rotateTo, reached, nil
}
