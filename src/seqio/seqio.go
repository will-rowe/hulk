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

var (
	lengthErr = errors.New("sequence and quality score lines are unequal lengths in fastq file")
	idErr     = errors.New("read ID in fastq file does not begin with @")
)

// sequence base type
type sequence struct {
	id  []byte
	seq []byte
}

// FastqRead
type FastqRead struct {
	sequence
	misc   []byte
	qual   []byte
	length int
}

// QualTrim is a method to quality trim the sequence held in a FastqRead
// the algorithm is based on bwa/cutadapt read quality trim functions: -1. for each index position, subtract qual cutoff from the quality score -2. sum these values across the read and trim at the index where the sum in minimal -3. return the high-quality region
func (FastqRead *FastqRead) QualTrim(minQual int) {
	start, qualSum, qualMax := 0, 0, 0
	end := len(FastqRead.qual)
	for i, qual := range FastqRead.qual {
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
	for i, j := 0, len(FastqRead.qual)-1; j >= i; j-- {
		qualSum += minQual - (int(FastqRead.qual[j]) - ENCODING)
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
	FastqRead.seq = FastqRead.seq[start:end]
	FastqRead.qual = FastqRead.qual[start:end]
}

// Seq method returns the sequence held by the FastqRead
func (FastqRead *FastqRead) Seq() []byte {
	return FastqRead.sequence.seq
}

// Length method returns the length of the sequence held by the FastqRead
func (FastqRead *FastqRead) Length() int {
	return FastqRead.length
}

// Shred method splits the read sequence into chunks given a chunk size fraction
func (FastqRead *FastqRead) Shred(n float64) <-chan []byte {
	// calculate chunk size
	chunkSize := int(n * float64(FastqRead.length))
	// create the channel and chunk the sequence
	sendChan := make(chan []byte)
	go func() {
		defer close(sendChan)
		i, j := 0, chunkSize
		for {
			if j > FastqRead.length {
				break
			}
			sendChan <- FastqRead.sequence.seq[i:j]
			i += chunkSize
			j += chunkSize
		}
	}()
	return sendChan
}

// NewFastqRead is the FastqRead constructor
func NewFastqRead(l1 []byte, l2 []byte, l3 []byte, l4 []byte) (*FastqRead, error) {
	// check that it looks like a fastq read TODO: need more fastq checks
	if len(l2) != len(l4) {
		return &FastqRead{}, lengthErr
	}
	if l1[0] != 64 {
		return &FastqRead{}, idErr
	}
	read := &FastqRead{
		sequence: sequence{
			id:  l1,
			seq: l2,
		},
		misc:   l3,
		qual:   l4,
		length: len(l2),
	}
	return read, nil
}
