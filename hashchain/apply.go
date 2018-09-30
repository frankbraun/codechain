package hashchain

import (
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/sync"
)

// Apply to current working directory and check head if not nil.
func (c *HashChain) Apply(head *[32]byte) error {
	targetHash, _ := c.LastSignedTreeHash()
	treeHashes := c.TreeHashes()
	if head != nil {
		if err := c.CheckHead(*head); err != nil {
			return err
		}
	}
	err := sync.Dir(".", targetHash, def.PatchDir, treeHashes, def.ExcludePaths, false)
	if err != nil {
		return err
	}
	return nil
}
