// Package command implements the Codechain commands.
package command

import (
	"path/filepath"

	"github.com/frankbraun/codechain/internal/def"
)

var (
	treeDirRoot = filepath.Join(def.CodechainDir, "tree")
	treeDirA    = filepath.Join(treeDirRoot, "a")
	treeDirB    = filepath.Join(treeDirRoot, "b")
)
