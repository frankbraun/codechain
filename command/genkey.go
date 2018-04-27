package command

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/frankbraun/codechain/keyfile"
	"github.com/frankbraun/codechain/util/bzero"
	"github.com/frankbraun/codechain/util/home"
	"github.com/frankbraun/codechain/util/terminal"
	"golang.org/x/crypto/ed25519"
)

const secretsDir = "secrets"

// GenKey implements the 'genkey' command.
func GenKey() error {
	var homeDir string
	app := os.Args[1]
	fs := flag.NewFlagSet(os.Args[0]+" "+app, flag.ExitOnError)
	seckey := fs.String("s", "", "Secret key file")
	fs.Parse(os.Args[2:])
	if *seckey == "" {
		homeDir = home.AppDataDir("codechain", false)
		homeDir = filepath.Join(homeDir, secretsDir)
		if err := os.MkdirAll(homeDir, 0700); err != nil {
			return err
		}
	}
	pass, err := terminal.ReadPassphrase(syscall.Stdin, true)
	if err != nil {
		return err
	}
	defer bzero.Bytes(pass)
	fmt.Println("comment (e.g., name; can be empty):")
	comment, err := terminal.ReadLine(os.Stdin)
	if err != nil {
		return err
	}
	pub, sec, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	sig := ed25519.Sign(sec, append(pub, comment...))
	pubEnc := base64.URLEncoding.EncodeToString(pub[:])
	if *seckey != "" {
		if err := keyfile.Create(*seckey, pass, sec, sig, comment); err != nil {
			return err
		}
	} else {
		filename := filepath.Join(homeDir, pubEnc)
		if err := keyfile.Create(filename, pass, sec, sig, comment); err != nil {
			return err
		}
		fmt.Println("secret key file created:")
		fmt.Println(filename)
	}
	fmt.Println("public key with signature and optional comment")
	fmt.Printf("%s %s", pubEnc,
		base64.URLEncoding.EncodeToString(sig))
	if len(comment) > 0 {
		fmt.Printf(" %s", string(comment))
	}
	fmt.Println("")
	return nil
}
