// Package sync implements directory tree syncing with patch files.
package sync

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/patchfile"
	"github.com/frankbraun/codechain/tree"
	"github.com/frankbraun/codechain/util"
	"github.com/frankbraun/codechain/util/hex"
	"github.com/frankbraun/codechain/util/log"
)

// Dir syncs the treeDir to the tree hash targetHash with patches from patchDir.
//
// treeHashes is a list of intermediate tree hashes which must start with
// tree.EmptyHash and contain the targetHash. Synchronization happens by
// applying one patch file after another, iterating through the treeHashes
// array until the targetHash is found.
//
// In order to find a suitable start, the tree hash of treeDir is calculated
// and treeHashes is searched for the result.
//
// If no suitable start can be found and canRemoveDir is true, all contents of
// treeDir are removed and the patches are applied starting from
// tree.EmptyHash. Otherwise, ErrCannotRemove is returned.
//
// Patch files (see patchfile package) are named after the outgoing (source)
// tree hash and must lead to the targetDir having the tree hash of the next
// treeHashes entry after they have been applied.
//
// The paths given in excludePaths are excluded from all tree hash calculations.
func Dir(
	treeDir, targetHash, patchDir string,
	treeHashes []string,
	excludePaths []string,
	canRemoveDir bool,
) error {
	// argument checking
	if treeHashes[0] != tree.EmptyHash {
		return fmt.Errorf("sync: treeHashes doesn't start with tree.EmptyHash")
	}
	if !util.ContainsString(treeHashes, targetHash) {
		return fmt.Errorf("sync: targetHash unknown: %s", targetHash)
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
		return fmt.Errorf("sync: could not find target hash: %s", targetHash)
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
			return ErrCannotRemove
		}
		log.Println("could not find a valid start to apply, trying with empty dir...")
		if err := os.RemoveAll(treeDir); err != nil {
			return err
		}
		if err := os.Mkdir(treeDir, 0755); err != nil {
			return err
		}
		i = 0
	}

	for ; i <= idx; i++ {
		h := treeHashes[i]

		// verify previous patch
		p, err := tree.Hash(treeDir, excludePaths)
		if err != nil {
			return err
		}
		if hex.Encode(p[:]) != h {
			return fmt.Errorf("sync: patch failed to create target: %s", h)
		}
		log.Printf("verified patch: %s\n", h)

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
		log.Printf("applying patch: %s\n", h)
		err = patchfile.Apply(treeDir, patch, def.ExcludePaths)
		if err != nil {
			patch.Close()
			return err
		}
		patch.Close()
	}
	return nil
}
