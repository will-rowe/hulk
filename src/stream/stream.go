/*
	the stream package contains a streaming implementation based on the Gopher Academy article by S. Lampa - Patterns for composable concurrent pipelines in Go (https://blog.gopheracademy.com/advent-2015/composable-pipelines-improvements/)
*/
package stream

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/will-rowe/hulk/src/histosketch"
	"github.com/will-rowe/hulk/src/misc"
	"github.com/will-rowe/hulk/src/seqio"
	"github.com/will-rowe/ntHash"
)

const (
	BUFFERSIZE = 128 // buffer size to use for channels
)

// error messages
var (
	shortReadErr = errors.New("found FASTQ read shorter than kSize")
	noDataErr = errors.New("no data received")
)


/*
  The process interface
*/
type process interface {
	Run()
}

/*
  The basic pipeline - takes a list of Processes and runs them in Go routines, the last process is ran in the fg
*/
type Pipeline struct {
	Processes []process
}

func NewPipeline() *Pipeline {
	return &Pipeline{}
}

func (pl *Pipeline) AddProcess(proc process) {
	pl.Processes = append(pl.Processes, proc)
}

func (pl *Pipeline) AddProcesses(procs ...process) {
	for _, proc := range procs {
		pl.AddProcess(proc)
	}
}

func (pl *Pipeline) Run() {
	for i, proc := range pl.Processes {
		if i < len(pl.Processes)-1 {
			go proc.Run()
		} else {
			proc.Run()
		}
	}
}

/*
  A process to stream data from STDIN/file
*/
type DataStreamer struct {
	process
	Output    chan []byte
	InputFile []string
}

func NewDataStreamer() *DataStreamer {
	return &DataStreamer{Output: make(chan []byte, BUFFERSIZE)}
}

func (proc *DataStreamer) Run() {
	var scanner *bufio.Scanner
	// if an input file path has not been provided, scan the contents of STDIN
	if len(proc.InputFile) == 0 {
		scanner = bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			// important: copy content of scan to a new slice before sending, this avoids race conditions (as we are using multiple go routines) from concurrent slice access
			proc.Output <- append([]byte(nil), scanner.Bytes()...)
		}
		if scanner.Err() != nil {
			log.Fatal(scanner.Err())
		}
	} else {
		for i := 0; i < len(proc.InputFile); i++ {
			fh, err := os.Open(proc.InputFile[i])
			misc.ErrorCheck(err)
			defer fh.Close()
			// handle gzipped input
			splitFilename := strings.Split(proc.InputFile[i], ".")
			if splitFilename[len(splitFilename)-1] == "gz" {
				gz, err := gzip.NewReader(fh)
				misc.ErrorCheck(err)
				defer gz.Close()
				scanner = bufio.NewScanner(gz)
			} else {
				scanner = bufio.NewScanner(fh)
			}
			for scanner.Scan() {
				proc.Output <- append([]byte(nil), scanner.Bytes()...)
			}
			if scanner.Err() != nil {
				log.Fatal(scanner.Err())
			}
		}
	}
	close(proc.Output)
}

/*
  A process to generate a FASTQ read from a stream of bytes
*/
type FastqHandler struct {
	process
	Input  chan []byte
	Output chan *seqio.FastqRead
	Fasta  bool
}

func NewFastqHandler() *FastqHandler {
	return &FastqHandler{Output: make(chan *seqio.FastqRead, BUFFERSIZE)}
}

