package command

import (
	"fmt"
	"syscall"

	"github.com/frankbraun/codechain/keyfile"
	"github.com/frankbraun/codechain/util/bzero"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/terminal"
	"golang.org/x/crypto/ed25519"
)

func seckeyRead(filename string) (*[64]byte, *[64]byte, []byte, error) {
	exists, err := file.Exists(filename)
	if err != nil {
		return nil, nil, nil, err
	}
	if !exists {
		return nil, nil, nil, fmt.Errorf("file '%s' does not exist", filename)
	}
	var pass []byte
	if testPass == "" {
		pass, err = terminal.ReadPassphrase(syscall.Stdin, false)
		if err != nil {
			return nil, nil, nil, err
		}
		defer bzero.Bytes(pass)
	} else {
		pass = []byte(testPass)
	}
	sec, sig, comment, err := keyfile.Read(filename, pass)
	if err != nil {
		return nil, nil, nil, err
	}
	if !ed25519.Verify(sec[32:], append(sec[32:], comment...), sig[:]) {
		return nil, nil, nil, fmt.Errorf("signature does not verify")
	}
	return sec, sig, comment, nil
}
