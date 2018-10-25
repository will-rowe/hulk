package histosketch

import (
	"testing"
)

var (
	epsilon = 0.001
	delta = 0.99
	decay   = 1.0
)

// test the count-min sketch constructor
func Test_NewCountMinSketch(t *testing.T) {
	cms := NewCountMinSketch(epsilon, delta, decay)
	t.Log(cms.Tables())
	t.Log(cms.Counters())
	if cms.Epsilon() != epsilon || cms.Delta() != delta {
		t.Fatal("size error in CMS constructor")
	}
}
