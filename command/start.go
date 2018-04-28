package command

import (
	"flag"
	"fmt"
	"os"

	//"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/util/file"
)

// Start implements the 'start' command.
func Start(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-s]\n", argv0)
		fmt.Fprintf(os.Stderr, "Initialized new .codechain/hashchain in current directory.\n")
		fs.PrintDefaults()
	}
	//seckey := fs.String("s", "", "Secret key file")
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	if err := os.MkdirAll(codechainDir, 0700); err != nil {
		return err
	}
	exists, err := file.Exists(hashchainFile)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%s: file '%s' exists already", argv0, hashchainFile)
	}

	/* TODO:

	c, entry, err := hashchain.Start(hashchainFile)
	if err != nil {
		return err
	}
	defer c.Close()
	fmt.Println(entry)
	*/
	return nil
}
