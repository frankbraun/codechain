package hashchain

import (
	"bytes"
	"crypto/sha256"
)

func (c *HashChain) verifyChainStartType(i int, fields []string) error {
	// TODO
	return nil
}

func (c *HashChain) verifySourceType(fields []string) error {
	// TODO
	return nil
}

func (c *HashChain) verifySignatureType(fields []string) error {
	// TODO
	return nil
}

func (c *HashChain) verifyAddKeyType(fields []string) error {
	// TODO
	return nil
}

func (c *HashChain) verifyRemoveKeyType(fields []string) error {
	// TODO
	return nil
}

func (c *HashChain) verifySignatureControlType(fields []string) error {
	// TODO
	return nil
}

// verify hash chain.
func (c *HashChain) verify() error {
	// basic check
	if len(c.chain) == 0 {
		return ErrEmpty
	}

	// set start values
	c.m = 1
	c.n = 1
	prevHash := emptyTree
	var prevDatum int64

	// iterate over all links
	for i, l := range c.chain {
		// make sure we actually have a hash chain
		if !bytes.Equal(prevHash[:], l.previous[:]) {
			return ErrLinkBroken
		}

		// make sure time is ascending
		if l.datum < prevDatum {
			return ErrDescendingTime
		}

		var err error
		switch l.linkType {
		case chainStartType:
			err = c.verifyChainStartType(i, l.typeFields)
		case sourceType:
			err = c.verifySourceType(l.typeFields)
		case signatureType:
			err = c.verifySignatureType(l.typeFields)
		case addKeyType:
			err = c.verifyAddKeyType(l.typeFields)
		case removeKeyType:
			err = c.verifyRemoveKeyType(l.typeFields)
		case signatureControlType:
			err = c.verifySignatureControlType(l.typeFields)
		default:
			err = ErrUnknownLinkType
		}
		if err != nil {
			return err
		}

		// prepare for next entry
		prevHash = sha256.Sum256([]byte(l.String()))
		prevDatum = l.datum
	}

	// all clear
	return nil
}
