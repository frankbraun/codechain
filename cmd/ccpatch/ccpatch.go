// ccpatch applies a patchfile to a directory tree.
package main

import (
	"fmt"
	"os"

	"github.com/frankbraun/codechain/command"
	"github.com/frankbraun/codechain/patchfile"
	"github.com/frankbraun/codechain/util"
)

func patch(dir, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return patchfile.Apply(dir, f, command.ExcludePaths)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s directory patchfile\n", os.Args[0])
	os.Exit(2)
}

func main() {
	if len(os.Args) != 3 {
		usage()
	}
	if err := patch(os.Args[1], os.Args[2]); err != nil {
		util.Fatal(err)
	}
}
