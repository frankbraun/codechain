package hashchain

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/frankbraun/codechain/util/file"
	"github.com/frankbraun/codechain/util/hex"
	"github.com/frankbraun/codechain/util/lockfile"
	"github.com/frankbraun/codechain/util/log"
	"github.com/frankbraun/codechain/util/time"
)

func (c *HashChain) read(r io.Reader) error {
	// read hash chain
	s := bufio.NewScanner(r)
	for s.Scan() {
		// the parsing is very basic, the actual verification is done in c.verify()
		text := s.Text()
		log.Println(text)
		line := strings.SplitN(text, " ", 4)
		if len(line) != 4 {
			return fmt.Errorf("could not split into 4 space separated parts: %s", line)
		}
		previous, err := hex.Decode(line[0], 32)
		if err != nil {
			return err
		}
		var prev [32]byte
		copy(prev[:], previous)
		t, err := time.Parse(line[1])
		if err != nil {
			return fmt.Errorf("hashchain: cannot parse time '%s': %s", line[1], err)
		}
		l := &link{
			previous:   prev,
			datum:      t,
			linkType:   line[2],
			typeFields: strings.SplitN(line[3], " ", 4),
		}
		if l.String() != text {
			return fmt.Errorf("hashchain: cannot reproduce line:\n%s", text)
		}
		c.chain = append(c.chain, l)
	}
	if err := s.Err(); err != nil {
		return err
	}

	// verify
	return c.verify()
}

// Read hash chain from r and verify it.
func Read(r io.Reader) (*HashChain, error) {
	log.Printf("hashchain.Read()")
	var c HashChain
	if err := c.read(r); err != nil {
		return nil, err
	}
	return &c, nil
}

// ReadFile reads hash chain from filename and verifies it.
func ReadFile(filename string) (*HashChain, error) {
	log.Printf("hashchain.ReadFile(%s)", filename)
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

	if err := c.read(c.fp); err != nil {
		c.Close()
		return nil, err
	}

	// having only one signer is VERY BAD NEWS, emit obnoxious warning here, so
	// all tools will display it
	if c.M() == 1 {
		fmt.Fprintf(os.Stderr, "WARNING: this Codechain can be updated by only 1 signer!\n")
	}

	return &c, nil
}
