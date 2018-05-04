package command

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/base64"
	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util/git"
	"github.com/frankbraun/codechain/util/interrupt"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/terminal"
)

func review(c *hashchain.HashChain, secKeyFile, treeHash string) error {
	// load secret key
	secKey, _, _, err := seckeyLoad(c, secKeyFile)
	if err != nil {
		return err
	}

	// get last tree hashes
	_, idx := c.LastSignedTreeHash()
	/*
		if signedTreeHash == c.LastTreeHash() {
			fmt.Printf("%s: already signed\n", signedTreeHash)
			return nil
		}
	*/
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
		for {
			fmt.Print("confirm signer/sigctl changes? [y/n]: ")
			answer, err := terminal.ReadLine(os.Stdin)
			if err != nil {
				return err
			}
			a := string(bytes.ToLower(answer))
			if strings.HasPrefix(a, "y") {
				signed = true
				break
			} else if strings.HasPrefix(a, "n") {
				return errors.New("aborted")
			} else {
				fmt.Println("answer not recognized")
			}
		}
	}

	// show commits which have been signed, but not by this signer
	barrier := c.SignerBarrier(pubKey)
	// TODO: deduplicate code
outer:
	for i := 1; i <= idx; i++ {
		if c.SourceLine(treeHashes[i]) > barrier {
			// show patche info
			pub, comment := c.SignerInfo(treeHashes[i])
			fmt.Printf("patch %d/%d\n", i-idx, len(treeHashes)-idx-1)
			if treeComments[i] != "" {
				fmt.Println(treeComments[i])
			}
			fmt.Printf("developer: %s\n", pub)
			if comment != "" {
				fmt.Println(comment)
			}
			for {
				fmt.Print("review already signed patch (no continues)? [y/n]: ")
				answer, err := terminal.ReadLine(os.Stdin)
				if err != nil {
					return err
				}
				a := string(bytes.ToLower(answer))
				if strings.HasPrefix(a, "y") {
					break
				} else if strings.HasPrefix(a, "n") {
					continue outer
				} else {
					fmt.Println("answer not recognized")
				}
			}

			// bring .codechain/tree/a in sync
			log.Println("bring .codechain/tree/a in sync")
			err = tree.Sync(treeDirA, treeHashes[i-1], patchDir, treeHashes, excludePaths, true)
			if err != nil {
				return err
			}

			// bring .codechain/tree/b in sync
			log.Println("bring .codechain/tree/b in sync")
			err = tree.Sync(treeDirB, treeHashes[i], patchDir, treeHashes, excludePaths, true)
			if err != nil {
				return err
			}

			// display diff *pager
			if err := git.DiffPager(treeDirA, treeDirB); err != nil {
				return err
			}
		}
	}

	for i := idx + 1; i < len(treeHashes); i++ {
		// bring .codechain/tree/a in sync
		log.Println("bring .codechain/tree/a in sync")
		err = tree.Sync(treeDirA, treeHashes[i-1], patchDir, treeHashes, excludePaths, true)
		if err != nil {
			return err
		}

		// bring .codechain/tree/b in sync
		log.Println("bring .codechain/tree/b in sync")
		err = tree.Sync(treeDirB, treeHashes[i], patchDir, treeHashes, excludePaths, true)
		if err != nil {
			return err
		}

		// show patche info
		pub, comment := c.SignerInfo(treeHashes[i])
		fmt.Printf("patch %d/%d\n", i-idx, len(treeHashes)-idx-1)
		if treeComments[i] != "" {
			fmt.Println(treeComments[i])
		}
		fmt.Printf("developer: %s\n", pub)
		if comment != "" {
			fmt.Println(comment)
		}
		for {
			fmt.Print("review patch (no aborts)? [y/n]: ")
			answer, err := terminal.ReadLine(os.Stdin)
			if err != nil {
				return err
			}
			a := string(bytes.ToLower(answer))
			if strings.HasPrefix(a, "y") {
				signed = true
				break
			} else if strings.HasPrefix(a, "n") {
				return errors.New("aborted")
			} else {
				fmt.Println("answer not recognized")
			}
		}

		// display diff *pager
		if err := git.DiffPager(treeDirA, treeDirB); err != nil {
			return err
		}

		// confirm patch
		for {
			fmt.Print("sign patch? [y/n]: ")
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
	}

	if !signed {
		for {
			fmt.Println("no new signer/sigctl changes or source publications to sign")
			fmt.Print("sign anyway? [y/n]: ")
			answer, err := terminal.ReadLine(os.Stdin)
			if err != nil {
				return err
			}
			a := string(bytes.ToLower(answer))
			if strings.HasPrefix(a, "y") {
				signed = true
				break
			} else if strings.HasPrefix(a, "n") {
				return errors.New("aborted")
			} else {
				fmt.Println("answer not recognized")
			}
		}
	}

	// sign patches and add to hash chain
	var linkHash [32]byte
	if treeHash != "" {
		linkHash = c.LinkHash(treeHash)
	} else {
		linkHash = c.LastEntryHash()
	}
	entry, err := c.Signature(linkHash, *secKey)
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
		fmt.Fprintf(os.Stderr, "Review code changes (all or up to treehash) and changes of signers and sigctl.\n")
		fs.PrintDefaults()
	}
	seckey := fs.String("s", "", "Secret key file")
	verbose := fs.Bool("v", false, "Be verbose")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := seckeyCheck(*seckey); err != nil {
		return err
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
	}
	if fs.NArg() != 0 && fs.NArg() != 1 {
		fs.Usage()
		return flag.ErrHelp
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
	c, err := hashchain.Read(hashchainFile)
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
		if err := review(c, *seckey, treeHash); err != nil {
			interrupt.ShutdownChannel <- err
			return
		}
		interrupt.ShutdownChannel <- nil
	}()
	return <-interrupt.ShutdownChannel
}
