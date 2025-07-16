package simtest

import (
	"sync"
	"time"

	"github.com/vlence/gossert"
)

type SimTicker struct{
        C <-chan time.Time
        d time.Duration
        mu *sync.RWMutex
        clock *SimClock
        stopped bool
        deadline time.Time
}

func newSimTicker(d time.Duration, ch <-chan time.Time, clock *SimClock) *SimTicker {
        ticker := new(SimTicker)
        ticker.C = ch
        ticker.d = d
        ticker.mu = new(sync.RWMutex)
        ticker.clock = clock
        ticker.stopped = false
        ticker.deadline = clock.Now().Add(d)

        return ticker
}

func (ticker *SimTicker) Reset(d time.Duration) {
        ticker.mu.Lock()
        defer ticker.mu.Unlock()

        ticker.d = d
        ticker.deadline = ticker.clock.Now().Add(d)
}

func (ticker *SimTicker) Stop() {
        ticker.mu.Lock()
        defer ticker.mu.Unlock()

        if ticker.stopped {
                return
        }

        ticker.stopped = true
        ticker.clock.removeTicker(ticker)
        _, ok := ticker.clock.tickers[ticker]
        gossert.Ok(!ok, "simticker: ticker was not removed from tickers list")
}
