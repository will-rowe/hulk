// kmer contains the functions to encode sequences ([]byte) to integers (uint64)
package kmer

import (
	"bytes"
	"fmt"
)

// EncodeSeq converts a sequence (slice of bytes) to an integer (Uint64). Ns are stored as As. If canonical is true, the func will return the canonical encoded kmer
func EncodeSeq(seq []byte, canonical bool) (uint64, error) {
	k := len(seq)
	// can only store a max K of 31 as a Uint64 encoded kmer
	if k < 1 || k > 31 {
		return 0, fmt.Errorf("sequences must be < 32 bases, yours was %d!\n", k)
	}
	// send the seq to lower case
	bytes.ToLower(seq)
	// encode the sequence
	var encoded uint64
	for _, base := range seq {
		encoded = (encoded << 2) | uint64(base2int(base))
	}
	// if canonical requested, encode the RC and keep the lexicographically smallest value
	if canonical == true {
		rcEncoded := rcEncode(encoded, k)
		if rcEncoded < encoded {
			encoded = rcEncoded
		}
	}
	return encoded, nil
}

// DecodeSeq converts an integer encoded sequence back to a byte slice
func DecodeSeq(seq uint64, kSize int) []byte {
	decodedSeq := make([]byte, kSize)
	for i := 0; i < kSize; i++ {
		decodedSeq[kSize-i-1] = "ACGT"[(byte(seq & 0x3))]
		seq >>= 2
	}
	return decodedSeq
}

// base2int converts a base to an index (0-3)
func base2int(base byte) byte {
	switch base {
	case 'A':
		return 0
	case 'C':
		return 1
	case 'G':
		return 2
	case 'T':
		return 3
	case 'N':
		return 0
	default:
		//panic(fmt.Errorf("non a/c/t/g/n base: %c", b))
		return 0
	}
}

// rcEncode returns the reverse complement of a uint64 encoded sequence
func rcEncode(seq uint64, k int) uint64 {
	var rc uint64
	for i := 0; i < k; i++ {
		rc <<= 2
		rc |= 3 &^ seq
		seq >>= 2
	}
	return rc
}
