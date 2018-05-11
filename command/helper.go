package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/keyfile"
	"github.com/frankbraun/codechain/util/bzero"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/home"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/terminal"
	"golang.org/x/crypto/ed25519"
)

func codechainHomeDir() string {
	if homeDir := os.Getenv("CODECHAINHOMEDIR"); homeDir != "" {
		log.Printf("$CODECHAINHOMEDIR=%s", homeDir)
		return homeDir
	}
	homeDir := home.AppDataDir("codechain", false)
	log.Printf("homeDir: %s", homeDir)
	return homeDir
}

func seckeyCheck(seckey string) error {
	if seckey != "" {
		exists, err := file.Exists(seckey)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("file '%s' doesn't exists", seckey)
		}
	} else {
		homeDir := codechainHomeDir()
		homeDir = filepath.Join(homeDir, secretsDir)
		// make sure we have the secrets directory present
		exists, err := file.Exists(homeDir)
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

func seckeyRead(filename string) (*[64]byte, *[64]byte, []byte, error) {
	exists, err := file.Exists(filename)
	if err != nil {
		return nil, nil, nil, err
	}
	if !exists {
		return nil, nil, nil, fmt.Errorf("file '%s' does not exist", filename)
	}
	fmt.Printf("opening keyfile: %s\n", filename)
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

func seckeyLoad(c *hashchain.HashChain, filename string) (*[64]byte, *[64]byte, []byte, error) {
	if filename != "" {
		return seckeyRead(filename)
	}
	homeDir := home.AppDataDir("codechain", false)
	homeDir = filepath.Join(homeDir, secretsDir)
	signer := c.Signer()

	files, err := ioutil.ReadDir(homeDir)
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
			fmt.Errorf("directory '%s' doesn' contain any matching secret keyfile", homeDir)
	}
	return seckeyRead(filepath.Join(homeDir, pubKey))
}
