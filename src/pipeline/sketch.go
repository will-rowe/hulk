package pipeline

/*
 this part of the pipeline will process sequences, decompose to minimizers and sketch them
*/

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/will-rowe/hulk/src/helpers"
	"github.com/will-rowe/hulk/src/histosketch"
	"github.com/will-rowe/hulk/src/kmerspectrum"
	"github.com/will-rowe/hulk/src/seqio"
	"github.com/will-rowe/hulk/src/sketchio"
)

// DataStreamer is a pipeline process that streams data from STDIN/file
type DataStreamer struct {
	info   *Info
	input  []string
	output chan []byte
}

// NewDataStreamer is the constructor
func NewDataStreamer(info *Info) *DataStreamer {
	return &DataStreamer{info: info, output: make(chan []byte, BUFFERSIZE)}
}

// Connect is the method to connect the DataStreamer to some data source
func (proc *DataStreamer) Connect(input []string) {
	proc.input = input
}

// Run is the method to run this process, which satisfies the pipeline interface
func (proc *DataStreamer) Run() {
	defer close(proc.output)
	var scanner *bufio.Scanner

	// if an input file path has not been provided, scan the contents of STDIN
	if len(proc.input) == 0 {
		scanner = bufio.NewScanner(os.Stdin)
		for scanner.Scan() {

			// important: copy content of scan to a new slice before sending, this avoids race conditions (as we are using multiple go routines) from concurrent slice access
			proc.output <- append([]byte(nil), scanner.Bytes()...)
		}
		if scanner.Err() != nil {
			log.Fatal(scanner.Err())
		}
	} else {
		for i := 0; i < len(proc.input); i++ {
			fh, err := os.Open(proc.input[i])
			helpers.ErrorCheck(err)
			defer fh.Close()

			// handle gzipped input
			splitFilename := strings.Split(proc.input[i], ".")
			if splitFilename[len(splitFilename)-1] == "gz" {
				gz, err := gzip.NewReader(fh)
				helpers.ErrorCheck(err)
				defer gz.Close()
				scanner = bufio.NewScanner(gz)
			} else {
				scanner = bufio.NewScanner(fh)
			}
			for scanner.Scan() {
				proc.output <- append([]byte(nil), scanner.Bytes()...)
			}
			if scanner.Err() != nil {
				log.Fatal(scanner.Err())
			}
		}
	}
}

// FastqHandler is a pipeline process to convert a pipeline to the FASTQ type
type FastqHandler struct {
	info   *Info
	input  chan []byte
	output chan *seqio.FASTQread
}

// NewFastqHandler is the constructor
func NewFastqHandler(info *Info) *FastqHandler {
	return &FastqHandler{info: info, output: make(chan *seqio.FASTQread, BUFFERSIZE)}
}

// Connect is the method to join the input of this process with the output of a DataStreamer
func (proc *FastqHandler) Connect(previous *DataStreamer) {
	proc.input = previous.output
}

// Run is the method to run this process, which satisfies the pipeline interface
func (proc *FastqHandler) Run() {
	defer close(proc.output)
	var l1, l2, l3, l4 []byte
	if proc.info.Sketch.Fasta {
		for line := range proc.input {
			if len(line) == 0 {
				break
			}
			// check for chevron
			if line[0] == 62 {
				if l1 != nil {

					// store current fasta entry (as FASTQ read)
					l1[0] = 64
					newRead, err := seqio.NewFASTQread(l1, l2, nil, nil)
					if err != nil {
						log.Fatal(err)
					}

					// send on the new read and reset the line stores
					proc.output <- newRead
				}
				l1, l2 = line, nil
			} else {
				l2 = append(l2, line...)
			}
		}

		// flush final fasta
		l1[0] = 64
		newRead, err := seqio.NewFASTQread(l1, l2, nil, nil)
		if err != nil {
			log.Fatal(err)
		}

		// send on the new read and reset the line stores
		proc.output <- newRead
	} else {

		// grab four lines and create a new FASTQread struct from them - perform some format checks and trim low quality bases
		for line := range proc.input {
			if l1 == nil {
				l1 = line
			} else if l2 == nil {
				l2 = line
			} else if l3 == nil {
				l3 = line
			} else if l4 == nil {
				l4 = line

				// create fastq read
				newRead, err := seqio.NewFASTQread(l1, l2, l3, l4)
				if err != nil {
					log.Fatal(err)
				}

				// send on the new read and reset the line stores
				proc.output <- newRead
				l1, l2, l3, l4 = nil, nil, nil, nil
			}
		}
	}
}

// SeqMinimizer is a process to collect minimizers from sequences
type SeqMinimizer struct {
	info     *Info
	input    chan *seqio.FASTQread
	output   chan *kmerspectrum.Bin // the pair is used to hold a minimzer and its frequency
	sketches []sketchio.SketchObject
}

