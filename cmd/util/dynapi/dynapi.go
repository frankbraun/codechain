// dynapi calls the Dyn Managed DNS API (https://help.dyn.com/dns-api-knowledge-base/).
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/util"
	"github.com/frankbraun/codechain/util/dyn"
	"github.com/frankbraun/codechain/util/log"
)

func callDynAPI(customerName, userName, password string) error {
	s, err := dyn.New(customerName, userName, password)
	if err != nil {
		return err
	}
	defer s.Close()
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Call Dyn Managed DNS API.\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	customerName := flag.String("c", "", "Customer name")
	userName := flag.String("u", "", "User name")
	password := flag.String("p", "", "Password")
	verbose := flag.Bool("v", false, "Be verbose")
	flag.Usage = usage
	flag.Parse()
	if *customerName == "" {
		util.Fatal(errors.New("Customer name (-c) is mandatory."))
	}
	if *userName == "" {
		util.Fatal(errors.New("User name (-u) is mandatory."))
	}
	if *password == "" {
		util.Fatal(errors.New("Password (-p) is mandatory."))
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
	}
	if flag.NArg() != 0 {
		usage()
	}
	err := callDynAPI(*customerName, *userName, *password)
	if err != nil {
		util.Fatal(err)
	}
}
