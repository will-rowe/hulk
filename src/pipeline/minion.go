package pipeline

import (
	"sync"

	"github.com/will-rowe/hulk/src/helpers"
	"github.com/will-rowe/hulk/src/minimizer"
)

// Minion is the base data type
type Minion struct {
	sync.RWMutex
	id            int
	info          *Info
	minionQueue   chan chan []byte
	inputChannel  chan []byte
	outputChannel chan uint64
	stop          chan struct{}
}

// newMinion is the constructor function
func newMinion(id int, runtimeInfo *Info, minionQueue chan chan []byte, returnChan chan uint64) *Minion {
	return &Minion{
		id:            id,
		info:          runtimeInfo,
		minionQueue:   minionQueue,
		inputChannel:  make(chan []byte),
		outputChannel: returnChan,
		stop:          make(chan struct{}),
	}
}

// Start is a method to start the minion running
func (minion *Minion) Start() {
	go func() {
		for {

			// when the minion is available for work, place its data channel in the queue
			minion.minionQueue <- minion.inputChannel

			// wait for work or stop signal
			select {

			// the minion has receieved some data from the boss
			case data := <-minion.inputChannel:

				// make sure the boss knows work is happening, incase a finish signal is sent
				minion.Lock()

				// get the minimizers for this sequence
				sketch, err := minimizer.NewMinimizerSketch(minion.info.Sketch.KmerSize, minion.info.Sketch.WindowSize, data)
				helpers.ErrorCheck(err)

				// send minimizers back to the boss
				for minimizer := range sketch.GetMinimizers() {
					minion.outputChannel <- minimizer.(uint64)
				}

				// this minion is done for now
				minion.Unlock()

			// end the minion go function if a stop signal has been sent
			case <-minion.stop:
				return
			}
		}
	}()
}

// Finish is a method to close down a minion, after checking it isn't currently working on something
func (minion *Minion) Finish() {
	close(minion.stop)
	return
}
