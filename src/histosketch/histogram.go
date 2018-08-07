package histosketch

import (
	"fmt"
)

// histogram base type
type histogram struct {
	bins     []*histogramElement
	labels   []string
	dupCheck map[string]struct{}
}

// histogramElement base type
type histogramElement struct {
	bin   int // bin corresponds to label (bin = labels index)
	value float64
}

// NewHistogram is a histogram constructor
func NewHistogram() *histogram {
	return &histogram{
		bins:     []*histogramElement{},
		labels:   []string{},
		dupCheck: make(map[string]struct{}),
	}
}

// Add method will add a label (l) and a value (v) to a histogram as a new bin.
// bins are added sequentially and bin number is linked to the label
func (histogram *histogram) Add(l string, v float64) error {
	// check this label hasn't been added yet
	if _, ok := histogram.dupCheck[l]; !ok {
		histogram.dupCheck[l] = struct{}{}
	} else {
		return fmt.Errorf("label already exists in histogram")
	}
	histogram.labels = append(histogram.labels, l)
	histogram.bins = append(histogram.bins, &histogramElement{
		bin:   len(histogram.bins),
		value: v,
	})
	return nil
}

// Dump method will print the contents of the histogram, each bin will be a new line
func (histogram *histogram) Dump() {
	for i, bin := range histogram.bins {
		fmt.Printf("bin: %d\tlabel: %s\tvalue: %.4f\n", bin.bin, histogram.labels[i], bin.value)
	}
}
