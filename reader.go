package simtest

import (
	"io"
	"math/rand/v2"
)

// SimReader is a simulated io.Reader that can randomly inject faults
// into the data being read.
type SimReader struct {
        prng *rand.Rand
        clock *SimClock
        reader io.Reader
        faults []Fault
}

// Read reads some data into p. It returns the number of 
// bytes read and any error encountered. If Read reached
// the end of the file then the returned error is io.EOF.
func (r *SimReader) Read(p []byte) (int, error) {
        var n int
        var err error

        n, err = r.reader.Read(p)
        // inject faults

        return n, err
}

func (r *SimReader) AddFault(f Fault) {
        r.faults = append(r.faults, f)
} 
