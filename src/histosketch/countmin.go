// countmin sketch for uint64 encoded kmers implementation in Go
package histosketch

import (
	"math"

	jump "github.com/dgryski/go-jump"
)

// CountMinSketch includes a decay weight for uniform scaling of counts
type CountMinSketch struct {
	q           [][]float64 // matrix of d x g
	d           uint32      // matrix depth (number of hash tables)
	g           uint32      // matrix width (number of counters per table)
	scaling     bool        // if true, uniform scaling will be applied to the counters using the decay weight
	weightDecay float64		// the decay weight for scaling
	epsilon	float64	// relative-accuracy factor
	delta	float64	// relative-accuracy probability
}

// NewCountMinSketch creates a new Count-Min Sketch whose relative accuracy is
// within a factor of epsilon with probability delta. Both of these parameters
// affect the space and time complexity.
func NewCountMinSketch(epsilon, delta, decayRatio float64) *CountMinSketch {
	// calculate the countminsketch dimensions
	g := uint32(math.Ceil(2 / epsilon))
	d := uint32(math.Ceil(math.Log(1-delta) / math.Log(0.5)))
	// make the matrix
	q := make([][]float64, d)
	for i := uint32(0); i < d; i++ {
		q[i] = make([]float64, g)
	}
	// create the CMS
	s := &CountMinSketch{
		q:      q,
		d:      d,
		g:      g,
		epsilon: epsilon,
		delta:	delta,
	}
	// set the decay weight
	if decayRatio != 1 {
		s.weightDecay = float64(math.Exp(-decayRatio))
		s.scaling = true
	}
	return s
}

// Tables is a method to return the number of hash tables (d) used in the CMS
func (CountMinSketch *CountMinSketch) Tables() uint32 {
	return CountMinSketch.d
}

// Counters is a method to return the number of counters (g) used in each table of the CMS
func (CountMinSketch *CountMinSketch) Counters() uint32 {
	return CountMinSketch.g
}

// Epsilon is a method to return the epsilon value of the CMS
func (CountMinSketch *CountMinSketch) Epsilon() float64 {
	return CountMinSketch.epsilon
}

// Delta is a method to return the delta value of the CMS
func (CountMinSketch *CountMinSketch) Delta() float64 {
	return CountMinSketch.delta
}

// Wipe is a method to clear the kmer from a CountMinSketch
func (CountMinSketch *CountMinSketch) Wipe() {
	q := make([][]float64, CountMinSketch.d)
	for i := uint32(0); i < CountMinSketch.d; i++ {
		q[i] = make([]float64, CountMinSketch.g)
	}
	CountMinSketch.q = q
}

// Copy is a method to return an empty copy of a CountMinSketch
func (cms *CountMinSketch) Copy() *CountMinSketch {
	q := make([][]float64, cms.d)
	for i := uint32(0); i < cms.d; i++ {
		q[i] = make([]float64, cms.g)
	}
	return &CountMinSketch{
		q:      q,
		d:      cms.d,
		g:      cms.g,
		epsilon:	cms.epsilon,
		delta:	cms.delta,
	}
}

// Get is a method to get the minimum from a count-min sketch for a given query kmer
func (CountMinSketch *CountMinSketch) Get(kmer uint64) float64 {
	return CountMinSketch.traverse(kmer, 0.0)
}

// Add is a method to add kmer to the count-min sketch
func (CountMinSketch *CountMinSketch) Add(kmer uint64, increment float64) float64 {
	if CountMinSketch.scaling == true {
		// uniform scaling of all sketch counters
		CountMinSketch.scale()
	}
	return CountMinSketch.traverse(kmer, increment)
}

// Merge is a method that merges the current sketch with another TODO: adds some checks and error return
func (CountMinSketch *CountMinSketch) Merge(sketch2 *CountMinSketch) error {
	for d := uint32(0); d < CountMinSketch.d; d++ {
		for g := uint32(0); g < CountMinSketch.g; g++ {
			CountMinSketch.q[d][g] += sketch2.q[d][g]
		}
	}
	return nil
}

// Dump is a method that returns each counter value in the CMS (returned via channel)
func (CountMinSketch *CountMinSketch) Dump() <-chan float64 {
	dumper := make(chan float64)
	go func() {
		for d := uint32(0); d < CountMinSketch.d; d++ {
			for g := uint32(0); g < CountMinSketch.g; g++ {
				dumper <- CountMinSketch.q[d][g]
			}
		}
		close(dumper)
	}()
	return dumper
}

// traverse is a method to traverse the count-min matrix for a given query
func (CountMinSketch *CountMinSketch) traverse(kmer uint64, increment float64) float64 {
	// set the counter minimum to a max value
	minimum := math.MaxFloat64
	// use the hashed kmer to look up the counter for this kmer in each row (d)
	for d := uint32(0); d < CountMinSketch.d; d++ {
		// get the hash permutation for this kmer and this table
		hash := kmer + (uint64(d) * kmer)
		// use consistent jump hash to get counter position
		pos := jump.Hash(hash, int(CountMinSketch.g))
		// increment the counter count if we are adding an element
		if increment != 0.0 {
			CountMinSketch.q[d][pos] += increment
		}
		// evaluate if the current counter is the minimum
		if CountMinSketch.q[d][pos] < minimum {
			minimum = CountMinSketch.q[d][pos]
		}
	}
	return minimum
}

// scale is a method that adjusts each counter in q using a decay weight
func (CountMinSketch *CountMinSketch) scale() {
	for d := uint32(0); d < CountMinSketch.d; d++ {
		for g := uint32(0); g < CountMinSketch.g; g++ {
			CountMinSketch.q[d][g] = CountMinSketch.q[d][g] * CountMinSketch.weightDecay
		}
	}
}
