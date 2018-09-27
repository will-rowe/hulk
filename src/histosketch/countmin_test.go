package histosketch

import (
	"testing"
)

var (
	SIZE = uint(10)
	DR = 1.0
)

// test the count-min sketch constructor
func Test_NewCountMinSketch(t *testing.T) {
	cms := NewCountMinSketch(SIZE, DR)
	t.Log(cms.Tables())
	t.Log(cms.Counters())
	if cms.SizeMB() != SIZE {
		t.Fatal("size error in CMS constructor")
	}
}
