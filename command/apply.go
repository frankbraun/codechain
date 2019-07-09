package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/archive"
	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/secpkg"
	"github.com/frankbraun/codechain/util/hex"
	"github.com/frankbraun/codechain/util/log"
)

// Apply implements the 'apply' command.
func Apply(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", argv0)
		fmt.Fprintf(os.Stderr, "Apply all patches with enough signatures to code tree.\n")
		fs.PrintDefaults()
	}
	filename := fs.String("f", "", "Distribution file")
	headStr := fs.String("head", "", "Check that the hash chain contains the given head")
	verbose := fs.Bool("v", false, "Be verbose")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	if err := secpkg.UpToDate("codechain"); err != nil {
		return err
	}
	var head *[32]byte
	if *headStr != "" {
		h, err := hex.Decode(*headStr, 32)
		if err != nil {
			return err
		}
		var ha [32]byte
		copy(ha[:], h)
		head = &ha
	}
	if *filename != "" {
		err := archive.ApplyFile(def.HashchainFile, def.PatchDir, *filename, head)
		if err != nil {
			return err
		}
	}
	c, err := hashchain.ReadFile(def.HashchainFile)
	if err != nil {
		return err
	}
	if err := c.Close(); err != nil {
		return err
	}
	return c.Apply(head)
}
