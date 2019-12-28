// Package seckey implements helper functions for secret key files.
package seckey

import (
	"crypto/ed25519"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"syscall"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/keyfile"
	"github.com/frankbraun/codechain/util/bzero"
	"github.com/frankbraun/codechain/util/def"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/terminal"
)

// TestPass is a passphrase used for testing purposes.
var TestPass string

// Check that the file seckey exists, if it is given.
// Otherwise make sure that at least the secrets subdirectory of homeDir
// exists.
func Check(homeDir, seckey string) error {
	if seckey != "" {
		exists, err := file.Exists(seckey)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("file '%s' doesn't exists", seckey)
		}
	} else {
		secretDir := filepath.Join(homeDir, def.SecretsSubDir)
		// make sure we have the secrets directory present
		exists, err := file.Exists(secretDir)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("directory '%s' doesn't exists: you have no secrets",
				homeDir)
		}
	}
	return nil
}

// Read reads the secret key from given filename.
// It reads the the passphrase from the terminal. If the wrong passphrase is
// given, the function reads the passphrase again.
func Read(filename string) (*[64]byte, *[64]byte, []byte, error) {
	exists, err := file.Exists(filename)
	if err != nil {
		return nil, nil, nil, err
	}
	if !exists {
		return nil, nil, nil, fmt.Errorf("keyfile '%s' does not exist", filename)
	}
	fmt.Printf("opening keyfile: %s\n", filename)
	var (
		pass    []byte
		sec     *[64]byte
		sig     *[64]byte
		comment []byte
	)
	for {
		if TestPass == "" {
			pass, err = terminal.ReadPassphrase(syscall.Stdin, false)
			if err != nil {
				return nil, nil, nil, err
			}
			defer bzero.Bytes(pass)
		} else {
			pass = []byte(TestPass)
		}
		sec, sig, comment, err = keyfile.Read(filename, pass)
		if err != nil {
			if TestPass == "" && err == keyfile.ErrDecrypt {
				fmt.Println("wrong passphrase, try again")
				continue
			}
			return nil, nil, nil, err
		}
		break
	}
	if !ed25519.Verify(sec[32:], append(sec[32:], comment...), sig[:]) {
		return nil, nil, nil, fmt.Errorf("signature does not verify")
	}
	return sec, sig, comment, nil
}

// Load loads secret from filename, if given.
// Otherwise it loads the secret corresponding to the signer in given hash
// chain and makes sure that only one such secret exists.
func Load(c *hashchain.HashChain, homeDir, filename string) (*[64]byte, *[64]byte, []byte, error) {
	if filename != "" {
		return Read(filename)
	}
	secretDir := filepath.Join(homeDir, def.SecretsSubDir)
	signer := c.Signer()
	files, err := ioutil.ReadDir(secretDir)
	if err != nil {
		return nil, nil, nil, err
	}
	var pubKey string
	for _, fi := range files {
		if signer[fi.Name()] {
			if pubKey == "" {
				pubKey = fi.Name()
			} else {
				return nil, nil, nil,
					fmt.Errorf("more than one matching keyfile found: you have too many secrets")
			}
		}
	}
	if pubKey == "" {
		return nil, nil, nil,
			fmt.Errorf("directory '%s' doesn't contain any matching secret keyfile", secretDir)
	}
	return Read(filepath.Join(secretDir, pubKey))
}
