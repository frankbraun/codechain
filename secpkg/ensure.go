package secpkg

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/util/def"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/hex"
	"github.com/frankbraun/codechain/util/homedir"
	"github.com/frankbraun/codechain/util/log"
)

// ensure the secure dependencies for package name are installed and up-to-date.
func ensure(
	ctx context.Context,
	visited map[string]bool,
	name string,
) (bool, error) {
	// If the directory ~/.config/secpkg/pkgs/NAME/src/.secdep exists and
	// contains any .secpkg files, ensure these secure dependencies are
	// installed and up-to-date.
	secdepDir := filepath.Join(homedir.SecPkg(), "pkgs", name, "src", ".secdep")
	exists, err := file.Exists(secdepDir)
	if err != nil {
		return false, err
	}
	if !exists {
		log.Println(".secdep: no dependencies found")
		return false, nil // no dependencies found
	}
	log.Printf(".secdep: scanning dir '%s'\n", secdepDir)

	// process .secdep directory
	files, err := ioutil.ReadDir(secdepDir)
	if err != nil {
		return false, err
	}
	depUpdated := false
	for _, fi := range files {
		if !strings.HasSuffix(fi.Name(), ".secpkg") {
			log.Printf(".secdep: skip '%s'", fi.Name())
			continue // not a .secpkg file
		}
		// load .secpkg file
		log.Printf(".secdep: load '%s'", fi.Name())
		pkg, err := Load(filepath.Join(secdepDir, fi.Name()))
		if err != nil {
			return false, err
		}
		// check for cycles
		if visited[pkg.Name] {
			return false, fmt.Errorf("secpkg: dependency cycle detected for package '%s'",
				pkg.Name)
		}
		// check if it is already installed
		pkgDir := filepath.Join(homedir.SecPkg(), "pkgs", pkg.Name)
		exists, err := file.Exists(pkgDir)
		if err != nil {
			return false, err
		}
		visited[pkg.Name] = true
		if !exists {
			// install
			log.Printf(".secdep: install package '%s'\n", pkg.Name)
			if err := pkg.install(ctx, visited); err != nil {
				return false, err
			}
			depUpdated = true
		} else {
			// parse head
			h, err := hex.Decode(pkg.Head, 32)
			if err != nil {
				return false, err
			}
			var head [32]byte
			copy(head[:], h)
			// update
			log.Printf(".secdep: update package '%s'\n", pkg.Name)
			updated, err := update(ctx, visited, pkg.Name)
			if err != nil {
				return false, err
			}
			if updated {
				depUpdated = true
			}
			// make sure HEAD of .secpkg is actually contained in hash chain
			// (that is, we have updated the correct package).
			hashchainFile := filepath.Join(pkgDir, "src", def.HashchainFile)
			c, err := hashchain.ReadFile(hashchainFile)
			if err != nil {
				return false, err
			}
			if err := c.Close(); err != nil {
				return false, err
			}
			if err := c.CheckHead(head); err != nil {
				if err == hashchain.ErrHeadNotFound {
					return false, fmt.Errorf("secpkg: head '%s' of .secpkg '%s' not found in '%s'. "+
						"Conflicting packages?", pkg.Head, fi.Name(), hashchainFile)
				}
				return false, err
			}
		}
		delete(visited, pkg.Name)
	}

	return depUpdated, nil
}

// ensureCheckUpdate ensures the secure dependencies for package name are up-to-date.
func ensureCheckUpdate(
	ctx context.Context,
	visited map[string]bool,
	name string,
) (bool, error) {
	// If the directory ~/.config/secpkg/pkgs/NAME/src/.secdep exists and
	// contains any .secpkg files, ensure these secure dependencies are
	// up-to-date.
	secdepDir := filepath.Join(homedir.SecPkg(), "pkgs", name, "src", ".secdep")
	exists, err := file.Exists(secdepDir)
	if err != nil {
		return false, err
	}
	if !exists {
		log.Println(".secdep: no dependencies found")
		return false, nil // no dependencies found
	}
	log.Printf(".secdep: scanning dir '%s'\n", secdepDir)

	// process .secdep directory
	files, err := ioutil.ReadDir(secdepDir)
	if err != nil {
		return false, err
	}
	needsUpdate := false
	for _, fi := range files {
		if !strings.HasSuffix(fi.Name(), ".secpkg") {
			log.Printf(".secdep: skip '%s'", fi.Name())
			continue // not a .secpkg file
		}
		// load .secpkg file
		log.Printf(".secdep: load '%s'", fi.Name())
		pkg, err := Load(filepath.Join(secdepDir, fi.Name()))
		if err != nil {
			return false, err
		}
		// check for cycles
		if visited[pkg.Name] {
			return false, fmt.Errorf("secpkg: dependency cycle detected for package '%s'",
				pkg.Name)
		}
		// check if it is already installed
		pkgDir := filepath.Join(homedir.SecPkg(), "pkgs", pkg.Name)
		exists, err := file.Exists(pkgDir)
		if err != nil {
			return false, err
		}
		visited[pkg.Name] = true
		if !exists {
			// not installled
			log.Printf(".secdep: package '%s' not installed\n", pkg.Name)
			needsUpdate = true
		} else {
			// parse head
			h, err := hex.Decode(pkg.Head, 32)
			if err != nil {
				return false, err
			}
			var head [32]byte
			copy(head[:], h)
			// update
			log.Printf(".secdep: check update for package '%s'\n", pkg.Name)
			update, err := checkUpdate(ctx, visited, pkg.Name)
			if err != nil {
				return false, err
			}
			if update {
				needsUpdate = true
			}
		}
		delete(visited, pkg.Name)
	}

	return needsUpdate, nil
}
