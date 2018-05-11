package hashchain

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/frankbraun/codechain/internal/hex"
	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/lockfile"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/time"
)

// Read hash chain from filename and verify it.
func Read(filename string) (*HashChain, error) {
	log.Printf("hashchain.Read(%s)", filename)
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
		c.lock.Release()
		return nil, err
	}

	// read hash chain
	s := bufio.NewScanner(c.fp)
	for s.Scan() {
		// the parsing is very basic, the actual verification is done in c.verify()
		text := s.Text()
		log.Println(text)
		line := strings.SplitN(text, " ", 4)
		previous, err := hex.Decode(line[0], 32)
		if err != nil {
			c.lock.Release()
			return nil, err
		}
		var prev [32]byte
		copy(prev[:], previous)
		t, err := time.Parse(line[1])
		if err != nil {
			c.lock.Release()
			return nil, fmt.Errorf("hashchain: cannot parse time '%s': %s", line[1], err)
		}
		l := &link{
			previous:   prev,
			datum:      t,
			linkType:   line[2],
			typeFields: strings.SplitN(line[3], " ", 4),
		}
		if l.String() != text {
			c.lock.Release()
			return nil, fmt.Errorf("hashchain: cannot reproduce line:\n%s", text)
		}
		c.chain = append(c.chain, l)
	}
	if err := s.Err(); err != nil {
		c.lock.Release()
		return nil, err
	}

	// verify
	if err := c.verify(); err != nil {
		c.lock.Release()
		return nil, err
	}
	return &c, nil
}
