package command

import (
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/frankbraun/codechain/internal/base64"
	"github.com/frankbraun/codechain/keyfile"
	"github.com/frankbraun/codechain/util/bzero"
	"github.com/frankbraun/codechain/util/terminal"
)

func changePassphrase(filename string, sec, sig *[64]byte, comment []byte) error {
	pass, err := terminal.ReadPassphrase(syscall.Stdin, true)
	if err != nil {
		return err
	}
	defer bzero.Bytes(pass)
	tmpfile := filename + ".new"
	os.Remove(tmpfile) // ignore error
	// create new keyfile
	if err := keyfile.Create(tmpfile, pass, *sec, *sig, comment); err != nil {
		return err
	}
	// move temp. file in place
	return os.Rename(tmpfile, filename)
}

// KeyFile implements the 'keyfile' command.
func KeyFile(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -s seckey.bin\n", argv0)
		fmt.Fprintf(os.Stderr, "Show pubkey, signature, and comment for encrypted secret key file.\n")
		fs.PrintDefaults()
	}
	change := fs.Bool("c", false, "Change passphrase")
	seckey := fs.String("s", "", "Secret key file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *seckey == "" {
		return fmt.Errorf("%s: option -s is mandatory", argv0)
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	sec, sig, comment, err := seckeyRead(*seckey)
	if err != nil {
		return err
	}
	if *change {
		fmt.Printf("%s read, please provide new ", *seckey)
		if err := changePassphrase(*seckey, sec, sig, comment); err != nil {
			return err
		}
		fmt.Println("passphrase changed")
	} else {
		fmt.Println("public key with signature and optional comment")
		fmt.Printf("%s %s", base64.Encode(sec[32:]), base64.Encode(sig[:]))
		if len(comment) > 0 {
			fmt.Printf(" '%s'", string(comment))
		}
		fmt.Println("")
	}
	return nil
}
