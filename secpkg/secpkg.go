package secpkg

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/frankbraun/codechain/internal/hex"
)

// File defines the default file (ending) for a secure package.
const File = ".secpkg"

// Package defines a package in secpkg format (stored in .secpkg files).
type Package struct {
	Name string // the project's package name
	Head string // head of project's Codechain
	DNS  string // fully qualified domain name for Codechain's TXT records (SSOT)
}

// New creates a new Package.
func New(name, dns string, head [32]byte) (*Package, error) {
	// validate arguments
	if strings.Contains(name, " ") {
		return nil, ErrPkgNameWhitespace
	}
	if _, err := url.Parse(dns); err != nil {
		return nil, err
	}
	// create package
	var pkg Package
	pkg.Name = strings.ToLower(name) // project names are lowercase
	pkg.Head = hex.Encode(head[:])
	pkg.DNS = dns
	return &pkg, nil
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
