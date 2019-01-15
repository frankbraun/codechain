// file2ascii85 computes the ascii85 encoding of a file and prints it to stdout.
package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/frankbraun/codechain/internal/ascii85"
)

func file2ascii85(filename string) error {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	enc, _, err := ascii85.Encode(src)
	if err != nil {
		return err
	}
	if _, err := os.Stdout.Write(enc); err != nil {
		return err
	}
	return nil
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], err)
	os.Exit(1)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s file\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Encode file as ascii85 and print it to stdout.\n")
	os.Exit(2)
}

func main() {
	if len(os.Args) != 2 {
		usage()
	}
	if err := file2ascii85(os.Args[1]); err != nil {
		fatal(err)
	}
}
