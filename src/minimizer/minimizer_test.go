package minimizer

import (
	"testing"
)

var (
	actgn = []byte("ACGTN")
	seq   = []byte("ACTGAAAATTTT")
	k     = uint(4)
	w     = uint(4)
)

func TestTable(t *testing.T) {
	if seq_nt4_table[actgn[0]] != 0 {
		t.Fatalf("%v should == 0", string(actgn[0]))
	}
	if seq_nt4_table[actgn[1]] != 1 {
		t.Fatalf("%v should == 1", string(actgn[1]))
	}
	if seq_nt4_table[actgn[2]] != 2 {
		t.Fatalf("%v should == 2", string(actgn[2]))
	}
	if seq_nt4_table[actgn[3]] != 3 {
		t.Fatalf("%v should == 3", string(actgn[3]))
	}
	if seq_nt4_table[actgn[4]] != 4 {
		t.Fatalf("%v should == 4", string(actgn[4]))
	}
}

func TestSketching(t *testing.T) {
	sketcher, err := NewMinimizerSketch(k, w, seq)
	if err != nil {
		t.Fatal(err)
	}
	results := sketcher.GetMinimizers()
	sketch1 := make(map[uint64]struct{})
	for minimizer := range results {
		sketch1[minimizer.(uint64)] = struct{}{}
	}

	sketcher2, err := NewMinimizerSketch(k, w, seq)
	if err != nil {
		t.Fatal(err)
	}
	results2 := sketcher2.GetMinimizers()
	sketch2 := make(map[uint64]struct{})
	for minimizer := range results2 {
		sketch2[minimizer.(uint64)] = struct{}{}
	}
	if len(sketch1) != len(sketch2) {
		t.Fatal("sketch lengths don't match")
	}
	for minimizer := range sketch1 {
		if _, ok := sketch2[minimizer]; !ok {
			t.Fatal("sketches don't match")
		}
	}

}