func (proc *FastqHandler) Run() {
	defer close(proc.Output)
	var l1, l2, l3, l4 []byte
	// get the FASTQ/FASTA data from the file
	if proc.Fasta == false {
		for line := range proc.Input {
			if l1 == nil {
				l1 = line
			} else if l2 == nil {
				l2 = line
			} else if l3 == nil {
				l3 = line
			} else if l4 == nil {
				l4 = line
				// create fastq read
				newRead, err := seqio.NewFastqRead(l1, l2, l3, l4)
				if err != nil {
					log.Fatal(err)
				}
				// send on the new read and reset the line stores
				proc.Output <- newRead
				l1, l2, l3, l4 = nil, nil, nil, nil
			}
		}
	} else {
		for line := range proc.Input {
			// check for chevron
			if line[0] == 62 {
				if l1 != nil {
					// store current fasta entry (as FASTQ read)
					l1[0] = 64
					newRead, err := seqio.NewFastqRead(l1, l2, l3, l2)
					if err != nil {
						log.Fatal(err)
					}
					// send on the new read and reset the line stores
					proc.Output <- newRead
				}
				l1, l2 = line, nil
			} else {
				l2 = append(l2, line...)
			}
		}
		// flush final fasta
		l1[0] = 64
		newRead, err := seqio.NewFastqRead(l1, l2, l3, l2)
		if err != nil {
			log.Fatal(err)
		}
		// send on the new read and reset the line stores
		proc.Output <- newRead
	}
}

/*
  A process to check FASTQ reads
*/
type FastqChecker struct {
	process
	Input  chan *seqio.FastqRead
	Output chan *seqio.FastqRead
	Ksize  int
}

func NewFastqChecker() *FastqChecker {
	return &FastqChecker{Output: make(chan *seqio.FastqRead, BUFFERSIZE)}
}

func (proc *FastqChecker) Run() {
	defer close(proc.Output)
	log.Printf("now streaming reads...")
	// count the number of reads and their lengths as we go
	rawCount, lengthTotal := 0, 0
	for read := range proc.Input {
		rawCount++
		if rawCount == 1 {
			if read.Length() < proc.Ksize {
				misc.ErrorCheck(shortReadErr)
			}
		}
		// tally the length so we can report the mean
		lengthTotal += read.Length()
		// send the read on
		proc.Output <- read
	}
	// check we have received reads & print stats
	if rawCount == 0 {
		misc.ErrorCheck(noDataErr)
	}
	log.Printf("\tnumber of reads received from input: %d\n", rawCount)
	meanRL := float64(lengthTotal) / float64(rawCount)
	log.Printf("\tmean read length: %.0f\n", meanRL)
}

/*
  A process to sketch the kmers in reads and populate a histosketch
*/
type Counter struct {
	process
	Input        chan *seqio.FastqRead
	TheCollector chan *histosketch.CountMinSketch
	NumCPU       int
	Ksize        int
	Fasta		bool
	ShredFrac	float64
	Interval     int
	SketchSize   uint
	Spectrum     *histosketch.CountMinSketch
}

func NewCounter() *Counter {
	return &Counter{TheCollector: make(chan *histosketch.CountMinSketch)}
}

func (proc *Counter) Run() {
	// make channels for the minions
	jobs := make(chan []byte)
	var wg sync.WaitGroup
	// set up the minion function
	minion := func(wg *sync.WaitGroup) {
		defer wg.Done()
		// create an initial CountMinSketch to record for this minion
		minionSpectrum := proc.Spectrum.Copy()
		// counter to record the number of processed reads TODO: this isn't used but could be in the future for sampling interval
		readCount := 0
		// collect the reads
		for sequence := range jobs {
			// get hashed kmers from read
			hasher, err := ntHash.New(&sequence, proc.Ksize)
			misc.ErrorCheck(err)
			for hash := range hasher.Hash(true) {
				// add the kmer to the spectrum, (this will return the minimum in the CMS for the eKmer)
				_ = minionSpectrum.Add(hash, 1.0)
			}
			// increment the read counter and send minionSpectrum on if interval reached
			readCount++
			if proc.Interval != 0 && (readCount%proc.Interval) == 0 {
				proc.TheCollector <- minionSpectrum
				minionSpectrum.Wipe()
				readCount = 0
			}
		}
		// send the final spectrum for this minion on to the collector
		proc.TheCollector <- minionSpectrum
		return
	}
	// launch the minions
	wg.Add(proc.NumCPU)
	for w := 1; w <= proc.NumCPU; w++ {
		go minion(&wg)
	}
	// close the process output channel once the minions are finished
	go func() {
		wg.Wait()
		close(proc.TheCollector)
	}()
	// send the reads to the minions
	go func() {
		for read := range proc.Input {
			// shred the read if we have a big FASTA entry
			if proc.Fasta {
				for chunk := range read.Shred(proc.ShredFrac) {
					jobs <- chunk
				}
			} else {
				jobs <- read.Seq()
			}
		}
		close(jobs)
	}()
}

