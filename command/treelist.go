package command

import (
	"os"

	"github.com/frankbraun/codechain/tree"
)

// TreeList implements the 'treelist' command.
func TreeList() error {
	list, err := tree.List(".", excludePaths)
	if err != nil {
		return err
	}
	os.Stdout.Write(list)
	return nil
}
