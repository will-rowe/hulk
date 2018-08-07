package kmer

import (
	"testing"
)

var (
	k4seq  = []byte("TTTT")
	k8seq  = []byte("TCGATGCA")
	k32seq = []byte("TCGATGCATCGATGCATCGATGCATCGATGCA")
)

// test encoding a canonical 4-mer
func Test_encodeSeq(t *testing.T) {
	encodedSeq, err := EncodeSeq(k4seq, true)
	if err != nil {
		t.Fatal(err)
	}
	if encodedSeq != 0 {
		t.Fatal("encoding failed")
	}
}

// test encoding a sequence greater than k31
func Test_overflow(t *testing.T) {
	_, err := EncodeSeq(k32seq, true)
	t.Log(err)
	if err == nil {
		t.Fatal("kmer encoder should overflow due to k>31")
	}
}

// test decoding an ecoded k8 sequence
func Test_decodeSeq(t *testing.T) {
	encodedSeq, err := EncodeSeq(k8seq, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(encodedSeq)
	decodedSeq := DecodeSeq(encodedSeq, 8)
	if string(decodedSeq) != string(k8seq) {
		t.Fatal("decoded sequence does not match input sequence")
	}
}
