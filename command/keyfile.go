package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/internal/base64"
)

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
	sec, sig, comment, err := seckeyRead(*seckey)
	if err != nil {
		return err
	}
	fmt.Println("public key with signature and optional comment")
	fmt.Printf("%s %s", base64.Encode(sec[32:]), base64.Encode(sig[:]))
	if len(comment) > 0 {
		fmt.Printf(" '%s'", string(comment))
	}
	fmt.Println("")
	return nil
}
