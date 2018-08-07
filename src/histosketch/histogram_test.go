package histosketch

import (
	"testing"
)

// test the histograms
func Test_histogram(t *testing.T) {
	// create histogram
	h := NewHistogram()
	// add a value
	if err := h.Add("first bin", 1.234); err != nil {
		t.Fatal(err)
	}
	// add a second value
	if err := h.Add("second bin", 2.345); err != nil {
		t.Fatal(err)
	}
	// try adding a duplicate label
	if err := h.Add("first bin", 1.234); err == nil {
		t.Fatal(err)
	}
	// dump the histogram to screen
	h.Dump()
}
