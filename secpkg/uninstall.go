package secpkg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/gnumake"
	"github.com/frankbraun/codechain/util/homedir"
)

// Uninstall package with name.
func Uninstall(name string) error {
	// 1. Make sure the project with NAME has been installed before.
	//    That is, the directory ~/.config/secpkg/pkgs/NAME exists.
	pkgDir := filepath.Join(homedir.SecPkg(), "pkgs", name)
	exists, err := file.Exists(pkgDir)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("package not installed: '%s' does not exist", pkgDir)
	}

	// 2. Call `make prefix= ~/.config/secpkg/local uninstall` in
	//    ~/.config/secpkg/pkgs/NAME/installed
	installedDir := filepath.Join(pkgDir, "installed")
	localDir := filepath.Join(homedir.SecPkg(), "local")
	if err := os.Chdir(installedDir); err != nil {
		return err
	}
	if err := gnumake.Uninstall(localDir); err != nil {
		return err
	}

	// 3. Remove package directory ~/.config/secpkg/pkgs/NAME
	return os.RemoveAll(pkgDir)
}
