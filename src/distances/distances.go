// Package distances contains the distance calculations for several common metrics
package distances

import (
	"fmt"
	"math"

	"github.com/JSchwehn/goDistances"
)

// GetDistance is a function to calculate an unweighted distance metric between 2 sets of floats
func GetDistance(setA, setB []float64, metric string) (float64, error) {
	var distVal float64
	var distErr error
	if len(setA) != len(setB) {
		return 0.0, fmt.Errorf("set size mismatch: %d vs %d\n", len(setA), len(setB))
	}
	switch metric {
	case "jaccard":
		intersect := 0.0
		for i := range setA {
			if setA[i] == setB[i] {
				intersect++
			}
		}
		distVal = 1.0 - (intersect / float64(len(setA)))
	case "braycurtis":
		bc := new(goDistances.BrayCurtisDistance)
		distVal, distErr = bc.Distance(setA, setB)
	case "canberra":
		cd := new(goDistances.CanberraDistance)
		distVal, distErr = cd.Distance(setA, setB)
	case "euclidean":
		ed := new(goDistances.EuclideanDistance)
		distVal, distErr = ed.Distance(setA, setB)
	default:
		distErr = fmt.Errorf("unknown distance metric: %v\n", metric)
	}
	return distVal, distErr
}

// GetWJD is a function to calculate the weighted jaccard distance between two sets
// <http://theory.stanford.edu/~sergei/papers/soda10-jaccard.pdf>
func GetWJD(setA, setB, weightsA, weightsB []float64) (float64, error) {
	intersect, union := 0.0, 0.0
	for i := uint(0); i < uint(len(setA)); i++ {

		// get the weight pair and select the largest value
		weightA := math.Max(math.Max(weightsA[i], 0), math.Max(-weightsA[i], 0))
		weightB := math.Max(math.Max(weightsB[i], 0), math.Max(-weightsB[i], 0))

		// get the intersection and union values
		if setA[i] == setB[i] {
			if weightA < weightB {
				intersect += weightA
				union += weightB
			} else {
				intersect += weightB
				union += weightA
			}
		} else {
			if weightA > weightB {
				union += weightA
			} else {
				union += weightB
			}
		}
	}

	// return the weighted jaccard distance
	return 1 - (intersect / union), nil
}
