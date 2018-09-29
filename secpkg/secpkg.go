// Package secpkg implements the secpkg package format.
package secpkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/ssot"
	"github.com/frankbraun/codechain/util/file"
)

// Package defines a package in secpkg format (stored in .secpkg files).
type Package struct {
	Name string // the project's package name
	Head string // head of project's Codechain
	DNS  string // fully qualified domain name for _codechain TXT records (SSOT)
	URL  string // URL to download project files of the from (URL/head.tar.gz)
}

// Load a .secpkg file from filename and return the Package struct.
func Load(filename string) (*Package, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var pkg Package
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}
	return &pkg, err
}

// Marshal pkg as string.
func (pkg *Package) Marshal() string {
	jsn, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		panic(err) // should never happen
	}
	return string(jsn)
}

// Install pkg.
func (pkg *Package) Install() error {
	txts, err := net.LookupTXT(def.CodechainTXTName + pkg.DNS)
	if err != nil {
		return err
	}
	// Parse TXT records and look for signed head
	var sh *ssot.SignedHead
	for _, txt := range txts {
		sh, err = ssot.Unmarshal(txt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot unmarshal: %s\n", txt)
			continue
		}
		fmt.Println(sh.Head())
		break /// TXT record found
	}
	if sh == nil {
		return errors.New("secpkg: no valid TXT record found")
	}
	// TODO: trust pubkey on first use
	// TODO: compare pubkey with stored one
	// TODO: pubkey rotation

	// download distribution
	filename := sh.Head() + ".tar.gz"
	url := pkg.URL + "/" + filename
	fmt.Printf("download %s\n", url)
	return file.Download(filename, url)
}
