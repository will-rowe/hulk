package bitvector

import (
	"testing"
)

var (
	testLimit   = 142
	testNumber  = 19
	testNumberB = 200
	testNumberC = 17
)

// test the constructor function
func TestNewBitVector(t *testing.T) {
	// get the max size
	MS := maxSize()
	// create a bit vector
	bv := NewBitVector(testLimit)
	if len(bv) != (testLimit/MS + 1) {
		t.Error("bitvector should contain 3 int(64) to hold 142")
	}
	t.Log(bv)
}

// test adding an element to a bit vector
func TestAdd(t *testing.T) {
	bv := NewBitVector(testLimit)
	// add something within bit vector limit
	if err := bv.Add(testNumber); err != nil {
		t.Fatal(err)
	}
	// check this went into the first bit bin
	if bv[0] == 0 {
		t.Fatal("bit vector didn't add value")
	}
	// check the other bins are empty still
	if bv[1] != 0 || bv[2] != 0 {
		t.Fatal("bit vector populated the incorrect bit bins")
	}
	// add something too large and check for error
	if err := bv.Add(testNumberB); err == nil {
		t.Fatal("bitvector should have prevented overflow")
	}
}

// test for checking an element is in a bit vector
func TestContains(t *testing.T) {
	bv := NewBitVector(testLimit)
	if ok := bv.Contains(testNumber); ok {
		t.Fatal("should return false as it has not been added yet")
	}
	if err := bv.Add(testNumber); err != nil {
		t.Fatal(err)
	}
	if ok := bv.Contains(testNumber); !ok {
		t.Fatal("should return true as it has already been added")
	}
	// add second number
	if ok := bv.Contains(testNumberC); ok {
		t.Fatal("should return false as it has not been added yet")
	}
	if err := bv.Add(testNumberC); err != nil {
		t.Fatal(err)
	}
	if ok := bv.Contains(testNumberC); !ok {
		t.Fatal("should return true as it has already been added")
	}
	if ok := bv.Contains(testNumber); !ok {
		t.Fatal("should return true as it has already been added")
	}
}

// test bitwise and of bit vectors
func TestBitWiseAND(t *testing.T) {
	bv1 := NewBitVector(testLimit)
	bv2 := NewBitVector(testLimit)
	bv3 := NewBitVector(testLimit * testLimit)
	if _, err := bv1.BWAND(bv3); err == nil {
		t.Fatal("should stop mismatched bit vectors being compared")
	}
	_ = bv1.Add(testNumber)
	_ = bv2.Add(testNumber)
	// with same number in each bit vector, the bitwise AND should result in the same value
	result, _ := bv1.BWAND(bv2)
	if result[0] != bv1[0] || result[0] != bv2[0] {
		t.Fatal("unexpeected bitwise AND result")
	}
	// add a new value to one of the vectors, then check only the shared value is returned in the result
	_ = bv1.Add(testNumberC)
	result, _ = bv1.BWAND(bv2)
	if ok := result.Contains(testNumber); !ok {
		t.Fatal("bitwise AND result should contain the shared number")
	}
	if ok := result.Contains(testNumberC); ok {
		t.Fatal("bitwise AND result should NOT contain the unique number")
	}
	t.Log(bv1, bv2, result)
}

// test popcount method
func TestPopCount(t *testing.T) {
	bv := NewBitVector(testLimit)
	_ = bv.Add(testNumber)
	if i := bv.PopCount(); i != 1 {
		t.Fatal("pop count failed to count flipped bits in bit vector")
	}
	_ = bv.Add(testNumberC)
	if i := bv.PopCount(); i != 2 {
		t.Fatal("pop count failed to count flipped bits in bit vector")
	}
}