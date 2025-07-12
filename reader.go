package dst

import (
	"io"
)

type FaultyReader struct {
        reader io.Reader
        faults []Fault
}

func NewFaultyReader(reader io.Reader) *FaultyReader {
        return &FaultyReader{
                reader,
                make([]Fault, 0, 10),
        }
}

func (r *FaultyReader) Read(p []byte) (int, error) {
        return r.Read(p)
}

func (r *FaultyReader) AddFault(f Fault) {
        r.faults = append(r.faults, f)
} 
