package dst

import "time"

// Fault represents a fault that occurs during reading or writing
// data.
type Fault interface {
        // Prob returns the probability of this fault occuring
        Prob() float64

        // Inject applies the fault to the bytes p. The function signature
        // is deliberately similar to the signature of the Read and Write
        // methods of the io.Reader and io.Writer interfaces. The return
        // values of this method should be returned by FaultyReader.Read
        // and FaultyWriter.Write.
        //
        // If injecting a fault while reading, Inject should be called
        // after the underlying data has been read and before the data
        // has been returned to the caller. Similarly Inject should be
        // called before writing the data using the underlying io.Writer.
        Inject(p []byte) (int, error)
}

// LatencyFault injects some latency while reading or writing bytes.
type LatencyFault struct {
        min time.Duration
        max time.Duration
        prob float64
}

func NewHighLatencyFault(min, max time.Duration, prob float64) *LatencyFault {
        return &LatencyFault{
                min,
                max,
                prob,
        }
}

func (fault *LatencyFault) Prob() float64 {
        return fault.prob
}

func (fault *LatencyFault) Inject(p []byte) (int, error) {
        // wait for some time before returning
        // time.Sleep(fault.latency)
        return len(p), nil
}
