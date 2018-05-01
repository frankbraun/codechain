package command

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/git"
	"github.com/frankbraun/codechain/util/home"
	"github.com/frankbraun/codechain/util/interrupt"
	"github.com/frankbraun/codechain/util/terminal"
)

func publish(c *hashchain.HashChain, verbose bool) error {

	// get last published treehash
	treeHash := c.LastTreeHash()

	// bring .codechain/tree/a in sync with last published treehash
	err := tree.Sync(treeDirA, treeHash, patchDir, verbose, excludePaths)
	if err != nil {
		return err
	}

	// calculate current treehash
	curHash, err := tree.Hash(".", excludePaths)
	if err != nil {
		return err
	}
	fmt.Printf("%x\n", curHash[:])

	// bring .codechain/tree/b in sync with last published treehash
	tmpHash, err := tree.Hash(treeDirB, excludePaths)
	if err != nil {
		return err
	}
	if !bytes.Equal(curHash[:], tmpHash[:]) {
		if err := os.RemoveAll(treeDirB); err != nil {
			return err
		}
		if err := file.CopyDirExclude(".", treeDirB, excludePaths); err != nil {
			return err
		}
	}

	// display diff pager
	if err := git.DiffPager(treeDirA, treeDirB); err != nil {
		return err
	}

	for {
		fmt.Print("publish path? [y/n]: ")
		answer, err := terminal.ReadLine(os.Stdin)
		if err != nil {
			return err
		}
		a := string(bytes.ToLower(answer))
		if strings.HasPrefix(a, "y") {
			break
		} else if strings.HasPrefix(a, "n") {
			return errors.New("aborted")
		} else {
			fmt.Println("answer not recognized")
		}
	}

	// get patch
	patch, err := git.Diff(treeDirA, treeDirB)
	if err != nil {
		return err
	}
	fmt.Printf(patch)

	// TODO: sign diff and add to hash chain

	return nil
}

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
	if err := os.MkdirAll(treeDirA, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(treeDirB, 0755); err != nil {
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
	// add interrupt handler
	interrupt.AddInterruptHandler(func() {
		c.Close()
	})
	// run publish
	go func() {
		if err := publish(c, *verbose); err != nil {
			interrupt.ShutdownChannel <- err
			return
		}
		interrupt.ShutdownChannel <- nil
	}()
	return <-interrupt.ShutdownChannel
}
