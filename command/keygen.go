package command

import (
	"crypto/ed25519"
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/frankbraun/codechain/keyfile"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/util/base64"
	"github.com/frankbraun/codechain/util/bzero"
	"github.com/frankbraun/codechain/util/def"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/seckey"
	"github.com/frankbraun/codechain/util/terminal"
)

// TestComment is a comment used for testing purposes. Do not set!
var TestComment string

// KeyGen implements the 'keygen' command.
func KeyGen(checkUpToDate, homeDir, argv0 string, args ...string) error {
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
	secKey := fs.String("s", "", "Secret key file")
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
	if err := secpkg.UpToDate(checkUpToDate); err != nil {
		return err
	}
	if *secKey != "" {
		exists, err := file.Exists(*secKey)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("file '%s' exists already", *secKey)
		}
	} else {
		secretsDir = filepath.Join(homeDir, def.SecretsSubDir)
		if err := os.MkdirAll(secretsDir, 0700); err != nil {
			return err
		}
	}
	if seckey.TestPass == "" {
		pass, err = terminal.ReadPassphrase(syscall.Stdin, true)
		if err != nil {
			return err
		}
		defer bzero.Bytes(pass)
	} else {
		pass = []byte(seckey.TestPass)
	}
	if TestComment == "" {
		fmt.Println("comment (e.g., John Doe <john@example.com>; can be empty; cannot be changed):")
		comment, err = terminal.ReadLine(os.Stdin)
		if err != nil {
			return err
		}
	} else {
		comment = []byte(TestComment)
	}
	pub, sec, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	sig := ed25519.Sign(sec, append(pub, comment...))
	pubEnc := base64.Encode(pub[:])
	var sk [64]byte
	copy(sk[:], sec)
	var signature [64]byte
	copy(signature[:], sig)
	if *secKey != "" {
		err := keyfile.Create(*secKey, pass, sk, signature, comment)
		if err != nil {
			return err
		}
	} else {
		filename := filepath.Join(secretsDir, pubEnc)
		err := keyfile.Create(filename, pass, sk, signature, comment)
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
