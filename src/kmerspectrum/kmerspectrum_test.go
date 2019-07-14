package kmerspectrum

import (
	"testing"
)

var (
	numBins = int32(10)
	hv1     = uint64(1)
	hv2     = uint64(1234)
)

// test the constructor function
func TestNewSpectrum(t *testing.T) {
	ks, err := NewKmerSpectrum(-1)
	if err == nil {
		t.Fatal("shouldn't accept negative numBins")
	}
	ks, err = NewKmerSpectrum(numBins)
	if err != nil {
		t.Fatal(err)
	}
	if int32(ks.Size()) != numBins || len(ks.bins) != int(numBins) {
		t.Fatal("failed to make spectrum with correct number of bins")
	}
}

// test the AddHash and Cardinality methods
func TestAddHash(t *testing.T) {
	ks, err := NewKmerSpectrum(numBins)
	if err != nil {
		t.Fatal(err)
	}
	if ks.Cardinality() != 0 {
		t.Fatal("cardinality should be 0 when spectrum is empty")
	}
	ks.AddHash(hv1)
	if ks.Cardinality() != 1 {
		t.Fatalf("incorrect cardinality - should be 1, not %d", ks.Cardinality())
	}
	ks.AddHash(hv2)
	if ks.Cardinality() != 2 {
		t.Fatalf("incorrect cardinality - should be 2, not %d", ks.Cardinality())
	}
}
