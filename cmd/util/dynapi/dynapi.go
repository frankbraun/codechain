// dynapi calls the Dyn Managed DNS API (https://help.dyn.com/dns-api-knowledge-base/).
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util"
	"github.com/frankbraun/codechain/util/dyn"
	"github.com/frankbraun/codechain/util/log"
)

func callDynAPI(
	customerName, userName, password, zone, fqdn, data string,
	ttl int,
	update, del bool,
) error {
	s, err := dyn.New(customerName, userName, password)
	if err != nil {
		return err
	}
	defer s.Close()
	if update {
		if err := s.TXTUpdate(zone, fqdn, data, ttl); err != nil {
			return err
		}
	} else if del {
		if err := s.TXTDelete(zone, fqdn); err != nil {
			return err
		}
	} else {
		if err := s.TXTCreate(zone, fqdn, data, ttl); err != nil {
			return err
		}
	}
	ret, err := s.ZoneChangeset(zone)
	if err != nil {
		return err
	}
	jsn, err := json.MarshalIndent(ret, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsn))
	if err := s.ZoneUpdate(zone); err != nil {
		return err
	}
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
	zone := flag.String("zone", "", "Zone")
	fqdn := flag.String("fqdn", "", "FQDN")
	data := flag.String("data", "", "TXT data")
	ttl := flag.Int("ttl", ssot.TTL, "TTL")
	update := flag.Bool("update", false, "Update TXT record")
	del := flag.Bool("delete", false, "Delete TXT record")
	verbose := flag.Bool("v", false, "Be verbose")
	flag.Usage = usage
	flag.Parse()
	if *customerName == "" {
		util.Fatal(errors.New("customer name (-c) is mandatory"))
	}
	if *userName == "" {
		util.Fatal(errors.New("user name (-u) is mandatory"))
	}
	if *password == "" {
		util.Fatal(errors.New("password (-p) is mandatory"))
	}
	if *zone == "" {
		util.Fatal(errors.New("zone (-zone) is mandatory"))
	}
	if *fqdn == "" {
		util.Fatal(errors.New("fqdn (-fqdn) is mandatory"))
	}
	if *data == "" && !*del {
		util.Fatal(errors.New("data (-data) is mandatory"))
	}
	if *update && *del {
		util.Fatal(errors.New("-update and -delete exclude each other"))
	}
	if *verbose {
		log.Std = log.NewStd(os.Stdout)
	}
	if flag.NArg() != 0 {
		usage()
	}
	err := callDynAPI(*customerName, *userName, *password, *zone, *fqdn, *data,
		*ttl, *update, *del)
	if err != nil {
		util.Fatal(err)
	}
}
