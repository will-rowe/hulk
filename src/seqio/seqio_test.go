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
	_, err := NewFASTQread(l1, l2, l3, l4)
	if err != nil {
		t.Fatalf("could not generate FASTQ read using NewFASTQread")
	}
	_, err = NewFASTQread(l1, l2[:len(l2)-2], l3, l4)
	if err == nil {
		t.Fatalf("bad FASTQ formatting now caught by NewFASTQread")
	}
	_, err = NewFASTQread(l1[1:], l2, l3, l4)
	if err == nil {
		t.Fatalf("bad FASTQ formatting now caught by NewFASTQread")
	}
}

func TestSeqMethods(t *testing.T) {
	_, err := NewFASTQread(l1, l2, l3, l4)
	if err != nil {
		t.Fatalf("could not generate FASTQ read using NewFASTQread")
	}
}
