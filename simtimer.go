package simtest

import (
	"sync"
	"time"

	"github.com/vlence/gossert"
)

type SimTimer struct {
        C <-chan time.Time
        mu *sync.Mutex
        clock *SimClock
        stopped bool
        deadline time.Time
}

func newSimTimer(d time.Duration, c <-chan time.Time, clock *SimClock) *SimTimer {
        timer := new(SimTimer)
        timer.C = c
        timer.mu = new(sync.Mutex)
        timer.clock = clock
        timer.stopped = false
        timer.deadline = clock.Now().Add(d)

        return timer
}

func (timer *SimTimer) Reset(d time.Duration) bool {
        timer.mu.Lock()
        defer timer.mu.Unlock()

        if timer.stopped {
                return false
        }

        timer.deadline = timer.deadline.Add(d)

        return true
}

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
