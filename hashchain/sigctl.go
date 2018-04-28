package hashchain

import (
	"fmt"
	"strconv"

	"github.com/frankbraun/codechain/util/time"
)

// SignatureControl adds a signature control entry to the hash chain.
func (c *HashChain) SignatureControl(m int) (string, error) {
	// TODO: check that we have enough keys to reach m.
	if m <= 0 {
		return "", ErrSignatureThresholdNonPositive
	}
	prev := c.LastEntryHash()
	l := &link{
		previous:   prev[:],
		datum:      time.Now(),
		linkType:   signatureControlType,
		typeFields: []string{strconv.Itoa(m)},
	}
	c.chain = append(c.chain, l)
	entry := l.String()
	if _, err := fmt.Fprintln(c.fp, entry); err != nil {
		return "", err
	}
	return entry, nil
}