/*
  A process to histosketch a kmer spectrum (stored as a countmin sketch) every interval
*/
type Sketcher struct {
	process
	Input      chan *histosketch.CountMinSketch
	NumCPU     int
	OutFile    string
	SketchSize uint
	MinCount   float64
	DecayRatio float64
	Spectrum   *histosketch.CountMinSketch
	Stream     bool
}

func NewSketcher() *Sketcher {
	return &Sketcher{}
}

func (proc *Sketcher) Run() {
	// create an initial empty histogram, where each bin is a counter position in the CMS (+ an initial 0 bin so we can zero the histosketch)
	emptyHistogram := histosketch.NewHistogram()
	for i := uint32(0); i <= proc.Spectrum.Counters(); i++ {
		_ = emptyHistogram.Add(strconv.Itoa(int(i)), 0)
	}
	// create the empty histoSketch
	hulkSketch := histosketch.NewHistoSketch(proc.SketchSize, emptyHistogram, proc.Spectrum.SizeMB(), proc.DecayRatio)
	// function to histosketch the spectrum
	updateHulk := func() {
		i := uint32(1)
		for cmsCounter := range proc.Spectrum.Dump() {
			// only process counters that have been incremented
			if cmsCounter >= proc.MinCount {
				// each counter position corresponds to a bin in the underlying histogram of the histosketch
				hulkSketch.Update(uint64(i), cmsCounter)
			}
			i++
			// once we have processed all the counters in one hash table, reset the iterator
			if i == (proc.Spectrum.Counters() + 1) {
				//i = uint32(1)
				break
			}
		}
	}
	// range over TheCollector and add each CMS to the combined spectrum
	minionCounter := 0
	for minionSpectrum := range proc.Input {
		minionCounter++
		// merge
		misc.ErrorCheck(proc.Spectrum.Merge(minionSpectrum))
		// if all minions have returned a spectrum, an interval has been reached - print it if there are more minions still working
		if ((minionCounter % proc.NumCPU) == 0) && ((minionCounter / proc.NumCPU) != 1) {
			// update the histosketch
			updateHulk()
			// print sketch to file or STDOUT
			filename := fmt.Sprintf("%v.interval-%d.sketch", proc.OutFile, ((minionCounter / proc.NumCPU) - 1))
			hulkSketch.SaveSketch(filename)
			if proc.Stream {
				fmt.Printf("%v\n", hulkSketch.GetSketch())
			}
			// wipe the combined spectrum
			proc.Spectrum.Wipe()
		}
	}
	// final histosketch update
	updateHulk()
	// encode the histosketch and write to disk - label it as a final sketch if we have used interval sampling
	var filename string
	if minionCounter == proc.NumCPU {
		filename = fmt.Sprintf("%v.sketch", proc.OutFile)
	} else {
		filename = fmt.Sprintf("%v.final.sketch", proc.OutFile)
	}
	hulkSketch.SaveSketch(filename)
	// also print the final sketch to STDOUT if streaming
	if proc.Stream {
		fmt.Printf("%v\n", hulkSketch.GetSketch())
	}
	// TODO: print some more stuff
	log.Printf("\tnumber of sketchers: %d", proc.NumCPU)
}
