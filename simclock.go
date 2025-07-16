package simtest

import (
	"fmt"
	"sync"
	"time"

	"github.com/vlence/gossert"
)

// SimClock represents a simulated clock. The clock moves forward
// in time only when Tick is called.
type SimClock struct{
        mu *sync.RWMutex
        now time.Time
        timers map[*SimTimer]chan time.Time
        tickers map[*SimTicker]chan time.Time
}

// NewSimClock returns a SimClock whose current time is now.
func NewSimClock(now time.Time) *SimClock {
        clock := new(SimClock)
        clock.mu = new(sync.RWMutex)
        clock.now = now
        clock.timers = make(map[*SimTimer]chan time.Time)

        return clock
}

// Now returns the current time. To move
// the time forward call Tick. Repeated calls to Now will return
// the same time until Tick is called.
func (clock *SimClock) Now() time.Time {
        clock.mu.RLock()
        defer clock.mu.RUnlock()

        return clock.now
}

func (clock *SimClock) NewTimer(d time.Duration) Timer {
        ch := make(chan time.Time)
        timer := newSimTimer(d, (<-chan time.Time)(ch), clock)
        clock.timers[timer] = ch

        return timer
}

func (clock *SimClock) NewTicker(d time.Duration) Ticker {
        ch := make(chan time.Time)
        ticker := newSimTicker(d, (<-chan time.Time)(ch), clock)
        clock.tickers[ticker] = ch

        return ticker
}

// Tick moves the clock's time forward by tickSize and returns
// the current time.
func (clock *SimClock) Tick(tickSize time.Duration) time.Time {
        clock.mu.Lock()
        defer clock.mu.Unlock()

        now := clock.now.Add(tickSize)
        clock.now = now

        clock.fireTimers(now)
        clock.fireTicks(now)

        return now
}

// fireTimers fires the timers that are due. Fired timers
// are removed from the timers list.
func (clock *SimClock) fireTimers(now time.Time) {
        for timer, ch := range clock.timers {
                timer.mu.Lock()

                passedDeadline := timer.deadline.Before(now)

                if passedDeadline {
                        gossert.Ok(timer.Stop(), fmt.Sprintf("simclock: stopped timer with deadline %s not removed from timers list", timer.deadline.String()))
                        ch <- now
                        delete(clock.timers, timer)
                }

                timer.mu.Unlock()
        }
}

// fireTicks fires ticks that are due and updates their next deadline.
func (clock *SimClock) fireTicks(now time.Time) {
        for ticker, ch := range clock.tickers {
                ticker.mu.Lock()

                gossert.Ok(!ticker.stopped, "simclock: stopped ticker not removed from tickers list")

                passedDeadline := ticker.deadline.Before(now)

                if passedDeadline {
                        ch <- now
                        ticker.deadline = ticker.deadline.Add(ticker.d)
                }

                ticker.mu.Unlock()
        }
}

// removeTicker removes the given ticker from clock.tickers.
// This function is meant to be called by SimTicker when it is stopped.
// Do NOT call this function inside Tick; will cause deadlock.
func (clock *SimClock) removeTicker(ticker *SimTicker) {
        clock.mu.Lock()
        defer clock.mu.Unlock()

        delete(clock.tickers, ticker)
}

// removeTimer removes the given timer from clock.timers.
// This function is meant to be called by SimTimer when it is stopped.
// Do NOT call this function inside Tick; will cause deadlock.
func (clock *SimClock) removeTimer(timer *SimTimer) {
        clock.mu.Lock()
        defer clock.mu.Unlock()

        _, ok := clock.timers[timer]

        if ok {
                delete(clock.timers, timer)
        }
}
