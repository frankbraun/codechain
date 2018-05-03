package tree

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/frankbraun/codechain/internal/hex"
	"github.com/frankbraun/codechain/util"
	"github.com/frankbraun/codechain/util/git"
)

// Sync treeDir to the state of treeHash with patches from patchDir.
// Prints status info if verbose is true.
func Sync(treeDir, targetHash, patchDir string, treeHashes []string, verbose bool, excludePaths []string) error {
	fmt.Println("Sync()")
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
		fmt.Printf("treeDir   : %x\n", hash[:])
		fmt.Printf("targetHash: %s\n", targetHash)
	}

	if hashStr == targetHash {
		fmt.Println("treeDir in sync")
		return nil
	}

	if err := os.RemoveAll(treeDir); err != nil {
		return err
	}
	if err := os.Mkdir(treeDir, 0755); err != nil {
		return err
	}

	prev := EmptyHash
	fmt.Println("iterate")
	for _, h := range treeHashes {
		fmt.Printf("h: %s\n", h)
		// verify previous patch
		p, err := Hash(treeDir, excludePaths)
		if err != nil {
			return err
		}
		if hex.Encode(p[:]) != prev {
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
		fmt.Println("apply")
		err = git.Apply(patch, 4, treeDir, false)
		if err != nil {
			patch.Close()
			return err
		}
		patch.Close()

		prev = h
	}
	if treeHashes[len(treeHashes)-1] == targetHash {
		fmt.Println("check")
		fmt.Println(treeDir)
		// targetHash was last hash, verify it
		p, err := Hash(treeDir, excludePaths)
		if err != nil {
			return err
		}
		fmt.Println(hex.Encode(p[:]))
		if hex.Encode(p[:]) != targetHash {
			return fmt.Errorf("tree: patch failed to create target: %s", targetHash)
		}
	}
	return nil
}
