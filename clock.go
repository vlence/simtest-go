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
        NewTimer(d time.Duration) (Timer, <-chan time.Time)

        // NewTicker returns a ticker that fires every d intervals.
        NewTicker(d time.Duration) Ticker

        // Sleep blocks this goroutine for d amount of time.
        Sleep(d time.Duration)
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

type sleepDeadline struct {
        ch chan time.Time
        deadline time.Time
}

type sleepEvents struct {
        add chan *sleepDeadline
        tick chan time.Time
        stop chan bool
}

// SimClock represents a simulated clock. The clock moves forward
// in time only when Tick is called. Care must be taken that Sleep,
// timers and tickers are not used in the same goroutine as the one
// where Tick is called; you will deadlock.
type SimClock struct {
        mu          *sync.RWMutex
        now         time.Time
        stopped     bool
        sleepEvents *sleepEvents
        timerEvents *timerEvents
        sleepTickSize time.Duration
}

// NewSimClock returns a SimClock whose current time is now.
func NewSimClock(now time.Time) *SimClock {
        clock := new(SimClock)
        clock.mu = new(sync.RWMutex)
        clock.now = now
        clock.stopped = false
        clock.sleepTickSize = 1 * time.Microsecond

        clock.timerEvents = &timerEvents{
                add:    make(chan *SimTimer),
                remove: make(chan *SimTimer),
                stop:   make(chan bool),
                tick:   make(chan time.Time),
        }

        clock.sleepEvents = &sleepEvents{
                add: make(chan *sleepDeadline),
                tick: make(chan time.Time),
                stop: make(chan bool),
        }

        go clock.listenForTimerEvents()
        go clock.listenForSleepEvents()

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

// NewTimer returns a *SimTimer. If the clock has been stopped it returns
// nil for the timer and channel.
func (clock *SimClock) NewTimer(d time.Duration) (Timer, <-chan time.Time) {
        if clock.stopped {
                return nil, nil
        }

        timer := newSimTimer(clock.Now().Add(d), clock.timerEvents)
        clock.timerEvents.add <- timer

        return timer, timer.ch
}

// NewTicker returns a *SimTicker. If the clock has been stopped it
// returns nil.
func (clock *SimClock) NewTicker(d time.Duration) Ticker {
        if clock.stopped {
                return nil
        }

        ch := make(chan time.Time)
        ticker := newSimTicker(d, (<-chan time.Time)(ch), clock)

        return ticker
}

// SetSleepTickSize sets the speed at which this clock will tick
// while sleeping. By default this tick size is 1 microsecond.
func (clock *SimClock) SetSleepTickSize(d time.Duration) {
        clock.mu.Lock()
        defer clock.mu.Unlock()

        clock.sleepTickSize = d
}

// Sleep blocks this goroutine for d amount of time. If you
// sleep in the same goroutine that ticks this clock then
// you will be in a deadlock.
func (clock *SimClock) Sleep(d time.Duration) {
        if clock.stopped {
                return
        }

        // time.Sleep is designed to block the current goroutine
        // until the given duration d has passed. We don't want
        // to do that here because that would mean timers and
        // tickers wouldn't be fired. So we just get the current
        // time and tick until the clock has moved forward by d
        // amount of time.
        ch := make(chan time.Time, 1)
        now := clock.Now()
        deadline := now.Add(d)

        clock.sleepEvents.add <- &sleepDeadline{ch, deadline}
        <-ch
}

// Tick moves the clock's time forward by tickSize and returns
// the current time. If the clock has been stopped it does
// nothing and returns the time when the clock was stopped.
func (clock *SimClock) Tick(tickSize time.Duration) time.Time {
        if clock.stopped {
                return clock.now
        }

        clock.mu.Lock()
        defer clock.mu.Unlock()

        now := clock.now.Add(tickSize)
        clock.now = now

        clock.timerEvents.tick <- now
        clock.sleepEvents.tick <- now

        return now
}

// Stop stops this clock and its associated timers and tickers.
// Any pending timers and tickers are never fired after Stop is
// called.
func (clock *SimClock) Stop() time.Time {
        clock.mu.Lock()
        defer clock.mu.Unlock()

        clock.timerEvents.stop <- true
        close(clock.timerEvents.add)
        close(clock.timerEvents.remove)
        close(clock.timerEvents.stop)
        close(clock.timerEvents.tick)

        clock.stopped = true

        return clock.now
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

func (clock *SimClock) listenForSleepEvents() {
        var yes bool
        var deadlines map[time.Time]chan time.Time

        deadlines = make(map[time.Time]chan time.Time)
        
        for {
                select {
                case sd := <-clock.sleepEvents.add:
                       deadlines[sd.deadline] = sd.ch
                case now := <-clock.sleepEvents.tick:
                       for deadline, ch := range deadlines {
                               if deadline.Equal(now) || now.After(deadline) {
                                       ch <- now
                                       delete(deadlines, deadline)
                               }
                       }
               case yes = <-clock.sleepEvents.stop:
                       if yes {
                               return
                       }
               }
        }
}
