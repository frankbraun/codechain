package tree

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/frankbraun/codechain/internal/hex"
	"github.com/frankbraun/codechain/util"
	"github.com/frankbraun/codechain/util/git"
)

// Sync treeDir to the state of treeHash with patches from patchDir.
// Prints status info if verbose is true.
func Sync(
	treeDir, targetHash, patchDir string,
	treeHashes []string,
	verbose bool,
	excludePaths []string,
	canRemoveDir bool,
) error {
	// argument checking
	if treeHashes[0] != EmptyHash {
		return fmt.Errorf("tree: treeHashes doesn't start with EmptyHash")
	}
	if !util.ContainsString(treeHashes, targetHash) {
		return fmt.Errorf("tree: targetHash unknown: %s", targetHash)
	}

	hash, err := Hash(treeDir, excludePaths)
	if err != nil {
		return err
	}
	hashStr := hex.Encode(hash[:])
	if verbose {
		fmt.Printf("treeDir    : %s\n", treeDir)
		fmt.Printf("treeDirHash: %x\n", hash[:])
		fmt.Printf("targetHash : %s\n", targetHash)
	}

	if hashStr == targetHash {
		if verbose {
			fmt.Println("treeDir in sync")
		}
		return nil
	}

	// find target hash index
	var idx int
	for ; idx < len(treeHashes); idx++ {
		if treeHashes[idx] == targetHash {
			break
		}
	}

	// find start position
	var i int
	for ; i < idx; i++ {
		if hashStr == treeHashes[i] {
			break
		}
	}
	if i == idx {
		if !canRemoveDir {
			return errors.New("tree: could not find a valid start to apply, try with empty dir")
		}
		if err := os.RemoveAll(treeDir); err != nil {
			return err
		}
		if err := os.Mkdir(treeDir, 0755); err != nil {
			return err
		}
	}

	for _, h := range treeHashes {
		if verbose {
			fmt.Printf("apply patch: %s\n", h)
		}

		// verify previous patch
		p, err := Hash(treeDir, excludePaths)
		if err != nil {
			return err
		}
		if hex.Encode(p[:]) != h {
			return fmt.Errorf("tree: patch failed to create target: %s", h)
		}

		// check if we are done
		if h == targetHash {
			break
		}

		// open patch file
		patch, err := os.Open(filepath.Join(patchDir, h))
		if err != nil {
			return err
		}

		// apply patch
		if verbose {
			fmt.Println("applying patch")
		}
		err = git.Apply(patch, 4, treeDir, false)
		if err != nil {
			patch.Close()
			return err
		}
		patch.Close()
	}
	return nil
}
