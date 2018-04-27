package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

// Review implements the 'review' command.
func Review(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-s seckey.bin] [treehash]\n", argv0)
		fmt.Fprintf(os.Stderr, "Review code changes (all or up to treehash) and changes of signers and sigctl.\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	// TODO
	return errors.New("not implemented")
}
