package command

import (
	"flag"
	"fmt"

	"github.com/frankbraun/codechain/tree"
)

// TreeHash implements the 'treehash' command.
func TreeHash(argv0 string, args ...string) error {
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
	hash, err := tree.Hash(".", excludePaths)
	if err != nil {
		return err
	}
	fmt.Printf("%x\n", hash[:])
	return nil
}
