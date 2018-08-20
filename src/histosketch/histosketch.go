/*
 histosketch is a Go implementation of HistoSketch: Fast Similarity-Preserving Sketching of Streaming Histograms with Concept Drift (https://exascale.info/assets/pdf/icdm2017_HistoSketch.pdf)

I've made some changes in my implementation compared to the paper:
  - Instead of providing the number of histogram bins (dimensions) and the number of countmin hash tables (d), I have decided to use epsilon and delta values to calculate CMS dimensions.
  - As I am using HistoSketch to sketch CMS counters, the dimensions of the histosketch are determined by the CMS dimensions
*/

package histosketch

import (
	"encoding/gob"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/JSchwehn/goDistances"
	"github.com/leesper/go_rng"
)

// CWS base type
type CWS struct {
	r [][]float64 // r in the paper
	c [][]float64 // c in the paper
	b [][]float64 // beta in the paper (adapted here to pre-calculate beta*r )
}

// HistoSketch base type
type HistoSketch struct {
	// structure
	length       uint            // number of minimums in the histosketch
	dimensions   uint            // number of histogram bins
	distSeed     int64           // seed used to generate the distributions
	samples      *CWS            // the consistent weighted samples
	sketch       []uint          // S in paper
	sketchHashes []float64       // A in paper
	cmSketch     *CountMinSketch // Q in the paper (d * g matrix, where g is sketch length)
	decayRatio   float64         // the decay ratio used for concept drift (1.00 = concept drift disabled)
}

// NewCWS generates the Consistent Weighted Samples for a histosketch
func (HistoSketch *HistoSketch) newCWS() {
	// create the holders
	r := make([][]float64, HistoSketch.length)
	c := make([][]float64, HistoSketch.length)
	b := make([][]float64, HistoSketch.length)
	// set up the consistent weighted sampling by drawing 3 sets of samples from a Gamma distribution, log Gamma distribution and a uniform distribution
	gammaGenerator := rng.NewGammaGenerator(HistoSketch.distSeed)     // a random number generator for gamma distribution
	uniformGenerator := rng.NewUniformGenerator(HistoSketch.distSeed) // a random number generator for a uniform distribution
	// create the samples
	for i := uint(0); i < HistoSketch.length; i++ {
		r[i] = make([]float64, HistoSketch.dimensions)
		c[i] = make([]float64, HistoSketch.dimensions)
		b[i] = make([]float64, HistoSketch.dimensions)
		for j := uint(0); j < HistoSketch.dimensions; j++ {
			r[i][j] = gammaGenerator.Gamma(2, 1)
			c[i][j] = math.Log(gammaGenerator.Gamma(2, 1))
			//b[i][j] = uniformGenerator.Float64Range(0, 1) // as in paper
			// I've multiplied beta by r and stored this instead of just beta
			b[i][j] = uniformGenerator.Float64Range(0, 1) * r[i][j]
		}
	}
	// set the samples
	HistoSketch.samples = &CWS{
		r: r,
		c: c,
		b: b,
	}
}

// NewHistoSketch is a HistoSketch constructor. It creates a sketch of length (l) from a histogram
func NewHistoSketch(l uint, h *histogram, epsilon, delta, dr float64) *HistoSketch {
	// create a new base histosketch
	hs := HistoSketch{
		length:     l,
		dimensions: uint(len(h.bins)),
		distSeed:   1, // TODO: should this be customisable?
		decayRatio: dr,
	}
	// create the empty sketches
	hs.createSketches(epsilon, delta)
	// create the CWS samples
	hs.newCWS()
	// zero the HistoSketch
	for i := range hs.sketch {
		hs.sketch[i] = 0
		hs.sketchHashes[i] = math.MaxFloat64
	}
	// return a pointer to the histosketch
	return &hs
}

// getSample method returns A_ka for a given element (i = bin, j = sketch position)
func (CWS *CWS) getSample(i uint, j int, freq float64) float64 {
	Y_ka := math.Exp(math.Log(freq) - CWS.b[j][i])
	return CWS.c[j][i] / (Y_ka * math.Exp(CWS.r[j][i]))
}

// createSketches creates the histosketch, as well the countmin sketch
func (HistoSketch *HistoSketch) createSketches(epsilon, delta float64) {
	HistoSketch.sketch = make([]uint, HistoSketch.length)
	HistoSketch.sketchHashes = make([]float64, HistoSketch.length)
	// create the empty count min sketch
	HistoSketch.cmSketch = NewCountMinSketch(epsilon, delta, HistoSketch.decayRatio)
}

