/*
	tests for the seqio package
*/
package seqio

import (
	"testing"
)

// setup variables
var (
	l1 = []byte("@0_chr1_0_186027_186126_263_(Bla)BIC-1:GQ260093:1-885:885")
	l2 = []byte("acagcaggaaggcttactggagaaacgtatcgactataagaatcgggtgatggaacctcactctcccatcagcgcacaacatagttcgacgggtatgacc")
	l3 = []byte("+")
	l4 = []byte("====@==@AAD?>D@@==DACBC?@BB@C==AB==A@D>AD==?CB==@=B?=A>D?=DB=?>>D@EB===??=@C=?C>@>@B>=?C@@>=====?@>=")
)

// test functions to check equality of slices
func ByteSliceCheck(a, b []byte) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func Uint64SliceCheck(a, b []uint64) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// begin the tests
func TestReadConstructor(t *testing.T) {
	_, err := NewFastqRead(l1, l2, l3, l4)
	if err != nil {
		t.Fatalf("could not generate FASTQ read using NewFastqRead")
	}
	_, err = NewFastqRead(l1, l2[:len(l2)-2], l3, l4)
	if err == nil {
		t.Fatalf("bad FASTQ formatting now caught by NewFastqRead")
	}
	_, err = NewFastqRead(l1[1:], l2, l3, l4)
	if err == nil {
		t.Fatalf("bad FASTQ formatting now caught by NewFastqRead")
	}
}

func TestShred(t *testing.T) {
	read, err := NewFastqRead(l1, l2, l3, l4)
	if err != nil {
		t.Fatalf("could not generate FASTQ read using NewFastqRead")
	}
	// get chunks that are 10% of the original read length
	var finalChunk []byte
	for chunk := range read.Shred(0.1) {
		finalChunk = chunk
	}
	if string(finalChunk) != string(l2[90:]) {
		t.Fatal("shredding has not worked")
	}
}
