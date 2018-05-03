package command

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/frankbraun/codechain/hashchain"
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

func showUnsignedReleases(c *hashchain.HashChain) {
	_, idx := c.LastSignedTreeHash()
	treeHashes := c.TreeHashes()
	if idx == len(treeHashes)-1 {
		fmt.Println("no unssigned releases")
		return
	}
	treeComments := c.TreeComments()
	fmt.Println("unsigned releases:")
	for i := idx + 1; i < len(treeHashes); i++ {
		fmt.Printf("%s %s\n", treeHashes[i], treeComments[i])
	}
}

func status(c *hashchain.HashChain) error {
	showSigner(c)
	showSignedReleases(c)
	showUnsignedReleases(c)
	return nil
}

// Status implement the 'status' command.
func Status(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", argv0)
		fmt.Fprintf(os.Stderr, "Show status of hashchain and tree.\n")
		fs.PrintDefaults()
	}
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
	defer c.Close()
	return status(c)
}
