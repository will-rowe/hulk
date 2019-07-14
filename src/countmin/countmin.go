// Package countmin is an implementation of the CountMin Sketch (https://sites.google.com/site/countminsketch/cm-latin.pdf?attredirects=0)
package countmin

import (
	"math"

	jump "github.com/dgryski/go-jump"
)

// EPSILON is a default value for epsilon (error rate)
const EPSILON float64 = 0.001

// DELTA is a default value for delta (confidence)
const DELTA float64 = 0.99

// CountMinSketch is the CountMin sketch data structure
type CountMinSketch struct {
	epsilon      float64     // relative-accuracy factor
	delta        float64     // relative-accuracy probability
	sketch       [][]float64 // this q in the paper, which is a matrix of d x g
	depth        uint32      // this is d in the paper, which is the number of hash tables in the sketch
	width        uint32      // this is g in the paper, which is the number of counters per hash table
	applyScaling bool        // if true, uniform scaling will be applied to the counters using the decay weight
	decayWeight  float64     // the decay weight for scaling
}

// NewCountMinSketch is the constructor function. The CountMin Sketch relative accuracy is within a factor of epsilon with probability delta.
func NewCountMinSketch(epsilon, delta, decayRatio float64) *CountMinSketch {

	// calculate the sketch size
	g := uint32(math.Ceil(2 / epsilon))
	d := uint32(math.Ceil(math.Log(1-delta) / math.Log(0.5)))

	// initialise the sketch
	q := make([][]float64, d)
	for i := uint32(0); i < d; i++ {
		q[i] = make([]float64, g)
	}

	// create the data structure
	newSketch := &CountMinSketch{
		epsilon: epsilon,
		delta:   delta,
		sketch:  q,
		depth:   d,
		width:   g,
	}

	// set the decay weight
	if decayRatio > 0.0 && decayRatio < 1.0 {
		newSketch.decayWeight = float64(math.Exp(-decayRatio))
		newSketch.applyScaling = true
	} else {
		newSketch.applyScaling = false
	}
	return newSketch
}

// Dump is a method that returns each counter value in the sketch
func (CountMinSketch *CountMinSketch) Dump() <-chan float64 {
	dumper := make(chan float64)
	go func() {
		for d := uint32(0); d < CountMinSketch.depth; d++ {
			for g := uint32(0); g < CountMinSketch.width; g++ {
				dumper <- CountMinSketch.sketch[d][g]
			}
		}
		close(dumper)
	}()
	return dumper
}

// Wipe is a method to clear all the counters in the sketch
func (CountMinSketch *CountMinSketch) Wipe() {
	q := make([][]float64, CountMinSketch.depth)
	for i := uint32(0); i < CountMinSketch.depth; i++ {
		q[i] = make([]float64, CountMinSketch.width)
	}
	CountMinSketch.sketch = q
}

// GetDepth is a method to return the number of hash tables (d) in the sketch
func (CountMinSketch *CountMinSketch) GetDepth() uint32 {
	return CountMinSketch.depth
}

// GetWidth is a method to return the number of counters (g) used in each table of the sketch
func (CountMinSketch *CountMinSketch) GetWidth() uint32 {
	return CountMinSketch.width
}

// GetDecayWeight is a method to return the decay weight set in the data structure
func (CountMinSketch *CountMinSketch) GetDecayWeight() float64 {
	return CountMinSketch.decayWeight
}

// GetEstimate is a method to get the estimated frequency of a query element
func (CountMinSketch *CountMinSketch) GetEstimate(element uint64) float64 {
	return CountMinSketch.traverse(element, 0.0)
}

// Add is a method to add an element to the sketch
func (CountMinSketch *CountMinSketch) Add(element uint64, increment float64) float64 {

	// determine if uniform scaling needs to be applied to the sketch counters
	if CountMinSketch.applyScaling == true {
		CountMinSketch.scale()
	}
	return CountMinSketch.traverse(element, increment)
}

// traverse is an unexported method to identify a counter for a query element in each table of the count-min sketch
func (CountMinSketch *CountMinSketch) traverse(element uint64, increment float64) float64 {

	// max the minimum to begin with, to ensure the element is placed somewhere
	currentMinimum := math.MaxFloat64

	// find the counter for the element in each table of the sketch
	for d := uint32(0); d < CountMinSketch.depth; d++ {

		// hash for element
		hash := element + (uint64(d) * element)

		// use consistent jump hash to get counter position in this table
		g := jump.Hash(hash, int(CountMinSketch.width))

		// increment the counter if requested
		if increment != 0.0 {
			CountMinSketch.sketch[d][g] += increment
		}

		// evaluate if the current counter is the minimum
		if CountMinSketch.sketch[d][g] < currentMinimum {
			currentMinimum = CountMinSketch.sketch[d][g]
		}
	}
	return currentMinimum
}

// scale is an unexported method to uniformly scale each counter in the sketch
func (CountMinSketch *CountMinSketch) scale() {
	for d := uint32(0); d < CountMinSketch.depth; d++ {
		for g := uint32(0); g < CountMinSketch.width; g++ {
			CountMinSketch.sketch[d][g] = CountMinSketch.sketch[d][g] * CountMinSketch.decayWeight
		}
	}
}
