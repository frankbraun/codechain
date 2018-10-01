package archive

import (
	"fmt"
	"os"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/frankbraun/codechain/internal/def"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/log"
)

// CreateDist creates a distribution file with filename for hash chain c.
// Filename must not exist.
func CreateDist(c *hashchain.HashChain, filename string) error {
	exists, err := file.Exists(filename)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("distribution file '%s' exists already", filename)
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Printf("creating distribution '%s'", filename)
	return Create(f, c, def.PatchDir)
}
