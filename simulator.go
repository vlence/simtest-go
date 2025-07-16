package simtest

import (
	"io"
	"math/rand/v2"
)

// A Simulator can deterministically simulate I/O operations
// and time. Simulators can be used to deterministically inject
// faults into applications and find bugs.
//
// The usefulness of this simulator depends on the dependency
// injection style of programming. It is common practice to
// inject the dependencies of functions/methods as their
// arguments. This way it is easy to test them. The simulator
// works at a much lower level, dealing with io.Reader and
// io.Writer values, and time.
//
// The difficulty in simulating time mainly comes from the fact
// that it's not really thought of as a dependency in general.
// Time is assumed to be always just there. By thinking of it
// as a dependency it is possible to simulate clocks, timers
// and tickers. This allows us to speed up and slow down the
// the time passing within our applications.
type Simulator struct {
        // Source of randomness
        src rand.Source

        // Deterministic psuedo-random number generator
        prng *rand.Rand

        // simulated clock
        clock *SimClock
}

// NewSimulator creates and returns a new simulator using the given
// PRNG src and clock.
func NewSimulator(src rand.Source, clock *SimClock) *Simulator {
        return &Simulator{
                src,
                rand.New(src),
                clock,
        }
}

// NewReader returns a *SimReader. Simulated readers inject
// random faults into the data being read.
func (sim *Simulator) NewReader(r io.Reader) *SimReader {
        reader := new(SimReader)
        reader.reader = r
        reader.prng = sim.prng
        reader.clock = sim.clock

        return reader
}

// NewWriter returns a *SimWriter. Simulated writers inject
// random faults into the data being written.
func (sim *Simulator) NewWriter(w io.Writer) {}
