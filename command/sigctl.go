package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

// SigCtl implements the 'sigctl' command.
func SigCtl(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -m\n", argv0)
		fmt.Fprintf(os.Stderr, "Change signature control value.\n")
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
