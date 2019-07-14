// Package bitvector contains types and methods for working with bit vectors
package bitvector

import (
	"fmt"
	"go/build"
)

// MAX_UINT
const MAX_UINT = ^uint(0)

// MAX_INT
const MAX_INT = int(MAX_UINT >> 1)

// MAX_SIZE is calculated at compile time
var MAX_SIZE = maxSize()

// PC64 is used to determine population count, e.g. PC64[i] is the population count of i
var PC64 = pc64()

// BitVector is the base type
type BitVector []int

// NewBitVector is the constructor
// Given the maximum value to hold in the bit vector, return an int array sized to do so
func NewBitVector(maxValue int) BitVector {
	sizeOfInt := 1
	if maxValue > MAX_SIZE {
		sizeOfInt = (maxValue / MAX_SIZE) + 1
	}
	return make([]int, sizeOfInt, sizeOfInt)
}

// Add is a method to add an integer to the bit vector
func (BitVector BitVector) Add(k int) error {
	if k > (len(BitVector) * MAX_SIZE) {
		return fmt.Errorf("%d is too large for the current bit vector capacity (%d)", k, (len(BitVector) * MAX_SIZE))
	}
	// check if it is already there before adding
	if ok := BitVector.Contains(k); !ok {
		BitVector[k/MAX_SIZE] |= (1 << uint(k%MAX_SIZE))
	}
	return nil
}

// Contains is a method to check a bit vector for a value
func (BitVector BitVector) Contains(k int) bool {
	return (BitVector[k/MAX_SIZE] & (1 << uint(k%MAX_SIZE))) != 0
}

// Delete is a method to delete a value from a bit vector
func (BitVector BitVector) Delete(k int) {
	BitVector[k/MAX_SIZE] &= ^(1 << uint(k%MAX_SIZE))
}

// MaxOut is a method to flip all the bits in the bit vector to 1
func (BitVector BitVector) MaxOut() {
	for i := 0; i < len(BitVector); i++ {
		BitVector[i] = MAX_INT
	}
}

// WipeOut is a method to flip all the bits in the bit vector to 0
func (BitVector BitVector) WipeOut() {
	for i := 0; i < len(BitVector); i++ {
		BitVector[i] = 0
	}
}

// BWAND is a method to return the bitwise AND of two BitVectors.
func (BitVector BitVector) BWAND(bv2 BitVector) (BitVector, error) {
	if len(BitVector) != len(bv2) {
		return nil, fmt.Errorf("Can't perform bitwise AND when BitVectors are different sizes")
	}
	result := NewBitVector(len(BitVector) * MAX_SIZE)
	for i := 0; i < len(BitVector); i++ {
		result[i] = BitVector[i] & bv2[i]
	}
	return result, nil
}

// PopCount is a method to report the number of flipped bits in the bit vector (using the population count algorithm)
func (BitVector BitVector) PopCount() int {
	count := 0
	if MAX_SIZE == 64 {
		for i := 0; i < len(BitVector); i++ {
			count += int(PC64[byte(BitVector[i]>>(0*8))] +
				PC64[byte(BitVector[i]>>(1*8))] +
				PC64[byte(BitVector[i]>>(2*8))] +
				PC64[byte(BitVector[i]>>(3*8))] +
				PC64[byte(BitVector[i]>>(4*8))] +
				PC64[byte(BitVector[i]>>(5*8))] +
				PC64[byte(BitVector[i]>>(6*8))] +
				PC64[byte(BitVector[i]>>(7*8))])
		}
	} else if MAX_SIZE == 32 {
		for i := 0; i < len(BitVector); i++ {
			count += int(PC64[byte(BitVector[i]>>(0*8))] +
				PC64[byte(BitVector[i]>>(1*8))] +
				PC64[byte(BitVector[i]>>(2*8))] +
				PC64[byte(BitVector[i]>>(3*8))])
		}
	} else {
		panic("I need to extend popcount to work beyond int64")
	}
	return count
}

// maxSize returns the system architecture in bits
func maxSize() int {
	var maxSize int
	switch arch := build.Default.GOARCH; arch {
	case "amd64":
		maxSize = 64
	case "arm64":
		maxSize = 64
	case "js":
		maxSize = 64
	default:
		maxSize = 32 // TODO: this should be an error
	}
	return maxSize
}

// populates the reference bit vector
func pc64() [256]byte {
	var pc64 [256]byte
	for i := range pc64 {
		pc64[i] = pc64[i/2] + byte(i&1)
	}
	return pc64
}
