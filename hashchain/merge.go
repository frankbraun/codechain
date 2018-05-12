package hashchain

import (
	"fmt"
)

// Merge hashchain src into c.
func (c *HashChain) Merge(src *HashChain) error {
	i := 0
	for ; i < len(c.chain) && i < len(src.chain); i++ {
		if !linkEqual(c.chain[i], src.chain[i]) {
			return ErrCannotMerge
		}
	}
	if len(src.chain) < len(c.chain) {
		return ErrNothingToMerge
	}
	for ; i < len(src.chain); i++ {
		var l link
		l = *src.chain[i]
		c.chain = append(c.chain, &l)
		// Verifiying the entire chain after every entry is a bit excessive,
		// especially because src is already verified.
		// But better be safe than sorry.
		if err := c.verify(); err != nil {
			return err
		}
		// save
		if _, err := fmt.Fprintln(c.fp, l.String()); err != nil {
			return err
		}
	}
	return nil
}
