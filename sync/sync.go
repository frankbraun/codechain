// Package sync implements directory syncing.
package sync

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/internal/hex"
	"github.com/frankbraun/codechain/patchfile"
	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util"
	"github.com/frankbraun/codechain/util/log"
)

// Dir syncs treeDir to the state of treeHash with patches from patchDir.
func Dir(
	treeDir, targetHash, patchDir string,
	treeHashes []string,
	excludePaths []string,
	canRemoveDir bool,
) error {
	// argument checking
	if treeHashes[0] != tree.EmptyHash {
		return fmt.Errorf("tree: treeHashes doesn't start with EmptyHash")
	}
	if !util.ContainsString(treeHashes, targetHash) {
		return fmt.Errorf("tree: targetHash unknown: %s", targetHash)
	}

	hash, err := tree.Hash(treeDir, excludePaths)
	if err != nil {
		return err
	}
	hashStr := hex.Encode(hash[:])
	log.Printf("treeDir    : %s\n", treeDir)
	log.Printf("treeDirHash: %x\n", hash[:])
	log.Printf("targetHash : %s\n", targetHash)

	if hashStr == targetHash {
		log.Println("treeDir in sync")
		return nil
	}

	// find target hash index
	var idx int
	for ; idx < len(treeHashes); idx++ {
		if treeHashes[idx] == targetHash {
			break
		}
	}
	if idx == len(treeHashes) {
		return fmt.Errorf("tree: could not find target hash: %s", targetHash)
	}

	// find start position
	var i int
	for ; i < idx; i++ {
		if hashStr == treeHashes[i] {
			log.Printf("start position %d found", i)
			break
		}
	}
	if i == idx {
		if !canRemoveDir {
			return errors.New("tree: could not find a valid start to apply, try with empty dir")
		}
		log.Println("could not find a valid start to apply, trying with empty dir...")
		if err := os.RemoveAll(treeDir); err != nil {
			return err
		}
		if err := os.Mkdir(treeDir, 0755); err != nil {
			return err
		}
	}

	for ; i <= idx; i++ {
		h := treeHashes[i]
		log.Printf("apply patch: %s\n", h)

		// verify previous patch
		p, err := tree.Hash(treeDir, excludePaths)
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
		log.Println("applying patch")
		err = patchfile.Apply(treeDir, patch, def.ExcludePaths)
		if err != nil {
			patch.Close()
			return err
		}
		patch.Close()
	}
	return nil
}
