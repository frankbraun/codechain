package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

// RemKey implements the 'remkey' command.
func RemKey(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s pubkey\n", argv0)
		fmt.Fprintf(os.Stderr, "Remove existing signer from hashchain.\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return flag.ErrHelp
	}
	// TODO
	return errors.New("not implemented")
}
