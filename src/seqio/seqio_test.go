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

func TestPop(t *testing.T) {
	read, err := NewFastqRead(l1, l2, l3, l4)
	if err != nil {
		t.Fatalf("could not generate FASTQ read using NewFastqRead")
	}
	t.Log(string(read.Seq()))
	// pop the first 10 bases from the read
	err = read.Pop(10)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(read.Seq()))
	if read.Length() != len(l2)-10 {
		t.Fatal("pop did not work")
	}
}

func TestShred(t *testing.T) {
	read, err := NewFastqRead(l1, l2, l3, l4)
	if err != nil {
		t.Fatalf("could not generate FASTQ read using NewFastqRead")
	}
	// get 10 base pair chunks from the original read length
	var finalChunk []byte
	shredChan, err := read.Shred(10)
	if err != nil {
		t.Fatal(err)
	}
	for chunk := range shredChan {
		finalChunk = chunk
	}
	if string(finalChunk) != string(l2[90:]) {
		t.Fatal("shredding has not worked")
	}
	if _, err := read.Shred(200); err == nil {
		t.Fatal("chunk size is longer than read, shredding should fail")
	}

}
