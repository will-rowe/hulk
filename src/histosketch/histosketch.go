// Package histosketch is a Go implementation of HistoSketch: Fast Similarity-Preserving Sketching of Streaming Histograms with Concept Drift (https://exascale.info/assets/pdf/icdm2017_HistoSketch.pdf)
// I've made some changes in my implementation compared to the paper:
// - Instead of providing the number of histogram bins (Dimensions) and the number of countmin hash tables (d), I have decided to use epsilon and delta values to calculate CMS Dimensions.
// - As I am using HistoSketch to Sketch CMS counters, the Dimensions of the histosketch are determined by the CMS Dimensions
package histosketch

import (
	"fmt"
	"math"

	rng "github.com/leesper/go_rng"
	"github.com/will-rowe/hulk/src/countmin"
	"github.com/will-rowe/hulk/src/helpers"
)

// MAX_K is the maximum k-mer size currently supported by HULK
const MAX_K uint = 31

// DISTRIBUTION_SEED is used to generate the distributions for the CWS
const DISTRIBUTION_SEED int64 = 1

// CWS is a struct to hold the consistent weighted sampling information
type CWS struct {
	r [][]float64 // r in the paper
	c [][]float64 // c in the paper
	b [][]float64 // beta in the paper (adapted here to pre-calculate beta*r )
}

// getSample is a method to yield A_ka from the CWS, given the incoming histogram bin and the current sketch position
func (CWS *CWS) getSample(i uint, j int, freq float64) float64 {
	Yka := math.Exp(math.Log(freq) - CWS.b[j][i])
	return CWS.c[j][i] / (Yka * math.Exp(CWS.r[j][i]))
}

// HistoSketch is the histosketch data structure
type HistoSketch struct {
	algorithm         string
	KmerSize          uint                     `json:"ksize"`              // the size of the k-mer used in the histosketch
	Md5sum            string                   `json:"md5sum"`             // md5sum of the sketch
	Sketch            []uint                   `json:"mins"`               // S in paper
	SketchWeights     []float64                `json:"weights"`            // A in paper
	SketchSize        uint                     `json:"num"`                // number of minimums in the histosketch
	Dimensions        int32                    `json:"num_histogram_bins"` // number of histogram bins
	ApplyConceptDrift bool                     `json:"concept_drift"`      // if true, uniform scaling will be applied to frequency estimates (in the CMS) and a decay ratio will be applied to sketch elements prior to assessing incoming elements
	cwsSamples        *CWS                     // the consistent weighted samples
	cmSketch          *countmin.CountMinSketch // Q in the paper (d * g matrix, where g is Sketch length)
}

// NewHistoSketch is the constructor function
func NewHistoSketch(kmerSize, histosketchLength uint, numHistogramBins int32, decayRatio float64) (*HistoSketch, error) {

	// run some basic checks
	if kmerSize > MAX_K {
		return nil, fmt.Errorf("histosketching only supports k <= %d", MAX_K)
	}
	checkFloat := func(f float64) bool {
		if f < 0.0 || f > 1.0 {
			return false
		}
		return true
	}
	if !checkFloat(decayRatio) {
		return nil, fmt.Errorf("decay ratio must be between 0.0 and 1.0")
	}
	if numHistogramBins < 2 {
		return nil, fmt.Errorf("histogram must have at least 2 bins")
	}

	// create the histosketch data structure
	newHistosketch := &HistoSketch{
		algorithm:     "histosketch",
		KmerSize:      kmerSize,
		Sketch:        make([]uint, histosketchLength),
		SketchWeights: make([]float64, histosketchLength),
		SketchSize:    histosketchLength,
		Dimensions:    numHistogramBins,
		cmSketch:      countmin.NewCountMinSketch(countmin.EPSILON, countmin.DELTA, decayRatio),
	}
	if decayRatio != 1.0 {
		newHistosketch.ApplyConceptDrift = true
	}

	// zero the HistoSketch sketch and weights
	for i := range newHistosketch.Sketch {
		newHistosketch.Sketch[i] = 0
		newHistosketch.SketchWeights[i] = math.MaxFloat64
	}

	// create the CWS samples and attach them to the data structure
	newHistosketch.newCWS()
	return newHistosketch, nil
}

