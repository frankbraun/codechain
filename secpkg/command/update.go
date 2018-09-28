package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

// Update implements the secpkg 'update' command.
func Update(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s project_name\n", argv0)
		fmt.Fprintf(os.Stderr, "Update installed package with given project_name, if necessary.\n")
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
