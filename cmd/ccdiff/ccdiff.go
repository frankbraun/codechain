// ccpatch caluates a patch between two directory trees and prints it to stdout.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/patchfile"
	"github.com/frankbraun/codechain/util"
	"github.com/frankbraun/codechain/util/log"
)

func diff(a, b string) error {
	return patchfile.Diff(os.Stdout, a, b, def.ExcludePaths)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s tree_a tree_b\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Caluates a patch between two directory trees and print it to stdout.\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	verbose := flag.Bool("v", false, "Be verbose (on stderr)")
	flag.Usage = usage
	flag.Parse()
	if *verbose {
		log.Std = log.NewStd(os.Stderr)
	}
	if flag.NArg() != 2 {
		usage()
	}
	if err := diff(flag.Arg(0), flag.Arg(1)); err != nil {
		util.Fatal(err)
	}
}
