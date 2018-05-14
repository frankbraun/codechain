package hashchain

import (
	"fmt"
	"strconv"

	"github.com/frankbraun/codechain/hashchain/linktype"
	"github.com/frankbraun/codechain/util/time"
)

// SignatureControl adds a signature control entry to the hash chain.
func (c *HashChain) SignatureControl(m int) (string, error) {
	// check argument
	if m <= 0 {
		return "", ErrSignatureThresholdNonPositive
	}
	if m > c.state.HeadN() {
		return "", ErrMLargerThanN
	}

	// create entry
	l := &link{
		previous:   c.Head(),
		datum:      time.Now(),
		linkType:   linktype.SignatureControl,
		typeFields: []string{strconv.Itoa(m)},
	}
	c.chain = append(c.chain, l)

	// verify
	if err := c.verify(); err != nil {
		return "", err
	}

	// save
	if _, err := fmt.Fprintln(c.fp, l.String()); err != nil {
		return "", err
	}
	return l.StringColor(), nil
}
