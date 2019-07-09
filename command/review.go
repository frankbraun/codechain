package command

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/base64"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/sync"
	"github.com/frankbraun/codechain/util/git"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/interrupt"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/seckey"
	"github.com/frankbraun/codechain/util/terminal"
)

func showPatchInfo(
	c *hashchain.HashChain, i, idx int,
	treeHashes, treeComments []string,
	foundBarrier int,
) {
	pub, comment := c.SignerInfo(treeHashes[i])
	if foundBarrier > 0 {
		fmt.Printf("patch %d/%d\n", i-foundBarrier+1, idx-foundBarrier+1)
	} else { // after barrier
		fmt.Printf("patch %d/%d\n", i-idx, len(treeHashes)-idx-1)
	}
	if treeComments[i] != "" {
		fmt.Println(treeComments[i])
	}
	fmt.Printf("developer: %s\n", pub)
	if comment != "" {
		fmt.Println(comment)
	}
}

func procDiff(i int, treeHashes []string, useGit bool) error {
	// bring .codechain/tree/a in sync
	log.Println("bring .codechain/tree/a in sync")
	err := sync.Dir(treeDirA, treeHashes[i-1], def.PatchDir, treeHashes, def.ExcludePaths, true)
	if err != nil {
		return err
	}

	// bring .codechain/tree/b in sync
	log.Println("bring .codechain/tree/b in sync")
	err = sync.Dir(treeDirB, treeHashes[i], def.PatchDir, treeHashes, def.ExcludePaths, true)
	if err != nil {
		return err
	}

	if useGit {
		// display diff pager
		if err := git.DiffPager(treeDirA, treeDirB); err != nil {
			return err
		}
	} else {
		fmt.Println("review diff between the following two directries:")
		fmt.Println(treeDirA)
		fmt.Println(treeDirB)
	}
	return nil
}

func review(c *hashchain.HashChain, secKeyFile, treeHash string, detached, useGit bool) error {
	// load secret key
	log.Println("review(): load secret key")
	secKey, _, _, err := seckey.Load(c, homedir.Codechain(), secKeyFile)
	if err != nil {
		return err
	}
	log.Println("review(): loaded")

	// get last tree hashes
	_, idx := c.LastSignedTreeHash()
	treeHashes := c.TreeHashes()
	treeComments := c.TreeComments()
	if len(treeHashes) != len(treeComments) {
		return fmt.Errorf("invariant failed: len(treeHashes) == len(treeComments)")
	}

	// deal with explicit treehash
	if treeHash != "" {
		log.Printf("treehash=%s", treeHash)
		var i int
		for ; i < len(treeHashes); i++ {
			if treeHash == treeHashes[i] {
				log.Printf("treehash found at index %d", i)
				break
			}
		}
		if i == len(treeHashes) {
			return errors.New("cannot find treehash in hashchain")
		}
		if i <= idx {
			return errors.New("given treehash is already signed")
		}
		idx = i
	}

	if log.Std != nil {
		log.Println("treeHashes :")
		for _, h := range treeHashes {
			log.Println(h)
		}
		log.Println("treeComments:")
		for _, c := range treeComments {
			log.Println(c)
		}
	}

	// show changes in signers/sigctl
	var signed bool
	pubKey := base64.Encode(secKey[32:])
	infos, err := c.UnsignedInfo(pubKey, treeHash, true)
	if err != nil {
		return err
	}
	if len(infos) > 0 {
		fmt.Println("signer/sigctl changes:")
		for _, info := range infos {
			fmt.Println(info)
		}
		err := terminal.Confirm("confirm signer/sigctl changes?")
		if err != nil {
			return err
		}
		signed = true
	}

	// show commits which have been signed, but not by this signer
	barrier := c.SignerBarrier(pubKey)
	log.Printf("review(): barrier=%d\n", barrier)
	var foundBarrier int
outer:
	for i := 1; i <= idx; i++ {
		if c.SourceLine(treeHashes[i]) > barrier {
			if foundBarrier == 0 {
				foundBarrier = i
			}
			showPatchInfo(c, i, idx, treeHashes, treeComments, foundBarrier)
			err := terminal.Confirm("review already released patch (no continues)?")
			if err != nil {
				if err == terminal.ErrAbort {
					break outer
				}
				return err
			}
			if err := procDiff(i, treeHashes, useGit); err != nil {
				return err
			}
		}
	}

	for i := idx + 1; i < len(treeHashes); i++ {
		showPatchInfo(c, i, idx, treeHashes, treeComments, 0)
		if c.SourceLine(treeHashes[i]) > barrier {
			if err := terminal.Confirm("review patch (no aborts)?"); err != nil {
				return err
			}
			if err := procDiff(i, treeHashes, useGit); err != nil {
				return err
			}
			if err := terminal.Confirm("sign patch?"); err != nil {
				return err
			}
			signed = true
		} else {
			fmt.Println("skipping already signed patch")
		}
	}

	if !signed {
		err := terminal.Confirm("no new signer/sigctl changes or source publications to sign\n" +
			"sign anyway?")
		if err != nil {
			return err
		}
	}

	// sign patches and add to hash chain
	var linkHash [32]byte
	if treeHash != "" {
		linkHash = c.LinkHash(treeHash)
	} else {
		linkHash = c.Head()
	}
	entry, err := c.Signature(linkHash, *secKey, detached)
	if err != nil {
		return err
	}
	fmt.Println(entry)
	return nil
}

func addDetached(c *hashchain.HashChain, linkHash, pubKey, signature string) error {
	entry, err := c.DetachedSignature(linkHash, pubKey, signature)
	if err != nil {
		return err
	}
	fmt.Println(entry)
	return nil
}

// Review implements the 'review' command.
func Review(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-s seckey.bin] [treehash]\n", argv0)
		fmt.Fprintf(os.Stderr, "       %s -a linkhash pubkey signature\n", argv0)
		fmt.Fprintf(os.Stderr, "Review code changes (all or up to treehash) and changes of signers and sigctl.\n")
		fs.PrintDefaults()
	}
	add := fs.Bool("a", false, "Add detached signature")
	detached := fs.Bool("d", false, "Create detached signature")
	useGit := fs.Bool("git", true, "Use git-diff to show diffs")
	secKey := fs.String("s", "", "Secret key file")
	verbose := fs.Bool("v", false, "Be verbose")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
	}
	if err := seckey.Check(homedir.Codechain(), *secKey); err != nil {
		return err
	}
	if *add {
		if fs.NArg() != 3 {
			fs.Usage()
			return flag.ErrHelp
		}
	} else {
		if fs.NArg() != 0 && fs.NArg() != 1 {
			fs.Usage()
			return flag.ErrHelp
		}
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
	var treeHash string
	if fs.NArg() == 1 {
		treeHash = fs.Arg(0)
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
	// run review
	go func() {
		var err error
		if *add {
			err = addDetached(c, fs.Arg(0), fs.Arg(1), fs.Arg(2))
		} else {
			err = review(c, *secKey, treeHash, *detached, *useGit)
		}
		if err != nil {
			interrupt.ShutdownChannel <- err
			return
		}
		interrupt.ShutdownChannel <- nil
	}()
	return <-interrupt.ShutdownChannel
}
