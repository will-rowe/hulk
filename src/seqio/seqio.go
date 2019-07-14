// Package seqio contains custom types and methods for holding and processing sequence data
package seqio

import (
	"fmt"
	"unicode"
)

// FASTQ_ENCODING used by the FASTQ file
const FASTQ_ENCODING = 33

// complementBases is used for reverse complementing
var complementBases = []byte{
	'A': 'T',
	'T': 'A',
	'C': 'G',
	'G': 'C',
	'N': 'N',
}

// Sequence is the base type for a FASTQ and FASTA read
type Sequence struct {
	ID  []byte
	Seq []byte
}

// FASTQread is a type that holds a single FASTQ read, along with the locations it mapped to
type FASTQread struct {
	Sequence
	Misc []byte
	Qual []byte
	RC   bool
}

// NewFASTQread is the constructor function, which takes 4 lines of a fastq entry and returns the FASTQread object
// TODO: this is garbage - needs replacing
func NewFASTQread(l1 []byte, l2 []byte, l3 []byte, l4 []byte) (*FASTQread, error) {
	if l1[0] != 64 {
		return nil, fmt.Errorf("read ID in fastq file does not begin with @: %v", string(l1))
	}

	// create a FASTQread struct
	return &FASTQread{
		Sequence: Sequence{ID: l1, Seq: l2},
		Misc:     l3,
		Qual:     l4,
	}, nil
}

// BaseCheck is a method to check for ACTGN bases and also to convert bases to upper case
// TODO: more garbage - not efficient and doesn't handle non actgn
func (Sequence *Sequence) BaseCheck() error {
	for i, j := 0, len(Sequence.Seq); i < j; i++ {
		switch base := unicode.ToUpper(rune(Sequence.Seq[i])); base {
		case 'A':
			Sequence.Seq[i] = byte(base)
		case 'C':
			Sequence.Seq[i] = byte(base)
		case 'T':
			Sequence.Seq[i] = byte(base)
		case 'G':
			Sequence.Seq[i] = byte(base)
		case 'N':
			Sequence.Seq[i] = byte(base)
		default:
			Sequence.Seq[i] = byte('N')
		}
	}
	return nil
}

// ReverseComplement is a method to reverse complement a sequence held by a FASTQread
func (FASTQread *FASTQread) ReverseComplement() {
	for i, j := 0, len(FASTQread.Seq); i < j; i++ {
		FASTQread.Seq[i] = complementBases[FASTQread.Seq[i]]
	}
	for i, j := 0, len(FASTQread.Seq)-1; i <= j; i, j = i+1, j-1 {
		FASTQread.Seq[i], FASTQread.Seq[j] = FASTQread.Seq[j], FASTQread.Seq[i]
		FASTQread.Qual[i], FASTQread.Qual[j] = FASTQread.Qual[j], FASTQread.Qual[i]
	}
	if FASTQread.RC == true {
		FASTQread.RC = false
	} else {
		FASTQread.RC = true
	}
}

// QualityTrim is a method to quality trim the sequence held by a FASTQread
/* the algorithm is based on bwa/cutadapt read quality trim functions:
-1. for each index position, subtract qual cutoff from the quality score
-2. sum these values across the read and trim at the index where the sum in minimal
-3. return the high-quality region
*/
func (FASTQread *FASTQread) QualityTrim(minQual int) {
	start, qualSum, qualMax := 0, 0, 0
	end := len(FASTQread.Qual)
	for i, qual := range FASTQread.Qual {
		qualSum += minQual - (int(qual) - FASTQ_ENCODING)
		if qualSum < 0 {
			break
		}
		if qualSum > qualMax {
			qualMax = qualSum
			start = i + 1
		}
	}
	qualSum, qualMax = 0, 0
	for i, j := 0, len(FASTQread.Qual)-1; j >= i; j-- {
		qualSum += minQual - (int(FASTQread.Qual[j]) - FASTQ_ENCODING)
		if qualSum < 0 {
			break
		}
		if qualSum > qualMax {
			qualMax = qualSum
			end = j
		}
	}
	if start >= end {
		start, end = 0, 0
	}
	FASTQread.Seq = FASTQread.Seq[start:end]
	FASTQread.Qual = FASTQread.Qual[start:end]
}
