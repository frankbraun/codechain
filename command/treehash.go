package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/tree"
)

// TreeHash implements the 'treehash' command.
func TreeHash(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", argv0)
		fmt.Fprintf(os.Stderr, "Show tree hash or tree list of current directory.\n")
		fs.PrintDefaults()
	}
	list := fs.Bool("l", false, "Print tree list instead of hash")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	if *list {
		l, err := tree.ListBytes(".", ExcludePaths)
		if err != nil {
			return err
		}
		os.Stdout.Write(l)
	} else {
		hash, err := tree.Hash(".", ExcludePaths)
		if err != nil {
			return err
		}
		fmt.Printf("%x\n", hash[:])
	}
	return nil
}
