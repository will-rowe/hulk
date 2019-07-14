package minhash

import (
	"container/heap"
	"fmt"
	"sort"

	"github.com/will-rowe/hulk/src/helpers"
)

// KMVsketch is the bottom-k MinHash sketch of a set
type KMVsketch struct {
	algo            string
	KmerSize        uint     `json:"ksize"`
	Md5sum          string   `json:"md5sum"`
	Sketch          []uint64 `json:"mins"`
	SketchSize      uint     `json:"num"`
	heap            *IntHeap
	maxCardinalty   int // not yet used
	multiplicitySum int // not yet used
}

// NewKMVsketch is the constructor for a KMVsketch
func NewKMVsketch(k, s uint) *KMVsketch {
	newSketch := &KMVsketch{
		algo:            "kmv",
		KmerSize:        k,
		SketchSize:      s,
		heap:            &IntHeap{},
		maxCardinalty:   0,
		multiplicitySum: 0,
	}

	// init the heap
	heap.Init(newSketch.heap)
	return newSketch
}

// AddHash is a method to evaluate a hash value and add any minimums to the sketch
func (KMVsketch *KMVsketch) AddHash(hv uint64) {
	/*
		TODO: if hash tracking:

		evaluate if the current hash tracker is currently less than KMVsketch.maxCardinalty
			- if it is, this needs to signal to the sketcher

		check if the incoming hash has been seen before
			- could use a Bloom Filter for this
			- if seen, then increment the tracker and try adding the hash to the heap
			- if Bloom says no, then just add it to the Bloom, increment the multiplicitySum and exit (stopping a possibly unique hash being added to the sketch)
			- once finished, the tracker will be off by 1 for each hash
	*/

	// increment the multiplicity
	KMVsketch.multiplicitySum++

	// if the heap isn't full yet, go ahead and add the hash
	if len(*KMVsketch.heap) < int(KMVsketch.SketchSize) {
		heap.Push(KMVsketch.heap, hv)

		// or if the incoming hash is smaller than the hash at the top of the heap, add the hash and remove the larger one from the heap
	} else if hv < (*KMVsketch.heap)[0] {

		// replace the largest value currently in the sketch with the new hash
		(*KMVsketch.heap)[0] = hv

		// re-establish the heap ordering after adding the new hash
		heap.Fix(KMVsketch.heap, 0)
	}
	return
}

/*
// Merge is a method to combine two bottomK MinHash objects
func (KMVsketch *KMVsketch) Merge(querySketch *KMVsketch) {

	// TODO: add sketch compatibility check

	// make sure the incoming sketch is set and sorted
	querySketch.SetSketch()

	// duplicateChecker stops the same hash being added to the sketch multiple times (we over-sketched the incoming sketch to allow for this)
	duplicateChecker := make(map[uint64]struct{})

	// set up the duplicateChecker with the existing sketch
	for _, hv := range *KMVsketch.heap {
		duplicateChecker[hv] = struct{}{}
	}

	// check each hash value in the incoming sketch
	for _, hv := range *querySketch.heap {

		// ignore hash values already seen
		if _, ok := duplicateChecker[hv]; ok {
			continue
		}
		duplicateChecker[hv] = struct{}{}

		// if the heap isn't full yet, go ahead and add in the hash from the incoming sketch
		if len(*KMVsketch.heap) < int(KMVsketch.SketchSize) {
			heap.Push(KMVsketch.heap, hv)
			continue
		}

		// if the heap is full but the new hash is smaller than the largest value in the heap, swap it in
		if hv < (*KMVsketch.heap)[0] {
			(*KMVsketch.heap)[0] = hv
			heap.Fix(KMVsketch.heap, 0)

			// as the incoming sketch is sorted low -> high, we can end the merge early if the heap is full and the current hash is too large
		} else {
			break
		}
	}
}
*/

// Similarity computes a similarity estimate for two KMV sketches
func (mh1 *KMVsketch) GetSimilarity(mh2 MinHash) (float64, error) {

	// check this is a pair of KMV
	if fmt.Sprintf("%T", mh1) != fmt.Sprintf("%T", mh2) {
		return 0.0, fmt.Errorf("mismatched MinHash types: %T vs. %T", mh1, mh2)
	}
	sketch2, ok := mh2.(*KMVsketch)
	if !ok {
		return 0.0, fmt.Errorf("could not assert sketch is a KMV")
	}

	// check the sketch lengths match
	//if mh1.SketchSize != sketch2.SketchSize {
	//	return 0.0, fmt.Errorf("KMV sketch lengths don't match")
	//}

	// assign longer and shorter heaps
	longer, shorter := &IntHeap{}, &IntHeap{}
	if len(*mh1.heap) > len(*sketch2.heap) {
		longer, shorter = mh1.heap, sketch2.heap
	} else {
		shorter, longer = mh1.heap, sketch2.heap
	}

	// make a map of one of the sketches, recording each unique minimum and its count
	minimums := make(map[uint64]int, len(*longer))
	for _, v := range *longer {
		minimums[v]++
	}

	// iterate over the other sketch and calculate the intersect
	intersect := 0
	for _, minimum := range *shorter {
		if count, ok := minimums[minimum]; ok && count > 0 {
			minimums[minimum] = count - 1
			intersect++
		}
	}

	return (float64(intersect) / float64(len(*longer))), nil
}

// SetSketch converts the current IntHeap into a []uint64 and sorts it low -> high
func (KMVsketch *KMVsketch) SetSketch() {
	KMVsketch.Sketch = make([]uint64, len(*KMVsketch.heap))
	for i, val := range *KMVsketch.heap {
		KMVsketch.Sketch[i] = val
	}
	KMVsketch.SketchSize = uint(len(KMVsketch.Sketch))
	sort.Slice(KMVsketch.Sketch, func(i, j int) bool { return KMVsketch.Sketch[i] < KMVsketch.Sketch[j] })
}

// GetSketch is a method to set and return the sketch held by a MinHash KMV sketch
func (KMVsketch *KMVsketch) GetSketch() []uint64 {
	if len(KMVsketch.Sketch) == 0 {
		KMVsketch.SetSketch()
	}
	return KMVsketch.Sketch
}

// SetMD5 is a method to calculate and store the MD5 for the sketch
func (KMVsketch *KMVsketch) SetMD5() {
	KMVsketch.Md5sum = fmt.Sprintf("%x", helpers.MD5sum(KMVsketch.GetSketch()))
	return
}

// GetMD5 is a method to return the MD5 currently calculated for the sketch
func (KMVsketch *KMVsketch) GetMD5() string {
	return KMVsketch.Md5sum
}

// GetAlgo is a method to return the sketching algorithm used
func (KMVsketch *KMVsketch) GetAlgo() string {
	return KMVsketch.algo
}
