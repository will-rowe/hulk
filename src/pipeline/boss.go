package pipeline

import (
	"github.com/will-rowe/hulk/src/helpers"
	"github.com/will-rowe/hulk/src/kmerspectrum"
	"github.com/will-rowe/hulk/src/minhash"
)

// theBoss is used to orchestrate the workers
type theBoss struct {
	inputSequences   chan []byte                // the boss uses this channel to receive sequence data from the main sketching pipeline
	theCollector     chan *kmerspectrum.Bin     // the boss uses this channel to send minimizer frequency data back to the main sketching pipeline
	minimizerChan    chan uint64                // minions send minimizers down this channel, back to the boss
	flush            chan bool                  // controls flushing of the minions
	finish           chan bool                  // the boss uses this channel to stop the minions
	minionRegister   []*Minion                  // a slice of all the minions controlled by this boss
	kmerSpectrum     *kmerspectrum.KmerSpectrum // the boss stores the minimizer frequencies in a k-mer spectrum
	kmvSketch        *minhash.KMVsketch         // optional sketch (not used yet)
	khfSketch        *minhash.KHFsketch         // optional sketch (not used yet)
	minimizerCounter int                        // a count of the minimizers the Boss has collected
}

// AddSeq is a method to give the boss a sequence
func (theBoss *theBoss) AddSeq(seq []byte) {
	theBoss.inputSequences <- seq
}

// StopWork is a method to initiate a controlled shut down of the boss and minions
func (theBoss *theBoss) StopWork() {
	theBoss.finish <- true
}

// Flush is a method to flush the current value held in the k-mer spectrum and then wipe it
func (theBoss *theBoss) Flush() {
	theBoss.flush <- true
}

// GetMinimizerCount is a method to return the number of minimizers the boss has collected
func (theBoss *theBoss) GetMinimizerCount() int {
	return theBoss.minimizerCounter
}

// CollectKHFsketch is a method to collect the KHF sketch
func (theBoss *theBoss) CollectKHFsketch() *minhash.KHFsketch {
	return theBoss.khfSketch
}

// CollectKMVsketch is a method to collect the KMV sketch
func (theBoss *theBoss) CollectKMVsketch() *minhash.KMVsketch {
	return theBoss.kmvSketch
}

// findMinimizers is a function to start off the minions to find minimizers, returning their boss
func findMinimizers(returnChannel chan *kmerspectrum.Bin, runtimeInfo *Info) (*theBoss, error) {

	// set up the base k-mer spectrum
	ks, err := kmerspectrum.NewKmerSpectrum(int32(runtimeInfo.Sketch.SpectrumSize))
	if err != nil {
		return nil, err
	}

	// create a boss to orchestrate the minions
	boss := &theBoss{
		inputSequences: make(chan []byte),
		theCollector:   returnChannel,
		minimizerChan:  make(chan uint64),
		finish:         make(chan bool),
		flush:          make(chan bool),
		kmerSpectrum:   ks,
		kmvSketch:      minhash.NewKMVsketch(runtimeInfo.Sketch.KmerSize, runtimeInfo.Sketch.SketchSize),
		khfSketch:      minhash.NewKHFsketch(runtimeInfo.Sketch.KmerSize, runtimeInfo.Sketch.SketchSize),
	}

	// set up the minion pool
	minionQueue := make(chan chan []byte)
	boss.minionRegister = make([]*Minion, runtimeInfo.Sketch.NumMinions)
	for id := 0; id < runtimeInfo.Sketch.NumMinions; id++ {

		// create a minion
		minion := newMinion(id, runtimeInfo, minionQueue, boss.minimizerChan)

		// start it running
		minion.Start()

		// add it to the boss's register of running minions
		boss.minionRegister[id] = minion
	}

	// start collecting the minimizers from the minions
	go func() {
		for minimizer := range boss.minimizerChan {
			boss.kmerSpectrum.AddHash(minimizer)
			boss.minimizerCounter++
		}
	}()

	// start processing the sequences
	go func() {
		for {
			select {

			// if there's a sequence to be processed, send it to a minion
			case sequence := <-boss.inputSequences:

				// wait for a minion to be available
				minion := <-minionQueue

				// hand the sequence over
				minion <- sequence

			// flush the boss's k-mer spectrum - which sends it back to the main pipeline sketching process and then wipes it
			case <-boss.flush:

				// TODO: pause the minimizer chan before doing this....

				// send the minimizers and frequencies to the main pipeline sketching process
				if boss.kmerSpectrum.Cardinality() != 0 {
					dump, err := boss.kmerSpectrum.Dump()
					helpers.ErrorCheck(err)
					for bin := range dump {
						if bin.Frequency != 0.0 {
							boss.theCollector <- bin
						}
					}

					// wipe the minion's spectrum, ready to collect more k-mers
					boss.kmerSpectrum.Wipe()
				}

			// stop the minions working when the boss receives word
			case <-boss.finish:

				// send the finish signal to the minions
				for _, minion := range boss.minionRegister {
					minion.Finish()
				}

				// close the channel sending sequences to the minions
				close(boss.inputSequences)

				// close the channel receiving minimizers from the minions
				close(boss.minimizerChan)

				// close the channel sending minimizer frequencies to the main pipeline
				close(boss.theCollector)

				return
			}
		}
	}()

	return boss, nil
}
