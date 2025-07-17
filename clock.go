package simtest

import (
	"sync"
	"time"

	"github.com/vlence/gossert"
)

// A Clock reports the current time whenever Now is called.
// An application should use the same clock everywhere to
// tell the time.
type Clock interface {
        // Now returns the current time.
        Now() time.Time

        // NewTimer returns a timer that fires after d time has passed.
        NewTimer(d time.Duration) Timer

        // NewTicker returns a ticker that fires every d intervals.
        NewTicker(d time.Duration) Ticker
}

// timerEvents represents all the timer events that can been
// happen. See Clock.listenForTimerEvents to see how the channels
// are used.
type timerEvents struct {
        // Newly created timers should be sent to this channel.
        add chan *SimTimer

        // Timers that have been fired or stopped should be sent to this channel.
        remove chan *SimTimer

        // Send true to this channel to stop listening for timer events.
        stop chan bool

        // Send clock ticks to this channel.
        tick chan time.Time
}

// SimClock represents a simulated clock. The clock moves forward
// in time only when Tick is called.
type SimClock struct {
        mu          *sync.RWMutex
        now         time.Time
        timerEvents *timerEvents
}

// NewSimClock returns a SimClock whose current time is now.
func NewSimClock(now time.Time) *SimClock {
        clock := new(SimClock)
        clock.mu = new(sync.RWMutex)
        clock.now = now
        clock.timerEvents = &timerEvents{
                add:    make(chan *SimTimer),
                remove: make(chan *SimTimer),
                stop:   make(chan bool),
                tick:   make(chan time.Time),
        }

        go clock.listenForTimerEvents()

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

// NewTimer returns a *SimTimer.
func (clock *SimClock) NewTimer(d time.Duration) Timer {
        timer := newSimTimer(clock.Now().Add(d), clock.timerEvents)
        clock.timerEvents.add <- timer

        return timer
}

// NewTicker returns a *SimTicker.
func (clock *SimClock) NewTicker(d time.Duration) Ticker {
        ch := make(chan time.Time)
        ticker := newSimTicker(d, (<-chan time.Time)(ch), clock)

        return ticker
}

// Tick moves the clock's time forward by tickSize and returns
// the current time.
func (clock *SimClock) Tick(tickSize time.Duration) time.Time {
        clock.mu.Lock()
        defer clock.mu.Unlock()

        now := clock.now.Add(tickSize)
        clock.now = now

        clock.timerEvents.tick <- now

        return now
}

// Stop stops this clock and its associated timers and tickers.
// Any pending timers and tickers are never fired after Stop is
// called.
func (clock *SimClock) Stop() {
        clock.timerEvents.stop <- true
        close(clock.timerEvents.add)
        close(clock.timerEvents.remove)
        close(clock.timerEvents.stop)
        close(clock.timerEvents.tick)
}

// listenForTimerEvents starts a goroutine that adds timers to
// a watchlist, fires them when they've reached their deadline
// and removes them from the watchlist.
//
// When a new timer is created it should be sent to addTimerChan.
// If not sent then the timer will never be fired even if it's
// deadline has passed.
//
// When the clock ticks it should send the correct time to
// timerTickChan. All timers are checked and the ones whose
// deadline has been reached are fired. Timers are automatically
// removed from the watchlist after they're fired.
//
// If a timer is stopped by calling Stop it should be sent to
// removeTimerChan to remove it from the watchlist.
func (clock *SimClock) listenForTimerEvents() {
        var yes bool
        var now time.Time
        var timer *SimTimer
        var timers map[*SimTimer]bool

        timers = make(map[*SimTimer]bool)

        for {
                select {
                case timer = <-clock.timerEvents.add:
                        timers[timer] = true

                case timer = <-clock.timerEvents.remove:
                        delete(timers, timer)

                case now = <-clock.timerEvents.tick:
                        for timer = range timers {
                                gossert.Ok(!timer.stopped, "simclock: stopped timer not removed from watchlist")

                                fired := timer.fire(now)

                                if fired {
                                        delete(timers, timer)
                                }
                        }
                case yes = <-clock.timerEvents.stop:
                        if yes {
                                return
                        }
                }
        }
}
