package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/interrupt"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/seckey"
)

func sign(c *hashchain.HashChain, secKeyFile string) error {
	head := c.Head()
	fmt.Printf("signing head %x\n", head)
	secKey, _, _, err := seckey.Read(secKeyFile)
	if err != nil {
		return err
	}
	// TODO: counter
	sh := ssot.SignHead(head, 0, *secKey)
	// print TXT entry
	fmt.Printf("_codechain.example.com.\t\t%d\tIN\tTXT\t\"value=%s\"\n", ssot.TTL,
		sh.Marshal())
	return nil
}

// Sign implements the ssotpub 'sign' command.
func Sign(argv0 string, args ...string) error {
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -s seckey.bin\n", argv0)
		fmt.Fprintf(os.Stderr, "Sign Codechain head and print it on stdout.\n")
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
	// run publish
	go func() {
		if err := sign(c, *secKey); err != nil {
			interrupt.ShutdownChannel <- err
			return
		}
		interrupt.ShutdownChannel <- nil
	}()
	return <-interrupt.ShutdownChannel
}
