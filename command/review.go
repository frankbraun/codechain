package command

import (
	"errors"
	"flag"
	"fmt"
)

// Review implements the 'review' command.
func Review(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [-s seckey.bin]\n", argv0)
		fmt.Fprintf(fs.Output(), "Review code changes ready for publication and changes of signers (or sig. ctl.).\n")
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
