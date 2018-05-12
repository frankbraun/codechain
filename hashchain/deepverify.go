package hashchain

import (
	"os"

	"github.com/frankbraun/codechain/sync"
	"github.com/frankbraun/codechain/util/file"
)

// DeepVerify hash chain. Use directory treeDir to apply patches from patchDir
// one after another and verify that they reflect the treehashes recorded in
// the hash chain.
func (c *HashChain) DeepVerify(treeDir, patchDir string, excludePaths []string) error {
	treeHashes := c.state.TreeHashes()

	// make sure treeDir exists
	if err := os.MkdirAll(treeDir, 0755); err != nil {
		return err
	}

	// remove treeDir contents first
	if err := file.RemoveAll(treeDir, excludePaths); err != nil {
		return err
	}

	// sync takes care of the rest
	targetHash := treeHashes[len(treeHashes)-1]
	return sync.Dir(treeDir, targetHash, patchDir, treeHashes, excludePaths, false)
}
