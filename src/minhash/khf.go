package minhash

import (
	"fmt"
	"math"

	"github.com/will-rowe/hulk/src/helpers"
)

// KHFsketch is the K-Hash Functions MinHash sketch of a set
type KHFsketch struct {
	algo       string
	KmerSize   uint     `json:"ksize"`
	Md5sum     string   `json:"md5sum"`
	Sketch     []uint64 `json:"mins"`
	SketchSize uint     `json:"num"`
}

// NewKHFsketch is the constructor for a KHFsketch
func NewKHFsketch(k, s uint) *KHFsketch {
	// init the sketch with maximum values
	sketch := make([]uint64, s)
	for i := range sketch {
		sketch[i] = math.MaxUint64
	}
	return &KHFsketch{
		algo:       "khf",
		KmerSize:   k,
		SketchSize: s,
		Sketch:     sketch,
	}
}

// AddHash is a method to evaluate a hash value and add any minimums to the sketch
func (KHFsketch *KHFsketch) AddHash(hv uint64) {
	// for each sketch slot, derive a new hash value
	for i := uint(0); i < KHFsketch.SketchSize; i++ {
		val := hv + (uint64(i) * hv)
		// evaluate and add to the current sketch slot if it is a minimum
		if val < KHFsketch.Sketch[i] {
			KHFsketch.Sketch[i] = val
		}
	}
	return
}

// Merge is a method to combine two MinHash objects
// TODO: this should check for consistency between MinHash objects
func (KHFsketch *KHFsketch) Merge(KHFsketch2 *KHFsketch) {
	for i, minimum := range KHFsketch2.Sketch {
		if minimum < KHFsketch.Sketch[i] {
			KHFsketch.Sketch[i] = minimum
		}
	}
}

// GetSketch is a method to return the sketch held by a MinHash KHF sketch object
func (KHFsketch *KHFsketch) GetSketch() []uint64 {
	return KHFsketch.Sketch
}

// SetMD5 is a method to calculate and store the MD5 for the sketch
func (KHFsketch *KHFsketch) SetMD5() {
	KHFsketch.Md5sum = fmt.Sprintf("%x", helpers.MD5sum(KHFsketch.Sketch))
	return
}

// GetMD5 is a method to return the MD5 currently calculated for the sketch
func (KHFsketch *KHFsketch) GetMD5() string {
	return KHFsketch.Md5sum
}

// GetAlgo is a method to return the sketching algorithm used
func (KHFsketch *KHFsketch) GetAlgo() string {
	return KHFsketch.algo
}

// GetSimilarity is a function to estimate the Jaccard similarity between sketches
func (KHFsketch *KHFsketch) GetSimilarity(mh2 MinHash) (float64, error) {

	// check that MinHash flavours match
	if fmt.Sprintf("%T", KHFsketch) != fmt.Sprintf("%T", mh2) {
		return 0.0, fmt.Errorf("mismatched MinHash types: %T vs. %T", KHFsketch, mh2)
	}

	// calculate the jaccard index
	intersect := 0.0
	sketch1 := KHFsketch.GetSketch()
	sketch2 := mh2.GetSketch()
	sharedLength := len(sketch1)
	if sharedLength > len(sketch2) {
		sharedLength = len(sketch2)
	}
	for i := 0; i < sharedLength; i++ {
		if sketch1[i] == sketch2[i] {
			intersect++
		}
	}
	return (intersect / float64(sharedLength)), nil
}
