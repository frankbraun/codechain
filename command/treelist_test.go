package command

import (
	"testing"
)

func TestTreeList(t *testing.T) {
	err := TreeList("treelist")
	if err != nil {
		t.Errorf("TreeList() failed: %v", err)
	}
}
