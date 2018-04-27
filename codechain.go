// codechain establishes code trust with a hashchain of threshold signatures.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/command"
)

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
	var err error
	switch os.Args[1] {
	case "treehash":
		err = command.TreeHash()
	case "treelist":
		err = command.TreeList()
	case "genkey":
		err = command.GenKey()
	case "pubkey":
		err = command.PubKey()
	case "init":
		err = command.InitChain()
	case "addkey":
		err = command.AddKey()
	case "verify":
		err = command.Verify()
	default:
		usage()
	}
	if err != nil && err != flag.ErrHelp {
		fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], err)
		os.Exit(1)
	}
}
