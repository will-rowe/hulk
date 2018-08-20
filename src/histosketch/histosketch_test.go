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
	t.Log(hs.GetSketch())
}

// test the Update method
func Test_Update(t *testing.T) {
	h := createHistogram()
	// histoSketch of 24 minimums
	hs := NewHistoSketch(24, h, EPSILON, DELTA, DR)
	t.Log(hs.GetSketch())
	// update the histoSketch
	hs.Update(2, 5.3434)
	t.Log(hs.GetSketch())
	hs.Update(2, 200.2324)
	t.Log(hs.GetSketch())
	hs.Update(1, 49.24353553)
	t.Log(hs.GetSketch())
}

// TODO: test SketchStore

/*
// test the GetDistance method on SketchStore
func Test_GetDistance(t *testing.T) {
	h := createHistogram()
	// histoSketch of 24 minimums
	hs1 := NewHistoSketch(24, h, EPSILON, DELTA, DR)
	hs2 := NewHistoSketch(24, h, EPSILON, DELTA, DR)
	// get jaccard distance
	if dist, err := hs1.GetDistance(hs2, "jaccard"); err != nil {
		if dist != 0 {
			t.Fatal("jaccard distance should be 0!")
		}
	} else {
		t.Fatal("could not get Jaccard Distance!")
	}
	// unrecognised distance
	if dist, err := hs1.GetDistance(hs2, "hellinger"); err == nil {
		t.Fatal("should through unrecognised distance metric error")
	}
}
*/
