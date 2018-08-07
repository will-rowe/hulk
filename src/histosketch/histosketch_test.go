package histosketch

import (
	"testing"
)

var (
	labels = []string{"w", "x", "y", "z"}
	values = []float64{1.2345, 2.3456, 3.4567, 4.5678}
)

// helper function to create a histogram
func createHistogram() *histogram {
	h := NewHistogram()
	for i, label := range labels {
		_ = h.Add(label, values[i])
	}
	return h
}

// test the histoSketch initialisation
func Test_NewHistoSketch(t *testing.T) {
	h := createHistogram()
	// histoSketch of 24 minimums
	hs := NewHistoSketch(24, h, EPSILON, DELTA, DR)
	// print the sketch and the hash values
	sketch, hashes := hs.GetSketches()
	t.Log(sketch, hashes)
}

// test the Update method
func Test_Update(t *testing.T) {
	h := createHistogram()
	// histoSketch of 24 minimums
	hs := NewHistoSketch(24, h, EPSILON, DELTA, DR)
	t.Log(hs.GetSketches())
	// update the histoSketch
	hs.Update(2, 5.3434)
	t.Log(hs.GetSketches())
	hs.Update(2, 200.2324)
	t.Log(hs.GetSketches())
	hs.Update(1, 49.24353553)
	t.Log(hs.GetSketches())
}
