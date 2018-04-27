package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/tree"
)

// TreeList implements the 'treelist' command.
func TreeList(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s\n", argv0)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	list, err := tree.List(".", excludePaths)
	if err != nil {
		return err
	}
	os.Stdout.Write(list)
	return nil
}
