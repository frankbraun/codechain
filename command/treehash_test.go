package command

import (
	"testing"
)

func TestTreeHash(t *testing.T) {
	err := TreeHash("treehash")
	if err != nil {
		t.Errorf("TreeHash() failed: %v", err)
	}
}
