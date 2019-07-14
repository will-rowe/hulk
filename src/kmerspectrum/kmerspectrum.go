// Package kmerspectrum uses a Go implementation of the jump consistent hash algorithm to bin hashed k-mers into k-mer spectrum bins (ala histogram)
package kmerspectrum

import (
	"fmt"

	"github.com/dgryski/go-jump"
	"github.com/will-rowe/hulk/src/bitvector"
	"github.com/will-rowe/hulk/src/helpers"
)

// MIN_USED_BINS is a threshold to change the behaviour of histosketching
// if the number of used bins in the k-mer spectrum < MIN_USED_BINS, then heuristics are used to add bins to a histosketch
const MIN_USED_BINS float64 = 0.01

// kmerSpectrumBin is a helper struct used to send data from the k-mer spectrum
type Bin struct {
	BinID     int32
	Frequency float64
}

// kmerSpectrum is the datastructure for a k-mer spectrum
type KmerSpectrum struct {
	numBins int32               // int32 used as this is the max allowed by jump hash implementation
	bins    []float64           // each index position in the slice corresponds to a bin
	bv      bitvector.BitVector // keeps track of which bins have been incremented
}

// NewKmerSpectrum is the constructor function
func NewKmerSpectrum(numBins int32) (*KmerSpectrum, error) {

	// check number of bins
	if numBins < int32(0) {
		return nil, fmt.Errorf("negative value used for number of k-mer spectrum bins: %d", numBins)
	}

	// zero the k-mer spectrum before use
	ks := &KmerSpectrum{
		numBins: numBins,
		bins:    make([]float64, numBins),
		bv:      bitvector.NewBitVector(int(numBins)),
	}
	ks.Wipe()
	return ks, nil
}

// Size is a method to return the number of bins in the k-mer spectrum
func (KmerSpectrum *KmerSpectrum) Size() uint {
	return uint(KmerSpectrum.numBins)
}

// Cardinality is a method to calculate the number of currently used bins in the k-mer spectrum
func (KmerSpectrum *KmerSpectrum) Cardinality() int {
	return KmerSpectrum.bv.PopCount()
}

// Wipe is a method to clear all the bins in a k-mer spectrum
func (KmerSpectrum *KmerSpectrum) Wipe() {
	for i := int32(0); i < KmerSpectrum.numBins; i++ {
		KmerSpectrum.bins[i] = 0
	}
	KmerSpectrum.bv.WipeOut()
	return
}

// AddHash is a method to add a hashed k-mer to the spectrum
func (KmerSpectrum *KmerSpectrum) AddHash(kmer uint64) error {

	// get the bin for this k-mer
	bin := jump.Hash(kmer, int(KmerSpectrum.numBins))

	// record the bin id
	if err := KmerSpectrum.bv.Add(int(bin)); err != nil {
		return err
	}

	// increment the bin
	KmerSpectrum.bins[bin]++

	return nil
}

// Dump is a method that returns each counter value in the CMS (returned via channel)
func (KmerSpectrum *KmerSpectrum) Dump() (<-chan *Bin, error) {

	// check how many bins in the k-mer spectrum need adding
	usedBins := KmerSpectrum.Cardinality()
	propUsed := float64(usedBins) / float64(KmerSpectrum.numBins)
	if usedBins == 0 {
		return nil, fmt.Errorf("k-mer spectrum is empty")
	}

	// TODO: if the proportion of used bins is below a threshold, use the bit vector to quicky determine which bins are used
	if propUsed < MIN_USED_BINS {
		return nil, fmt.Errorf("not used yet")
	}

	// make the dumper channel
	dumper := make(chan *Bin)
	go func() {

		// iterate over the bins in the k-mer spectrum, sending any used bin to the dumper
		for i := int32(0); i < KmerSpectrum.numBins; i++ {
			if KmerSpectrum.bins[i] != 0.0 {
				dumper <- &Bin{i, KmerSpectrum.bins[i]}
			}
		}

		close(dumper)
	}()
	return dumper, nil
}

// Print is a method to print the k-mer spectrum bin values as a comma separated list
func (KmerSpectrum *KmerSpectrum) Print() string {
	return helpers.FloatSlice2string(KmerSpectrum.bins, ",")
}
