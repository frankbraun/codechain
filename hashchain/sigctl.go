package hashchain

import (
	"fmt"
	"strconv"

	"github.com/frankbraun/codechain/util/time"
)

// SignatureControl adds a signature control entry to the hash chain.
func (c *HashChain) SignatureControl(m int) (string, error) {
	// check argument
	if m <= 0 {
		return "", ErrSignatureThresholdNonPositive
	}
	if m > c.n {
		return "", ErrMLargerThanN
	}

	// create entry
	l := &link{
		previous:   c.LastEntryHash(),
		datum:      time.Now(),
		linkType:   signatureControlType,
		typeFields: []string{strconv.Itoa(m)},
	}
	c.chain = append(c.chain, l)

	// verify
	if err := c.verify(); err != nil {
		return "", err
	}

	// save
	entry := l.String()
	if _, err := fmt.Fprintln(c.fp, entry); err != nil {
		return "", err
	}
	return entry, nil
}
