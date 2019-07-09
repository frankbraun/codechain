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
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/sync"
	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/git"
	"github.com/frankbraun/codechain/util/hex"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/interrupt"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/seckey"
	"github.com/frankbraun/codechain/util/terminal"
)

func publish(
	c *hashchain.HashChain, secKeyFile, message string,
	dryRun, useGit, yesPrompt bool,
	version int,
) error {
	var (
		secKey *[64]byte
		err    error
	)

	// get last published treehash
	treeHash := c.LastTreeHash()

	// make sure patch file doesn't exist for last tree hash
	patchFile := filepath.Join(def.PatchDir, treeHash)
	exists, err := file.Exists(patchFile)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%s: patch file already exists", patchFile)
	}

	// calculate current treehash
	curHash, err := tree.Hash(".", def.ExcludePaths)
	if err != nil {
		return err
	}
	curHashStr := hex.Encode(curHash[:])
	log.Printf("current tree hash: %s", curHashStr)

	// make sure the tree is dirty
	if curHashStr == treeHash {
		return fmt.Errorf("tree not dirty, nothing to publish")
	}

	// load secret key
	if !dryRun {
		secKey, _, _, err = seckey.Load(c, homedir.Codechain(), secKeyFile)
		if err != nil {
			return err
		}
	}

	// bring .codechain/tree/a in sync with last published treehash
	log.Println("sync tree/a")
	treeHashes := c.TreeHashes()
	err = sync.Dir(treeDirA, treeHash, def.PatchDir, treeHashes, def.ExcludePaths, true)
	if err != nil {
		return err
	}
	log.Println("done")

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

	if useGit && !yesPrompt {
		// display diff pager
		if err := git.DiffPager(treeDirA, treeDirB); err != nil {
			return err
		}
	} else {
		fmt.Println("the patch to publish is the diff between the following two directries:")
		fmt.Println(treeDirA)
		fmt.Println(treeDirB)
	}
	if dryRun {
		return nil
	}

	// confirm patch
	if yesPrompt {
		fmt.Println("patch published automatically (-y was used).")
	} else {
		if err := terminal.Confirm("publish patch?"); err != nil {
			return err
		}
	}

	// read comment
	var comment []byte
	if message != "" {
		comment = []byte(message)
	} else {
		fmt.Println("comment describing code change (can be empty; cannot be changed later):")
		comment, err = terminal.ReadLine(os.Stdin)
		if err != nil {
			return err
		}
	}

	// get and write patch
	f, err := os.OpenFile(patchFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	err = patchfile.Diff(version, f, treeDirA, treeDirB, def.ExcludePaths)
	if err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	log.Printf("%s: written\n", patchFile)

	// apply patch file to .codechain/tree/a to make sure it works
	treeHashes = append(treeHashes, curHashStr)
	err = sync.Dir(treeDirA, curHashStr, def.PatchDir, treeHashes, def.ExcludePaths, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: is faulty (this is a bug, please report it)\n",
			patchFile)
		return err
	}

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
	dryRun := fs.Bool("d", false, "Dry run, just show diff without signing anything")
	message := fs.String("m", "", "Use the given message as the comment describing the code change")
	useGit := fs.Bool("git", true, "Use git-diff to show diffs")
	secKey := fs.String("s", "", "Secret key file")
	verbose := fs.Bool("v", false, "Be verbose")
	version := fs.Int("version", patchfile.Version, "Patchfile version to publish")
	yesPrompt := fs.Bool("y", false, "Automatic yes to prompts, use with care!")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
	}
	if *version < 1 || *version > patchfile.Version {
		return patchfile.ErrHeaderVersion
	}
	if !*dryRun {
		if err := seckey.Check(homedir.Codechain(), *secKey); err != nil {
			return err
		}
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	if err := secpkg.UpToDate("codechain"); err != nil {
		return err
	}
	if err := os.MkdirAll(treeDirA, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(treeDirB, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(def.PatchDir, 0755); err != nil {
		return err
	}
	c, err := hashchain.ReadFile(def.HashchainFile)
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
		err := publish(c, *secKey, *message, *dryRun, *useGit, *yesPrompt, *version)
		if err != nil {
			interrupt.ShutdownChannel <- err
			return
		}
		interrupt.ShutdownChannel <- nil
	}()
	return <-interrupt.ShutdownChannel
}
