// ccpatch caluates a patch between two directory trees and prints it to stdout.
package main

import (
	"fmt"
	"os"

	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/patchfile"
	"github.com/frankbraun/codechain/util"
)

func diff(a, b string) error {
	return patchfile.Diff(os.Stdout, a, b, def.ExcludePaths)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s tree_a tree_b\n", os.Args[0])
	os.Exit(2)
}

func main() {
	if len(os.Args) != 3 {
		usage()
	}
	if err := diff(os.Args[1], os.Args[2]); err != nil {
		util.Fatal(err)
	}
}
