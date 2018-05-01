package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/git"
	"github.com/frankbraun/codechain/util/home"
)

// Publish implements the 'publish' command.
func Publish(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-s seckey.bin]\n", argv0)
		fmt.Fprintf(os.Stderr, "Add signed changes in tree to .codechain ready for publication.\n")
		fs.PrintDefaults()
	}
	seckey := fs.String("s", "", "Secret key file")
	verbose := fs.Bool("v", false, "Be verbose")
	if err := fs.Parse(args); err != nil {
		return err
	}
	var homeDir string
	if *seckey != "" {
		exists, err := file.Exists(*seckey)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("file '%s' doesn't exists", *seckey)
		}
	} else {
		homeDir = home.AppDataDir("codechain", false)
		homeDir = filepath.Join(homeDir, secretsDir)
		// make sure we have the secrets directory at least present
		exists, err := file.Exists(homeDir)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("directory '%s' doesn't exists: you have no secrets",
				homeDir)
		}
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	if err := os.MkdirAll(treeDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(patchDir, 0755); err != nil {
		return err
	}

	c, err := hashchain.Read(hashchainFile)
	if err != nil {
		return err
	}
	defer c.Close()

	// get last published treehash
	treeHash := c.LastTreeHash()

	// bring .codechain/tree in sync with last published treehash
	err = tree.Sync(treeDir, treeHash, patchDir, *verbose, excludePaths)
	if err != nil {
		return err
	}

	// calculate current treehash
	hash, err := tree.Hash(".", excludePaths)
	if err != nil {
		return err
	}
	fmt.Printf("%x\n", hash[:])

	// display diff
	if err := git.DiffPager(treeDir, "."); err != nil {
		return err
	}

	// TODO: display diff

	// TODO: sign diff and add to hash chain

	return errors.New("not implemented")
}
