package command

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/frankbraun/codechain/keyfile"
	"github.com/frankbraun/codechain/util/bzero"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/terminal"
	"golang.org/x/crypto/ed25519"
)

// PubKey implements the 'pubkey' command.
func PubKey() error {
	app := os.Args[1]
	fs := flag.NewFlagSet(os.Args[0]+" "+app, flag.ContinueOnError)
	seckey := fs.String("s", "", "Secret key file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}
	if *seckey == "" {
		return fmt.Errorf("%s: option -s is mandatory", app)
	}
	exists, err := file.Exists(*seckey)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%s: file '%s' does not exist", app, *seckey)
	}
	pass, err := terminal.ReadPassphrase(syscall.Stdin, false)
	if err != nil {
		return err
	}
	defer bzero.Bytes(pass)
	sec, sig, comment, err := keyfile.Read(*seckey, pass)
	if err != nil {
		return err
	}
	if !ed25519.Verify(sec[32:], append(sec[32:], comment...), sig) {
		return fmt.Errorf("signature does not verify")
	}
	fmt.Println("public key with signature and optional comment")
	fmt.Printf("%s %s",
		base64.URLEncoding.EncodeToString(sec[32:]),
		base64.URLEncoding.EncodeToString(sig))
	if len(comment) > 0 {
		fmt.Printf(" %s", string(comment))
	}
	fmt.Println("")
	return nil
}
