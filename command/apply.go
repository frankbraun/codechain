package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/archive"
	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/sync"
	"github.com/frankbraun/codechain/util/log"
)

func applyDist(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Printf("applying distribution '%s'", filename)
	return archive.Apply(def.HashchainFile, def.PatchDir, f)
}

func apply(c *hashchain.HashChain) error {
	targetHash, _ := c.LastSignedTreeHash()
	treeHashes := c.TreeHashes()
	err := sync.Dir(".", targetHash, def.PatchDir, treeHashes, def.ExcludePaths, false)
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
	filename := fs.String("f", "", "Distribution file")
	verbose := fs.Bool("v", false, "Be verbose")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	if *filename != "" {
		if err := applyDist(*filename); err != nil {
			return err
		}
	}
	c, err := hashchain.ReadFile(def.HashchainFile)
	if err != nil {
		return err
	}
	if err := c.Close(); err != nil {
		return err
	}
	return apply(c)
}
