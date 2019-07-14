package minhash

import (
	"testing"
)

var (
	kmerSize   = uint(7)
	sketchSize = uint(10)
	sequence   = []byte("ACTGCGTGCGTGAAACGTGCACGTGACGTG")
	sequence2  = []byte("TGACGCACGCACTTTGCACGTGCACTGCAC")
	hashvalues = []uint64{12345, 54321, 9999999, 98765}
	hashvalues2 = []uint64{12345, 54321, 111111, 222222}
)

func TestMinHashConstructors(t *testing.T) {
	mhKMV := NewKMVsketch(kmerSize, sketchSize)
	if mhKMV.SketchSize != sketchSize || mhKMV.KmerSize != kmerSize {
		t.Fatalf("NewKMVsketch constructor did not initiate MinHash KMV sketch correctly")
	}

}

func TestSimilarityEstimates(t *testing.T) {
	// test KMV
	mhKMV1 := NewKMVsketch(kmerSize, sketchSize)
	for _, hash := range hashvalues {
		mhKMV1.AddHash(hash)
	}
	mhKMV2 := NewKMVsketch(kmerSize, sketchSize)
	for _, hash := range hashvalues {
		mhKMV2.AddHash(hash)
	}
	if js, err := mhKMV1.GetSimilarity(mhKMV2); err != nil {
		t.Fatal(err)
		if js != 0.5 {
			t.Fatalf("incorrect similarity estimate: %f", js)
		}
	}
}

// benchmark KMV
func BenchmarkKMV(b *testing.B) {
	mhKMV1 := NewKMVsketch(kmerSize, sketchSize)
	// run the add method b.N times
	for n := 0; n < b.N; n++ {
		for _, hash := range hashvalues {
			mhKMV1.AddHash(hash)
		}
	}
}
