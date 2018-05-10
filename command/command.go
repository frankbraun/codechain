// Package command implements the Codechain commands.
package command

import (
	"path/filepath"
)

const (
	codechainDir = ".codechain"
)

var (
	hashchainFile = filepath.Join(codechainDir, "hashchain")
	treeDirRoot   = filepath.Join(codechainDir, "tree")
	treeDirA      = filepath.Join(treeDirRoot, "a")
	treeDirB      = filepath.Join(treeDirRoot, "b")
	patchDir      = filepath.Join(codechainDir, "patches")
	// ExcludePaths is the default list of paths not considered by Codechain.
	ExcludePaths = []string{
		codechainDir,
		".git",
		".gitignore",
		".travis.yml",
	}
)
