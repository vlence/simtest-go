package simtest

import (
	"sync"
	"time"

	"github.com/vlence/gossert"
)

// A Timer represents an action that needs to be completed in the future.
// The interface is deliberately kept similar to that of *time.Timer.
type Timer interface {
        Reset(d time.Duration) bool

        Stop() bool
}

// SimTimer represents a simulater timer. Timers are used to represent
// a moment in the future. We can do something when a timer fires.
// SimTimer uses SimClock internally so the timer never fires unless
// the SimClock is Tick'ed and the timer's deadline is passed.
type SimTimer struct {
        C <-chan time.Time
        mu *sync.Mutex
        clock *SimClock
        stopped bool
        deadline time.Time
}

// newSimTimer returns a new SimTimer.
func newSimTimer(d time.Duration, c <-chan time.Time, clock *SimClock) *SimTimer {
        timer := new(SimTimer)
        timer.C = c
        timer.mu = new(sync.Mutex)
        timer.clock = clock
        timer.stopped = false
        timer.deadline = clock.Now().Add(d)

        return timer
}

// Reset updates the timer to fire after d time has passed
// since Reset is called. If the timer has already fired
// or was previously stopped then Reset does nothing.
func (timer *SimTimer) Reset(d time.Duration) bool {
        timer.mu.Lock()
        defer timer.mu.Unlock()

        if timer.stopped {
                return false
        }

        timer.deadline = timer.deadline.Add(d)

        return true
}

// Stop stops the timer. If the timer hasn't been fired when
// Stop is called then it'll never be fired. The timer's
// channel is not closed after the timer has been stopped.
func (timer *SimTimer) Stop() bool {
        timer.mu.Lock()
        defer timer.mu.Unlock()

        if timer.stopped {
                return false
        }

        timer.stopped = true
        timer.clock.removeTimer(timer)
        _, ok := timer.clock.timers[timer]
        gossert.Ok(!ok, "simtimer: timer not removed from clock's timers list")

        return true
}
