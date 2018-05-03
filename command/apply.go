package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/tree"
)

func apply(c *hashchain.HashChain, verbose bool) error {
	targetHash, _ := c.LastSignedTreeHash()
	treeHashes := c.TreeHashes()
	err := tree.Sync(".", targetHash, patchDir, treeHashes, verbose, excludePaths, false)
	if err != nil {
		return err
	}

	return nil
}

// Apply implements the 'apply' command.
func Apply(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", argv0)
		fmt.Fprintf(os.Stderr, "Apply all patches with enough signatures to code tree.\n")
		fs.PrintDefaults()
	}
	verbose := fs.Bool("v", false, "Be verbose")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	c, err := hashchain.Read(hashchainFile)
	if err != nil {
		return err
	}
	if err := c.Close(); err != nil {
		return err
	}
	return apply(c, *verbose)
}
