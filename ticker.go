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
        simClockEvent
        clock *SimClock
}

// Reset resets this simulated ticker to fire after d amount of time
// has passed. If d is not greater than 0 Reset panics. Reset does
// nothing if the ticker has stopped.
func (ticker *simTicker) Reset(d time.Duration) {
        gossert.Ok(nil != ticker, "ticker: cannot reset nil ticker")
        gossert.Ok(d > 0, "ticker: duration is not greater than 0")

        if ticker.stopped {
                return
        }

        ticker.clock.updateEvent(d, &ticker.simClockEvent)
}

// Stop stops this simulated ticker. The underlying channel is not closed.
func (ticker *simTicker) Stop() {
        gossert.Ok(nil != ticker, "ticker: cannot reset nil ticker")

        if ticker.stopped {
                return
        }

        ticker.clock.updateEvent(-1, &ticker.simClockEvent)
}
