package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/interrupt"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/seckey"
)

func createPkg(c *hashchain.HashChain, secKeyFile string) error {
	head := c.Head()
	fmt.Printf("create package for head %x\n", head)
	/*
		secKey, _, _, err := seckey.Read(secKeyFile)
		if err != nil {
			return err
		}
	*/
	return nil
}

// CreatePKG implements the ssotpub 'createpkg' command.
func CreatePkg(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -s seckey.bin\n", argv0)
		fmt.Fprintf(os.Stderr, "Create secure package and first signed head.\n")
		fs.PrintDefaults()
	}
	secKey := fs.String("s", "", "Secret key file")
	verbose := fs.Bool("v", false, "Be verbose")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
	}
	if err := seckey.Check(homedir.SSOTPub(), *secKey); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		fs.Usage()
		return flag.ErrHelp
	}
	c, err := hashchain.ReadFile(def.HashchainFile)
	if err != nil {
		return err
	}
	defer c.Close()
	// add interrupt handler
	interrupt.AddInterruptHandler(func() {
		c.Close()
	})
	// run createPkg
	go func() {
		if err := createPkg(c, *secKey); err != nil {
			interrupt.ShutdownChannel <- err
			return
		}
		interrupt.ShutdownChannel <- nil
	}()
	return <-interrupt.ShutdownChannel
}
