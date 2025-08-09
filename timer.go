package simtest

import (
	"time"

	"github.com/vlence/gossert"
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
        simClockEvent
        clock *SimClock
}

// Reset resets the timer to fire after d amount of
// time has passed since calling Reset and returns
// true. It returns false if the timer has already
// been fired or has expired.
func (timer *simTimer) Reset(d time.Duration) bool {
        gossert.Ok(nil != timer, "simtimer: resetting nil timer")

        if timer.stopped {
                return false
        }

        if d < 0 {
                // fire the timer in the next tick
                d = 0
        }

        timer.clock.updateEvent(d, &timer.simClockEvent)

        return true
}

// Stop cancels this timer and returns true. If the timer
// has already fired or has expired then it does nothing
// and returns false.
func (timer *simTimer) Stop() bool {
        gossert.Ok(nil != timer, "simtimer: stopping nil timer")

        if timer.stopped {
                return false
        }

        timer.clock.updateEvent(-1, &timer.simClockEvent)

        return true
}
