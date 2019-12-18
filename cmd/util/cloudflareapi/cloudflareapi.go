// cloudflareapi calls the Cloudflare API (https://api.cloudflare.com/).
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util"
	"github.com/frankbraun/codechain/util/cloudflare"
	"github.com/frankbraun/codechain/util/log"
)

func callCloudflareAPI(
	apiKey, email, zone, fqdn, data string,
	ttl int,
	update, del bool,
) error {
	s, err := cloudflare.New(apiKey, email)
	if err != nil {
		return err
	}
	if update {
		err := s.TXTUpdate(zone, fqdn, data, ttl)
		if err != nil {
			return err
		}
	} else if del {
		err = s.TXTDelete(zone, fqdn)
		if err != nil {
			return err
		}
	} else {
		jsn, err := s.TXTCreate(zone, fqdn, data, ttl)
		if err != nil {
			return err
		}
		fmt.Println(jsn)
	}
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Call Cloudflare API.\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	apiKey := flag.String("a", "", "API key")
	email := flag.String("e", "", "Email")
	zone := flag.String("zone", "", "Zone")
	fqdn := flag.String("fqdn", "", "FQDN")
	data := flag.String("data", "", "TXT data")
	ttl := flag.Int("ttl", ssot.TTL, "TTL")
	update := flag.Bool("update", false, "Update TXT record")
	del := flag.Bool("delete", false, "Delete TXT record")
	verbose := flag.Bool("v", false, "Be verbose")
	flag.Usage = usage
	flag.Parse()
	if *apiKey == "" {
		util.Fatal(errors.New("api key (-a) is mandatory"))
	}
	if *email == "" {
		util.Fatal(errors.New("email (-e) is mandatory"))
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
	err := callCloudflareAPI(*apiKey, *email, *zone, *fqdn, *data,
		*ttl, *update, *del)
	if err != nil {
		util.Fatal(err)
	}
}
