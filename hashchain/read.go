package hashchain

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/lockfile"
	"github.com/frankbraun/codechain/util/time"
)

// Read hash chain from filename.
func Read(filename string) (*HashChain, error) {
	// check arguments
	exists, err := file.Exists(filename)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("hashchain: file '%s' doesn't exist", filename)
	}

	// init
	var c HashChain
	c.lock, err = lockfile.Create(filename)
	if err != nil {
		return nil, err
	}
	c.fp, err = os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	// read hash chain
	s := bufio.NewScanner(c.fp)
	for s.Scan() {
		// the parsing is very basic, the actual verification is done in c.verify()
		line := strings.SplitN(s.Text(), " ", 4)
		previous, err := hex.DecodeString(line[0])
		if err != nil {
			return nil, fmt.Errorf("hashchain: cannot decode hash '%s': %s", line[0], err)
		}
		if len(previous) != 32 {
			return nil, fmt.Errorf("hashchain: decoded hash has wrong length '%s': %s", line[0], err)
		}
		var prev [32]byte
		copy(prev[:], previous)
		t, err := time.Parse(line[1])
		if err != nil {
			return nil, fmt.Errorf("hashchain: cannot parse time '%s': %s", line[1], err)
		}
		l := &link{
			previous:   prev,
			datum:      t,
			linkType:   line[2],
			typeFields: strings.SplitN(line[3], " ", -1),
		}
		c.chain = append(c.chain, l)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	// verify
	if err := c.verify(); err != nil {
		return nil, err
	}
	return &c, nil
}