// NewSeqMinimizer is the constructor
func NewSeqMinimizer(info *Info) *SeqMinimizer {
	return &SeqMinimizer{info: info, output: make(chan *kmerspectrum.Bin, BUFFERSIZE)}
}

// Connect is the method to join the input of this process with the output of FastqHandler
func (proc *SeqMinimizer) Connect(previous *FastqHandler) {
	proc.input = previous.output
}

// Run is the method to run this process, which satisfies the pipeline interface
func (proc *SeqMinimizer) Run() {
	log.Printf("finding minimizers...")

	// count the number of sequences and their lengths as we go
	seqCount := uint(0)
	reportInterval := uint(100000)
	multiplier := uint(1)
	lengthTotal := 0

	// set up the boss and minion pool, ready to find minimizers
	theBoss, err := findMinimizers(proc.output, proc.info)
	helpers.ErrorCheck(err)

	// start processing sequences
	sketchingInterval := 0
	for sequence := range proc.input {

		// add the seq to the queue for minimizer finding
		theBoss.AddSeq(sequence.Seq)

		// print progress to screen
		seqCount++
		if (seqCount % 100000) == 0 {
			log.Printf("\tprocessed %d sequences", reportInterval*multiplier)
			multiplier++
		}
		lengthTotal += len(sequence.Seq)

		// if an interval is reached, get the minions to flush their k-mer spectra, sending the data to the next pipeline process
		if (proc.info.Sketch.Interval) != 0 && (seqCount%proc.info.Sketch.Interval) == 0 {
			sketchingInterval++
			log.Printf("\treached interval %d -> histosketching", sketchingInterval)
			theBoss.Flush()
		}

	} // all sequences have been sent for processing

	// final flush of the minions
	log.Printf("generating final histosketch of k-mer spectra...")
	theBoss.Flush()

	// signal the end of the sequences and close the channels
	theBoss.StopWork()

	// collect the secondary sketches if applicable TODO: add this functionality back in
	if proc.info.Sketch.KMV {
		sketch := theBoss.CollectKMVsketch()
		proc.sketches = append(proc.sketches, sketch)
	}
	if proc.info.Sketch.KHF {
		sketch := theBoss.CollectKHFsketch()
		proc.sketches = append(proc.sketches, sketch)
	}

	// check we received some sequence data & print some info
	if seqCount == 0 {
		helpers.ErrorCheck(fmt.Errorf("no sequences received"))
	}
	meanRL := uint(float64(lengthTotal) / float64(seqCount))
	log.Printf("\tprocessed %d sequences in total\n", seqCount)
	log.Printf("\tmean sequence length: %d\n", meanRL)
	log.Printf("\tfound %d minimizers\n", theBoss.GetMinimizerCount())
	log.Printf("\thistosketching across %d bins\n", proc.info.Sketch.SpectrumSize)
	if proc.info.Sketch.NumMinions > 1 {
		log.Printf("merging sketches and cleaning up...")
	} else {
		log.Printf("cleaning up...")
	}
}

// Sketcher is a pipeline process that receives k-mer spectra data from minions and histosketches it
type Sketcher struct {
	info     *Info
	input    chan *kmerspectrum.Bin
	sketches *[]sketchio.SketchObject
}

// NewSketcher is the constructor
func NewSketcher(info *Info) *Sketcher {
	return &Sketcher{info: info}
}

// Connect is the method to join the input of this process with the output of SeqMinimizer
func (proc *Sketcher) Connect(previous *SeqMinimizer) {
	proc.input = previous.output
	proc.sketches = &previous.sketches
}

// Run is the method to run this process, which satisfies the pipeline interface
func (proc *Sketcher) Run() {

	// create the HULKdata object, ready to store the sketches
	hulkData := sketchio.NewHULKdata()

	// create the histosketch
	hs, err := histosketch.NewHistoSketch(proc.info.Sketch.KmerSize, proc.info.Sketch.SketchSize, proc.info.Sketch.SpectrumSize, proc.info.Sketch.DecayRatio)
	helpers.ErrorCheck(err)

	// collect the k-mer spectra data from minions and histosketch it
	for bin := range proc.input {

		// TODO: change histosketch to accept int32 as binID
		hs.AddElement(uint64(bin.BinID), bin.Frequency)
	}

	// once we get here, the previous process has finished and we are ready to save all the HULK data
	// add the histosketch to the HULKdata
	helpers.ErrorCheck(hulkData.Add(hs))

	// add any other sketches we asked the previous process for
	for _, sketch := range *proc.sketches {
		helpers.ErrorCheck(hulkData.Add(sketch))
	}

	// add any final info to the HULKdata before writing the sketch to disk
	hulkData.Banner = proc.info.Sketch.BannerLabel
	hulkData.WriteJSON(proc.info.Sketch.OutFile + ".json")
	log.Printf("\twritten sketch to disk: %v\n", proc.info.Sketch.OutFile+".json")
}
