package command

import (
	"fmt"

	"github.com/frankbraun/codechain/tree"
)

// TreeHash implements the 'treehash' command.
func TreeHash() error {
	hash, err := tree.Hash(".", excludePaths)
	if err != nil {
		return err
	}
	fmt.Printf("%x\n", hash[:])
	return nil
}
