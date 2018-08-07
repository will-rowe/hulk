package histosketch

import (
	"testing"
)

var (
	EPSILON = 0.001
	DELTA   = 0.99
	DR      = 1.0
)

// test the count-min sketch constructor
func Test_NewCountMinSketch(t *testing.T) {
	cms := NewCountMinSketch(EPSILON, DELTA, DR)
	t.Log(cms.Tables())
	t.Log(cms.Counters())
	if cms.Epsilon() != EPSILON {
		t.Fatal("epsilon error in CMS constructor")
	}
	if cms.Delta() != DELTA {
		t.Fatal("delta error in CMS constructor")
	}
}
