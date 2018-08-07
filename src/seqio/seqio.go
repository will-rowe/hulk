/*
	the seqio package contains custom types and methods for holding and processing sequence data
*/
package seqio

import (
	"errors"
)

const (
	ENCODING = 33 // fastq encoding used
)

/*
  the base type
*/
type Sequence struct {
	ID  []byte
	Seq []byte
}

/*
  struct to hold FASTQ data and seed locations for a single read
*/
type FASTQread struct {
	Sequence
	Misc []byte
	Qual []byte
}

// method to quality trim the sequence held in a FASTQread
// the algorithm is based on bwa/cutadapt read quality trim functions: -1. for each index position, subtract qual cutoff from the quality score -2. sum these values across the read and trim at the index where the sum in minimal -3. return the high-quality region
func (self *FASTQread) QualTrim(minQual int) {
	start, qualSum, qualMax := 0, 0, 0
	end := len(self.Qual)
	for i, qual := range self.Qual {
		qualSum += minQual - (int(qual) - ENCODING)
		if qualSum < 0 {
			break
		}
		if qualSum > qualMax {
			qualMax = qualSum
			start = i + 1
		}
	}
	qualSum, qualMax = 0, 0
	for i, j := 0, len(self.Qual)-1; j >= i; j-- {
		qualSum += minQual - (int(self.Qual[j]) - ENCODING)
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
	self.Seq = self.Seq[start:end]
	self.Qual = self.Qual[start:end]
}

/*
  function to generate new fastq read from 4 lines of a fastq
*/
func NewFASTQread(l1 []byte, l2 []byte, l3 []byte, l4 []byte) (FASTQread, error) {
	// check that it looks like a fastq read TODO: need more fastq checks
	if len(l2) != len(l4) {
		return FASTQread{}, errors.New("sequence and quality score lines are unequal lengths in fastq file")
	}
	if l1[0] != 64 {
		return FASTQread{}, errors.New("read ID in fastq file does not begin with @")
	}
	// create a FASTQread struct
	read := new(FASTQread)
	read.ID = l1
	read.Seq = l2
	read.Misc = l3
	read.Qual = l4
	return *read, nil
}
