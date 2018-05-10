// Package command implements the Codechain commands.
package command

import (
	"path/filepath"

	"github.com/frankbraun/codechain/internal/def"
)

var (
	hashchainFile = filepath.Join(def.CodechainDir, "hashchain")
	treeDirRoot   = filepath.Join(def.CodechainDir, "tree")
	treeDirA      = filepath.Join(treeDirRoot, "a")
	treeDirB      = filepath.Join(treeDirRoot, "b")
	patchDir      = filepath.Join(def.CodechainDir, "patches")
)
