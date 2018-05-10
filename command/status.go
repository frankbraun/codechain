package command

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/util/log"
)

// TODO: move to hashchain?
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

func status(c *hashchain.HashChain) error {
	showSignedReleases(c)
	fmt.Println()
	showSigner(c)
	fmt.Println()
	return showUnsigned(c)
}

// Status implement the 'status' command.
func Status(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", argv0)
		fmt.Fprintf(os.Stderr, "Show status of hashchain and tree.\n")
		fs.PrintDefaults()
	}
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
	c, err := hashchain.Read(hashchainFile)
	if err != nil {
		return err
	}
	defer c.Close()
	if *print {
		c.Print()
		return nil
	}
	return status(c)
}
