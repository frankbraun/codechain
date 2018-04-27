// Package command implements the Codechain commands.
package command

import (
	"path/filepath"
)

const (
	// CodechainDir is the default directory used to store Codechain data.
	CodechainDir = ".codechain"
)

var hashchainFile = filepath.Join(CodechainDir, "hashchain")
