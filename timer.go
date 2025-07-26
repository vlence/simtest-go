package simtest

import (
	"time"
)

// A Timer represents an event in the future. This interface is kept deliberately
// similar to the design of time.Timer.
type Timer interface {
        Reset(d time.Duration) bool

        Stop() bool
}

// simTimer is a simulated timer. A simulated timer is much like a real timer,
// in that it represents some event in the future, can be stopped and can be
// reset. Simulated timers expose a channel that returns the time at which they
// were fired. The channel is unbuffered like an actual timer. This means the
// simulated clock cannot progress forward until a receiver receives from the
// timer's channel.
type simTimer struct {
        event
        clock *SimClock
}

// Reset resets the timer to fire after d amount of
// time has passed since calling Reset and returns
// true. It returns false if the timer has already
// been fired or has expired.
func (cb *simTimer) Reset(d time.Duration) bool {
        if cb.stopped {
                return false
        }

        if d < 0 {
                // fire the timer in the next tick
                d = 0
        }

        cb.clock.updateEvent(d, &cb.event)

        return true
}

// Stop cancels this timer and returns true. If the timer
// has already fired or has expired then it does nothing
// and returns false.
func (cb *simTimer) Stop() bool {
        if cb.stopped {
                return false
        }

        cb.clock.updateEvent(-1, &cb.event)

        return true
}
