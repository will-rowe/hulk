// Package minimizer takes a sequence and finds the minimizers (for w consecutive k-mers)
// NOTE: currently it uses mapset to store the minimizers - this is unordered, which is fine for HULK but the minimizer sketch isn't that useful for anything else yet
package minimizer

import (
	"fmt"

	mapset "github.com/deckarep/golang-set"
	"github.com/will-rowe/hulk/src/queue"
)

// seq_nt4_table is used to convert "ACGTN" to 01234 - from minimap2
var seq_nt4_table = [256]uint8{
	0, 1, 2, 3, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 0, 4, 1, 4, 4, 4, 2, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 0, 4, 1, 4, 4, 4, 2, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
}

// hash64 - from minimap2
func hash64(key, mask uint64) uint64 {
	key = (^key + (key << 21)) & mask // key = (key << 21) - key - 1;
	key = key ^ key>>24
	key = ((key + (key << 3)) + (key << 8)) & mask // key * 265
	key = key ^ key>>14
	key = ((key + (key << 2)) + (key << 4)) & mask // key * 21
	key = key ^ key>>28
	key = (key + (key << 31)) & mask
	return key
}

// minimizerSketch
type minimizerSketch struct {
	k      uint // k-mer size
	w      uint // number of consecutive k-mers to minimize
	seq    []byte
	seqLen uint
	sketch mapset.Set // NOTE: this implementation of set is unordered
}

// GetMinimizers returns the sketch minimizers via a channel
func (minimizerSketch *minimizerSketch) GetMinimizers() <-chan interface{} {
	return minimizerSketch.sketch.Iter()
}

// NewMinimizerSketch is the constructor for a minimizerSketch
func NewMinimizerSketch(k, w uint, seq []byte) (*minimizerSketch, error) {

	// check the parameters
	if w < 0 || w > 256 {
		return nil, fmt.Errorf("w must be: 0 < w < 257")
	}
	if k < 0 || k > 31 {
		return nil, fmt.Errorf("k size must be: 0 < k < 32")
	}

	// check length of sequence
	len := uint(len(seq))
	if len < 1 {
		return nil, fmt.Errorf("sequence length must be > 0")
	}
	if len < (w + k - 1) {
		return nil, fmt.Errorf("sequence length must be >= w + k - 1")
	}

	// create the sketcher
	sketcher := &minimizerSketch{
		k:      k,
		w:      w,
		seq:    seq,
		seqLen: len,
		sketch: mapset.NewThreadUnsafeSet(),
	}

	// find the minimizers
	err := sketcher.findMinimizers()

	// return the minimizer sketch and any error
	return sketcher, err

}

// findMinimizers is the method to process a sequence and collect the minimzers
func (minimizerSketch *minimizerSketch) findMinimizers() error {

	// get a holder ready for the k-mer pair
	kmers := [2]uint64{0, 0}
	kmerSpan := uint(0)

	// bitmask is used to update the previous k-mer with the next base
	bitmask := (uint64(1) << uint64(2*minimizerSketch.k)) - uint64(1)
	bitshift := uint64(2 * (minimizerSketch.k - 1))

	q := queue.NewQueue()

	// start processing the sequence
	for i := uint(0); i < minimizerSketch.seqLen; i++ {

		// windowIndex helps keeps track of how many consecutive k-mers have been processed
		windowIndex := i - minimizerSketch.w + 1

		// get the nucleotide and convert to uint8
		c := seq_nt4_table[minimizerSketch.seq[i]]

		// if the nucleotide == N
		if c > 3 {

			// TODO: handle these, by skiping the base and starting w again?

		}

		// TODO: could do some homopolymer handling here (ala minimap2)

		// get the span of the k-mer we are about to collect
		if (windowIndex + 1) < minimizerSketch.k {
			kmerSpan = windowIndex + 1
		} else {
			kmerSpan = minimizerSketch.k
		}

		// get the forward k-mer
		kmers[0] = (kmers[0]<<2 | uint64(c)) & bitmask

		// get the reverse k-mer
		kmers[1] = (kmers[1] >> 2) | (uint64(3)^uint64(c))<<bitshift

		// don't try for minimizers until a full k-mer has been collected
		if i < minimizerSketch.k-1 {
			continue
		}

		// skip symmetric k-mers as we don't know the the strand
		if kmers[0] == kmers[1] {
			continue
		}

		// set the canonical k-mer
		var strand uint = 0
		if kmers[0] > kmers[1] {
			strand = 1
		}

		// hash the canonical k-mer
		currentKmer := queue.Pair{
			Value: hash64(kmers[strand], bitmask)<<8 | uint64(kmerSpan),
			ID:    i, // we are using the ID field of the pair for the k-mer location
		}

		// if there are already minimizers in the q, refresh the q
		if !q.IsEmpty() {

			// if minimizers are in the q from the previous window, remove them
			for {
				if q.IsEmpty() || (q.Front().ID > (i - minimizerSketch.w)) {
					break
				}
				q.PopFront()
			}

			// hashed k-mers less than equal to the currentKmer are not required, so remove them from the back of the q
			for {
				if q.IsEmpty() || (q.Back().Value < currentKmer.Value) {
					break
				}
				q.PopBack()
			}

		}

		// push the currentKmer and its position to the back of the q
		q.PushBack(currentKmer)

		// once the w consecutive k-mers have been processed, start adding minimizers from the q to the minimizer sketch
		if windowIndex >= 0 {

			// if the minimizer sketch is empty, add the minimizer from the start of the q to the sketch
			if minimizerSketch.sketch.Cardinality() == 0 {
				minimizerSketch.sketch.Add(q.Front().Value)
				continue
			}

			// if the sketch does not already have the minimizer at the start of the q, add it
			if minimizerSketch.sketch.Contains(q.Front().Value) {
				continue
			}
			minimizerSketch.sketch.Add(q.Front().Value)
		}

	} // end of sequence

	return nil
}
