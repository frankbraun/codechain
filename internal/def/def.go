// Package def defines default values used in Codechain.
package def

import (
	"path/filepath"
)

// CodechainDir is the default directory used for Codechain related files.
const CodechainDir = ".codechain"

// ExcludePaths is the default list of paths not considered by Codechain.
var ExcludePaths = []string{
	CodechainDir,
	".git",
	".gitignore",
	".travis.yml",
}

// HashchainFile is the default name of the hashchain file.
var HashchainFile = filepath.Join(CodechainDir, "hashchain")

// PatchDir is the default name of the patch file directory.
var PatchDir = filepath.Join(CodechainDir, "patches")
