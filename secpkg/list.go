package secpkg

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/homedir"
)

// List all installed packages and return them as.
func List() ([]string, error) {
	pkgDir := filepath.Join(homedir.SecPkg(), "pkgs")
	exists, err := file.Exists(pkgDir)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("no package installed: '%s' does not exist", pkgDir)
	}
	files, err := ioutil.ReadDir(pkgDir)
	if err != nil {
		return nil, err
	}
	var pkgs []string
	for _, file := range files {
		pkgs = append(pkgs, file.Name())
	}
	return pkgs, nil
}
