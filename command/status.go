package command

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util"
	"github.com/frankbraun/codechain/util/hex"
	"github.com/frankbraun/codechain/util/log"
)

func showSigner(c *hashchain.HashChain) {
	fmt.Printf("signers (%d-of-%d required):\n", c.M(), c.N())
	var signer []string
	for s := range c.Signer() {
		signer = append(signer, s)
	}
	sort.Strings(signer)
	for _, s := range signer {
		fmt.Printf("%d %s %s\n", c.SignerWeight(s), s, c.SignerComment(s))
	}
}

func showSignedReleases(c *hashchain.HashChain) {
	_, idx := c.LastSignedTreeHash()
	if idx == 0 {
		fmt.Println("no signed releases yet")
		return
	}
	treeHashes := c.TreeHashes()
	treeComments := c.TreeComments()
	fmt.Println("signed releases:")
	for i := 1; i <= idx; i++ {
		fmt.Printf("%s %s\n", treeHashes[i], treeComments[i])
	}
}

func showUnsigned(c *hashchain.HashChain) error {
	infos, err := c.UnsignedInfo("", "", false)
	if err != nil {
		return err
	}
	if len(infos) == 0 {
		fmt.Println("no unsigned entries")
		return nil
	}
	fmt.Println("unsigned entries:")
	for _, info := range infos {
		fmt.Println(info)
	}
	return nil
}

func showTreeStatus(c *hashchain.HashChain) error {
	treeHash, err := tree.Hash(".", def.ExcludePaths)
	if err != nil {
		return err
	}
	treeHashes := c.TreeHashes()
	if util.ContainsString(treeHashes, hex.Encode(treeHash[:])) {
		fmt.Printf("tree matches %x\n", treeHash[:])
	} else {
		fmt.Println("tree is dirty")
	}
	return nil
}

func status(c *hashchain.HashChain) error {
	showSignedReleases(c)
	fmt.Println()
	showSigner(c)
	fmt.Println()
	if err := showUnsigned(c); err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("head:")
	fmt.Printf("%x\n", c.Head())
	fmt.Println()
	return showTreeStatus(c)
}

// Status implements the 'status' command.
func Status(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", argv0)
		fmt.Fprintf(os.Stderr, "Show status of hashchain and tree.\n")
		fs.PrintDefaults()
	}
	deepVerify := fs.Bool("deep-verify", false, "Verify all patch files match hash chain entries")
	print := fs.Bool("p", false, "Print hashchain to stdout")
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
	if err := secpkg.UpToDate("codechain"); err != nil {
		return err
	}
	c, err := hashchain.ReadFile(def.HashchainFile)
	if err != nil {
		return err
	}
	defer c.Close()
	if *deepVerify {
		err := c.DeepVerify(treeDirA, def.PatchDir, def.ExcludePaths)
		if err != nil {
			return err
		}
	}
	if *print {
		c.Print()
		return nil
	}
	return status(c)
}
