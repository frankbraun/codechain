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
	argv0 := os.Args[0] + " " + os.Args[1]
	args := os.Args[2:]
	var err error
	switch os.Args[1] {
	case "treehash":
		err = command.TreeHash(argv0, args...)
	case "treelist":
		err = command.TreeList(argv0, args...)
	case "genkey":
		err = command.GenKey(argv0, args...)
	case "pubkey":
		err = command.PubKey(argv0, args...)
	case "init":
		err = command.InitChain(argv0, args...)
	case "addkey":
		err = command.AddKey(argv0, args...)
	case "verify":
		err = command.Verify(argv0, args...)
	default:
		usage()
	}
	if err != nil {
		if err != flag.ErrHelp {
			fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], err)
			os.Exit(1)
		}
		os.Exit(2)
	}
}
