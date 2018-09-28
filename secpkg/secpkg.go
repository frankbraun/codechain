// Package secpkg implements the secpkg package format.
package secpkg

import (
	"encoding/json"
	"io/ioutil"
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

// Marshal secpkg p as string.
func (pkg *Package) Marshal() string {
	jsn, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		panic(err) // should never happen
	}
	return string(jsn)
}
