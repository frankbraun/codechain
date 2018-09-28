// ssotpub publishes Codechain heads with a single source of truth (SSOT).
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/ssot/command"
)

func usage() {
	cmd := os.Args[0]
	fmt.Fprintf(os.Stderr, "Usage: %s keygen [-s seckey.bin]\n", cmd)
	fmt.Fprintf(os.Stderr, "       %s keyfile -s seckey.bin [-c]\n", cmd)
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
	case "keygen":
		err = command.KeyGen(argv0, args...)
	case "keyfile":
		err = command.KeyFile(argv0, args...)
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
