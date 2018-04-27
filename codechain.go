// codechain establishes code trust with a hashchain of threshold signatures.
package main

import (
	"fmt"
	"os"

	"github.com/frankbraun/codechain/command"
	"github.com/frankbraun/codechain/tree"
)

var excludePaths = []string{
	command.CodechainDir,
	".git",
	".gitignore",
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], err)
	os.Exit(1)
}

func usage() {
	cmd := os.Args[0]
	fmt.Fprintf(os.Stderr, "Usage: %s treehash\n", cmd)
	fmt.Fprintf(os.Stderr, "       %s treelist\n", cmd)
	fmt.Fprintf(os.Stderr, "       %s genkey [-s seckey.bin]\n", cmd)
	fmt.Fprintf(os.Stderr, "       %s pubkey -s seckey.bin\n", cmd)
	fmt.Fprintf(os.Stderr, "       %s init [-m]\n", cmd)
	fmt.Fprintf(os.Stderr, "       %s addkey [-w] pubkey signature [comment]\n", cmd)
	fmt.Fprintf(os.Stderr, "       %s verify\n", cmd)
	os.Exit(2)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "treehash":
		hash, err := tree.Hash(".", excludePaths)
		if err != nil {
			fatal(err)
		}
		fmt.Printf("%x\n", hash[:])
	case "treelist":
		list, err := tree.List(".", excludePaths)
		if err != nil {
			fatal(err)
		}
		os.Stdout.Write(list)
	case "genkey":
		if err := command.GenKey(); err != nil {
			fatal(err)
		}
	case "pubkey":
		if err := command.PubKey(); err != nil {
			fatal(err)
		}
	case "init":
		if err := command.InitChain(); err != nil {
			fatal(err)
		}
	case "addkey":
		if err := command.AddKey(); err != nil {
			fatal(err)
		}
	case "verify":
		if err := command.Verify(); err != nil {
			fatal(err)
		}
	default:
		usage()
	}
}
