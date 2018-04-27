// Package command implements the Codechain commands.
package command

import (
	"path/filepath"
)

const codechainDir = ".codechain"

var hashchainFile = filepath.Join(codechainDir, "hashchain")

var excludePaths = []string{
	codechainDir,
	".git",
	".gitignore",
}