// Update method is used to update the HistoSketch with a new element
func (HistoSketch *HistoSketch) Update(bin uint64, value float64) error {
	// the countMinSketch Add() method first updates the cmSketch (applies uniform scaling), adds the bin and returns the estimated frequency
	estiFreq := HistoSketch.cmSketch.Add(bin, value)
	// consistent weighted sampling for the incoming element
	for j := range HistoSketch.sketch {
		// get A_ka for element at this sketch position
		A_ka := HistoSketch.samples.getSample(uint(bin), j, estiFreq)
		// evaluate A_ka against the existing minimum (using decay weight adjustment if requested)
		var curMin float64
		if HistoSketch.decayRatio != 1.0 {
			curMin = HistoSketch.sketchHashes[j] / HistoSketch.cmSketch.weightDecay
		} else {
			curMin = HistoSketch.sketchHashes[j]
		}
		// apply decay weight to old sketch element and see if the A_ka is a new minimum
		if A_ka < curMin {
			// replace minimum bin index and hash
			HistoSketch.sketch[j] = uint(bin)
			HistoSketch.sketchHashes[j] = A_ka
		}
	}
	return nil
}

// GetSketch method will return the HistoSketch as a comma-separated string
func (HistoSketch *HistoSketch) GetSketch() string {
	var sketch string
	for i := uint(0); i < HistoSketch.length; i++ {
		if i == (HistoSketch.length - 1) {
			sketch = fmt.Sprintf("%v%d", sketch, HistoSketch.sketch[i])
		} else {
			sketch = fmt.Sprintf("%v%d,", sketch, HistoSketch.sketch[i])
		}
	}
	return sketch
}

// the SketchStore type is used to retain the minimum info needed about a sketch and to save it to disk. All fields are exported so that gob will encode (is there a work around?)
type SketchStore struct {
	File         string    // the file that the sketch was loaded from
	Length       uint      // number of minimums in the histosketch
	Dimensions   uint      // number of histogram bins
	DistSeed     int64     // seed used to generate the distributions
	Sketch       []uint    // S in paper
	SketchHashes []float64 // A in paper
}

// SaveSketch method will encode the sketch (and minimum required info) and then write them to disk
func (HistoSketch *HistoSketch) SaveSketch(outfile string) error {
	// create the SketchStore
	store := &SketchStore{
		File:       outfile,
		Length:     HistoSketch.length,
		Dimensions: HistoSketch.dimensions,
		DistSeed:   HistoSketch.distSeed,
	}
	store.Sketch = HistoSketch.sketch
	store.SketchHashes = HistoSketch.sketchHashes
	// encode and write it to disk
	fh, err := os.Create(outfile)
	defer fh.Close()
	if err != nil {
		return err
	}
	encoder := gob.NewEncoder(fh)
	err = encoder.Encode(store)
	return err
}

// LoadSketch function will open a saved sketch and return the sketch and sketch hashes
func LoadSketch(infile string) (*SketchStore, error) {
	fh, err := os.Open(infile)
	defer fh.Close()
	if err != nil {
		return nil, err
	}
	store := new(SketchStore)
	decoder := gob.NewDecoder(fh)
	err = decoder.Decode(store)
	if store.File != filepath.Base(infile) {
		return nil, fmt.Errorf("file is corrupted (mismatched file names): %v\n", filepath.Base(infile))
	}
	return store, err
}

// Distance is a method to calculate a distance metric between 2 sketches
func (SketchStore *SketchStore) GetDistance(SketchStore2 *SketchStore, metric string) (float64, error) {
	// check that the sketches are compatible
	if err := SketchStore.SketchCheck(SketchStore2); err != nil {
		return 0.0, err
	}
	// convert to float64 slices (for use with goDistances)
	s1 := make([]float64, len(SketchStore.Sketch))
	s2 := make([]float64, len(SketchStore2.Sketch))
	for i := range s1 {
		s1[i] = float64(SketchStore.Sketch[i])
		s2[i] = float64(SketchStore2.Sketch[i])
	}
	// return the required distance
	var distance float64
	var distErr error
	switch metric {
	case "braycurtis":
		bc := new(goDistances.BrayCurtisDistance)
		distance, distErr = bc.Distance(s1, s2)
	case "canberra":
		cd := new(goDistances.CanberraDistance)
		distance, distErr = cd.Distance(s1, s2)
	case "euclidean":
		ed := new(goDistances.EuclideanDistance)
		distance, distErr = ed.Distance(s1, s2)
	case "jaccard":
		intersect := 0.0
		for i := range s1 {
			if s1[i] == s2[i] {
				intersect++
			}
		}
		distance = 1.0 - (intersect / float64(SketchStore.Length))
	default:
		distErr = fmt.Errorf("unknown distance metric: %v\n", metric)
	}
	return distance, distErr
}

// SketchCheck is a method to check that two sketches are compatible
func (SketchStore *SketchStore) SketchCheck(SketchStore2 *SketchStore) error {
	if SketchStore.Length != SketchStore2.Length {
		return fmt.Errorf("these sketches have different lengths: %v and %v\n", SketchStore.File, SketchStore2.File)
	}
	if SketchStore.Dimensions != SketchStore2.Dimensions {
		return fmt.Errorf("these sketches are from histograms with different dimensions: %v and %v\n", SketchStore.File, SketchStore2.File)
	}
	if SketchStore.DistSeed != SketchStore2.DistSeed {
		return fmt.Errorf("these sketches were built with different seeds: %v and %v\n", SketchStore.File, SketchStore2.File)
	}
	return nil
}
