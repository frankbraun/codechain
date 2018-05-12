package hashchain

import (
	"os"

	"github.com/frankbraun/codechain/sync"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/log"
)

// DeepVerify hash chain. Use directory treeDir to apply patches from patchDir
// one after another and verify that they reflect the treehashes recorded in
// the hash chain.
func (c *HashChain) DeepVerify(treeDir, patchDir string, excludePaths []string) error {
	treeHashes := c.state.TreeHashes()

	// remove treeDir first
	log.Printf("rm -rf %s", treeDir)
	if err := file.RemoveAll(treeDir, excludePaths); err != nil {
		return err
	}
	log.Printf("mkdir -p %s", treeDir)
	if err := os.MkdirAll(treeDir, 0755); err != nil {
		return err
	}

	// sync takes care of the rest
	targetHash := treeHashes[len(treeHashes)-1]
	return sync.Dir(treeDir, targetHash, patchDir, treeHashes, excludePaths, false)
}
