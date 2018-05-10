package command

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/patchfile"
	"github.com/frankbraun/codechain/sync"
	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/git"
	"github.com/frankbraun/codechain/util/interrupt"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/terminal"
)

func publish(c *hashchain.HashChain, secKeyFile string, dryRun bool) error {
	var (
		secKey *[64]byte
		err    error
	)
	// load secret key
	if !dryRun {
		secKey, _, _, err = seckeyLoad(c, secKeyFile)
		if err != nil {
			return err
		}
	}

	// get last published treehash
	treeHash := c.LastTreeHash()

	// make sure patch file doesn't exist for last tree hash
	patchFile := filepath.Join(patchDir, treeHash)
	exists, err := file.Exists(patchFile)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%s: patch file already exists", patchFile)
	}

	// bring .codechain/tree/a in sync with last published treehash
	log.Println("sync tree/a")
	treeHashes := c.TreeHashes()
	err = sync.Dir(treeDirA, treeHash, patchDir, treeHashes, def.ExcludePaths, true)
	if err != nil {
		return err
	}
	log.Println("done")

	// calculate current treehash
	curHash, err := tree.Hash(".", def.ExcludePaths)
	if err != nil {
		return err
	}
	log.Printf("%x\n", curHash[:])

	// bring .codechain/tree/b in sync with the tree hash to be published
	tmpHash, err := tree.Hash(treeDirB, def.ExcludePaths)
	if err != nil {
		return err
	}
	if !bytes.Equal(curHash[:], tmpHash[:]) {
		if err := os.RemoveAll(treeDirB); err != nil {
			return err
		}
		if err := file.CopyDirExclude(".", treeDirB, def.ExcludePaths); err != nil {
			return err
		}
	}

	// display diff pager
	if err := git.DiffPager(treeDirA, treeDirB); err != nil {
		return err
	}
	if dryRun {
		return nil
	}

	// confirm patch
	if err := terminal.Confirm("publish patch?"); err != nil {
		return err
	}

	// read comment
	fmt.Println("comment describing code change (can be empty):")
	comment, err := terminal.ReadLine(os.Stdin)
	if err != nil {
		return err
	}

	// get and write patch
	f, err := os.OpenFile(patchFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	err = patchfile.Diff(f, treeDirA, treeDirB, def.ExcludePaths)
	if err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	log.Printf("%s: written\n", patchFile)

	// sign patch and add to hash chain
	entry, err := c.Source(*curHash, *secKey, comment)
	if err != nil {
		return err
	}
	fmt.Println(entry)

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
	dryRun := fs.Bool("d", false, "Dry run, just show diff without signing anything")
	verbose := fs.Bool("v", false, "Be verbose")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !*dryRun {
		if err := seckeyCheck(*seckey); err != nil {
			return err
		}
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
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
		if err := publish(c, *seckey, *dryRun); err != nil {
			interrupt.ShutdownChannel <- err
			return
		}
		interrupt.ShutdownChannel <- nil
	}()
	return <-interrupt.ShutdownChannel
}