// newCWS is a method to generate a set of Consistent Weighted Samples
func (HistoSketch *HistoSketch) newCWS() {

	// create the matrices
	r := make([][]float64, HistoSketch.SketchSize)
	c := make([][]float64, HistoSketch.SketchSize)
	b := make([][]float64, HistoSketch.SketchSize)

	// set up the CWS by taking 3 sets of samples: from a Gamma distribution, log Gamma distribution and a uniform distribution respectively
	gammaGenerator := rng.NewGammaGenerator(DISTRIBUTION_SEED)     // a random number generator for gamma distribution
	uniformGenerator := rng.NewUniformGenerator(DISTRIBUTION_SEED) // a random number generator for a uniform distribution

	// create the samples
	for i := uint(0); i < HistoSketch.SketchSize; i++ {
		r[i] = make([]float64, HistoSketch.Dimensions)
		c[i] = make([]float64, HistoSketch.Dimensions)
		b[i] = make([]float64, HistoSketch.Dimensions)
		for j := int32(0); j < HistoSketch.Dimensions; j++ {
			r[i][j] = gammaGenerator.Gamma(2, 1)
			c[i][j] = math.Log(gammaGenerator.Gamma(2, 1))
			//b[i][j] = uniformGenerator.Float64Range(0, 1) // as in paper
			// I've multiplied beta by r and stored this instead of just beta
			b[i][j] = uniformGenerator.Float64Range(0, 1) * r[i][j]
		}
	}

	// set the cwsSamples
	HistoSketch.cwsSamples = &CWS{
		r: r,
		c: c,
		b: b,
	}
}

// AddElement is a method to assess an incoming histogram element and add it to the histosketch if required
func (HistoSketch *HistoSketch) AddElement(bin uint64, value float64) error {

	// add the incoming element to the persistent countmin sketch
	estiFreq := HistoSketch.cmSketch.Add(bin, value) // the countmin sketch will apply uniform scaling (if decayRatio!= 0.0), prior to adding the new element

	// use consistent weighted sampling to determine if the incoming element should be added to a histosketch slot
	for histosketchSlot := range HistoSketch.Sketch {

		// get the CWS value (A_ka) for the incoming element
		Aka := HistoSketch.cwsSamples.getSample(uint(bin), histosketchSlot, estiFreq)

		// get the current minimum in the histosketchSlot, accounting for concept drift if requrested
		var curMin float64
		if HistoSketch.ApplyConceptDrift {
			curMin = HistoSketch.SketchWeights[histosketchSlot] / HistoSketch.cmSketch.GetDecayWeight()
		} else {
			curMin = HistoSketch.SketchWeights[histosketchSlot]
		}

		// if A_ka is a new minimum, replace both the bin and the weight held at this slot in the histosketch
		if Aka < curMin {
			HistoSketch.Sketch[histosketchSlot] = uint(bin)
			HistoSketch.SketchWeights[histosketchSlot] = Aka
		}
	}
	return nil
}

// GetSketch is a method to return the current histosketch
func (HistoSketch *HistoSketch) GetSketch() []uint64 {
	sketch := make([]uint64, len(HistoSketch.Sketch))
	for i, val := range HistoSketch.Sketch {
		sketch[i] = uint64(val)
	}
	return sketch
}

// SetMD5 is a method to calculate and store the MD5 for the histosketch
func (HistoSketch *HistoSketch) SetMD5() {
	HistoSketch.Md5sum = fmt.Sprintf("%x", helpers.MD5sum(HistoSketch.GetSketch()))
	return
}

// GetMD5 is a method to return the MD5 currently calculated for the histosketch
func (HistoSketch *HistoSketch) GetMD5() string {
	return HistoSketch.Md5sum
}

// GetAlgo is a method to return the sketching algorithm used
func (HistoSketch *HistoSketch) GetAlgo() string {
	return HistoSketch.algorithm
}
