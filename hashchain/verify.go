package hashchain

import (
	"bytes"
	"crypto/sha256"
)

func (c *HashChain) verifyChainStartType(i int, fields []string) error {
	if i != 0 {
		return ErrIllegalCStart
	}
	return nil
}

func (c *HashChain) verifySourceType(i int, fields []string) error {
	if i == 0 {
		return ErrMustStartWithCStart
	}
	// TODO
	return nil
}

func (c *HashChain) verifySignatureType(i int, fields []string) error {
	if i == 0 {
		return ErrMustStartWithCStart
	}
	// TODO
	return nil
}

func (c *HashChain) verifyAddKeyType(i int, fields []string) error {
	if i == 0 {
		return ErrMustStartWithCStart
	}
	// TODO
	return nil
}

func (c *HashChain) verifyRemoveKeyType(i int, fields []string) error {
	if i == 0 {
		return ErrMustStartWithCStart
	}
	// TODO
	return nil
}

func (c *HashChain) verifySignatureControlType(i int, fields []string) error {
	if i == 0 {
		return ErrMustStartWithCStart
	}
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
			err = c.verifySourceType(i, l.typeFields)
		case signatureType:
			err = c.verifySignatureType(i, l.typeFields)
		case addKeyType:
			err = c.verifyAddKeyType(i, l.typeFields)
		case removeKeyType:
			err = c.verifyRemoveKeyType(i, l.typeFields)
		case signatureControlType:
			err = c.verifySignatureControlType(i, l.typeFields)
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
