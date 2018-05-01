package tree

import (
	"errors"
	"fmt"

	"github.com/frankbraun/codechain/internal/hex"
)

// Sync treeDir to the state of treeHash with patches from patchDir.
// Prints status info if verbose is true.
func Sync(treeDir, treeHash, patchDir string, verbose bool, excludePaths []string) error {
	hash, err := Hash(treeDir, excludePaths)
	if err != nil {
		return err
	}
	hashStr := hex.Encode(hash[:])
	if verbose {
		fmt.Printf("treeDir : %x\n", hash[:])
		fmt.Printf("treeHash: %s\n", treeHash)
	}

	if hashStr == treeHash {
		fmt.Println("treeDir in sync")
		return nil
	}

	// TODO
	return errors.New("not implemented")
}
