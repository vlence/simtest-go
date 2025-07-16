package simtest

import (
	"sync"
	"time"

	"github.com/vlence/gossert"
)

// A Ticker represents an action that needs to be executed at intervals.
// The interface is deliberately kept similar to that of *time.Ticker.
type Ticker interface {
        Reset(d time.Duration)

        Stop()
}

// SimTicker represents a simulated ticker. A ticker ticks at the
// given interval and can be used to do something whenever it ticks.
// Since SimTicker uses SimClock internally the time is simulated
// ticks are not fired unless SimClock is Tick'ed manually.
type SimTicker struct{
        C <-chan time.Time
        d time.Duration
        mu *sync.RWMutex
        clock *SimClock
        stopped bool
        deadline time.Time
}

// newSimTicker returns a new simulated ticker.
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

// Reset resets this ticker and fires the next tick
// after d time has passed since Reset was called.
func (ticker *SimTicker) Reset(d time.Duration) {
        ticker.mu.Lock()
        defer ticker.mu.Unlock()

        ticker.d = d
        ticker.deadline = ticker.clock.Now().Add(d)
}

// Stop stops the ticker. The ticker's channel stops receiving
// messages after it has been stopped. The channel is not closed.
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
