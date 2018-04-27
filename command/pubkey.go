package command

import (
	"encoding/base64"
	"flag"
	"fmt"
	"syscall"

	"github.com/frankbraun/codechain/keyfile"
	"github.com/frankbraun/codechain/util/bzero"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/terminal"
	"golang.org/x/crypto/ed25519"
)

// PubKey implements the 'pubkey' command.
func PubKey(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s -s seckey.bin\n", argv0)
		fs.PrintDefaults()
	}
	change := fs.Bool("c", false, "Change passphrase")
	seckey := fs.String("s", "", "Secret key file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *change {
		// TODO
		return fmt.Errorf("%s: option -c not implemented yet", argv0)
	}
	if *seckey == "" {
		return fmt.Errorf("%s: option -s is mandatory", argv0)
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	exists, err := file.Exists(*seckey)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%s: file '%s' does not exist", argv0, *seckey)
	}
	var pass []byte
	if testPass == "" {
		pass, err = terminal.ReadPassphrase(syscall.Stdin, false)
		if err != nil {
			return err
		}
		defer bzero.Bytes(pass)
	} else {
		pass = []byte(testPass)
	}
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
