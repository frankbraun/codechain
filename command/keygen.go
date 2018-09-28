package command

import (
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/frankbraun/codechain/internal/base64"
	"github.com/frankbraun/codechain/keyfile"
	"github.com/frankbraun/codechain/util/bzero"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/terminal"
	"golang.org/x/crypto/ed25519"
)

const secretsSubDir = "secrets"

var (
	testPass    string
	testComment string
)

// KeyGen implements the 'keygen' command.
func KeyGen(homeDir, argv0 string, args ...string) error {
	var (
		secretsDir string
		pass       []byte
		comment    []byte
		err        error
	)
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-s seckey.bin]\n", argv0)
		fmt.Fprintf(os.Stderr, "Generate new encrypted secret key file and show pubkey, signature, and comment.\n")
		fs.PrintDefaults()
	}
	seckey := fs.String("s", "", "Secret key file")
	verbose := fs.Bool("v", false, "Be verbose")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	if *seckey != "" {
		exists, err := file.Exists(*seckey)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("file '%s' exists already", *seckey)
		}
	} else {
		secretsDir = filepath.Join(homeDir, secretsSubDir)
		if err := os.MkdirAll(secretsDir, 0700); err != nil {
			return err
		}
	}
	if testPass == "" {
		pass, err = terminal.ReadPassphrase(syscall.Stdin, true)
		if err != nil {
			return err
		}
		defer bzero.Bytes(pass)
	} else {
		pass = []byte(testPass)
	}
	if testComment == "" {
		fmt.Println("comment (e.g., name; can be empty):")
		comment, err = terminal.ReadLine(os.Stdin)
		if err != nil {
			return err
		}
	} else {
		comment = []byte(testComment)
	}
	pub, sec, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	sig := ed25519.Sign(sec, append(pub, comment...))
	pubEnc := base64.Encode(pub[:])
	var secKey [64]byte
	copy(secKey[:], sec)
	var signature [64]byte
	copy(signature[:], sig)
	if *seckey != "" {
		err := keyfile.Create(*seckey, pass, secKey, signature, comment)
		if err != nil {
			return err
		}
	} else {
		filename := filepath.Join(secretsDir, pubEnc)
		err := keyfile.Create(filename, pass, secKey, signature, comment)
		if err != nil {
			return err
		}
		fmt.Println("secret key file created:")
		fmt.Println(filename)
	}
	fmt.Println("public key with signature and optional comment:")
	fmt.Printf("%s %s", pubEnc, base64.Encode(sig))
	if len(comment) > 0 {
		fmt.Printf(" '%s'", string(comment))
	}
	fmt.Println("")
	return nil
}
