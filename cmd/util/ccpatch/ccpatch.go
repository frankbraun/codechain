// ccpatch applies a patchfile to a directory tree.
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

func patch(dir, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return patchfile.Apply(dir, f, def.ExcludePaths)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s directory patchfile\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Apply a patchfile to a directory tree.\n")
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
	if err := patch(flag.Arg(0), flag.Arg(1)); err != nil {
		util.Fatal(err)
	}
}
