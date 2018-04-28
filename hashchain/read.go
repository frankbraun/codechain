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
	var c HashChain
	exists, err := file.Exists(filename)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("hashchain: file '%s' doesn't exist", filename)
	}
	c.lock, err = lockfile.Create(filename)
	if err != nil {
		return nil, err
	}
	c.fp, err = os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	s := bufio.NewScanner(c.fp)
	for s.Scan() {
		line := strings.SplitN(s.Text(), " ", 4)
		previous, err := hex.DecodeString(line[0])
		if err != nil {
			return nil, fmt.Errorf("hashchain: cannot decode hash '%s': %s", line[0], err)
		}
		t, err := time.Parse(line[1])
		if err != nil {
			return nil, fmt.Errorf("hashchain: cannot parse time '%s': %s", line[1], err)
		}
		l := &link{
			previous:   previous,
			datum:      t,
			linkType:   line[2],
			typeFields: strings.SplitN(line[3], " ", -1),
		}
		c.chain = append(c.chain, l)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	c.m = 1
	if err := c.verify(); err != nil {
		return nil, err
	}
	return &c, nil
}
