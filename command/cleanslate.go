package command

import (
	"errors"
	"flag"
	"fmt"
)

// CleanSlate implements the 'cleanslate' command.
func CleanSlate(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s\n", argv0)
		fmt.Fprintf(fs.Output(), "Remove all files except .codechain, .git, and .gitignore from current directory.\n")
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
