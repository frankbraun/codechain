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
	treeDir       = filepath.Join(codechainDir, "tree")
	patchDir      = filepath.Join(codechainDir, "patches")
	excludePaths  = []string{
		codechainDir,
		".git",
		".gitignore",
		".travis.yml",
	}
)
