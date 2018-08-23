package histosketch

import (
	"encoding/binary"
	"math"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// PlotHistogram method TODO: this is just an idea and I don't know if it will be useful yet
func (HistoSketch *HistoSketch) PlotHistogram() {
	v := make(plotter.Values, HistoSketch.length)
	for i := range v {
		v[i] = HistoSketch.sketchWeights[i]
	}
	// Make a plot and set its title.
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "Histogram"
	// Create a histogram
	h, err := plotter.NewHist(v, int(HistoSketch.length))
	if err != nil {
		panic(err)
	}
	// Normalize the area under the histogram to
	// sum to one.
	h.Normalize(1)
	p.Add(h)
	// Save the plot to a PNG file.
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "hist.png"); err != nil {
		panic(err)
	}
}

// float64ToBytes conversion
func float64ToBytes(f float64) []byte {
	bits := math.Float64bits(f)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}

// float64FromBytes conversion
func float64FromBytes(b []byte) float64 {
	bits := binary.LittleEndian.Uint64(b)
	f := math.Float64frombits(bits)
	return f
}
