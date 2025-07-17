package simtest

import (
        "sync"
        "time"
)

// send timers to this channel to watch them
var addTimerChan = make(chan *SimTimer)

// send ticks to this channel so that timers can be fired
var timerTickChan = make(chan time.Time)

// send true to this channel to stop listening for timer events
var stopTimersChan = make(chan bool)

// send timers to this channel to remove them from the watch list
var removeTimerChan = make(chan *SimTimer)

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
        C        <-chan time.Time
        ch       chan time.Time
        mu       *sync.Mutex
        stopped  bool
        deadline time.Time
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
func listenForTimerEvents() {
        var yes bool
        var now time.Time
        var timer *SimTimer
        var timers map[*SimTimer]bool

        timers = make(map[*SimTimer]bool)

        for {
                select {
                case timer = <-addTimerChan:
                        timers[timer] = true

                case timer = <-removeTimerChan:
                        delete(timers, timer)

                case now = <-timerTickChan:
                        for timer = range timers {
                                fired := timer.fire(now)

                                if fired {
                                        delete(timers, timer)
                                }
                        }
                case yes = <-stopTimersChan:
                        if yes {
                                return
                        }
                }
        }
}

func stopListeningForTimerEvents() {
        stopTimersChan <- true
}

// newSimTimer returns a new SimTimer.
func newSimTimer(deadline time.Time) *SimTimer {
        ch := make(chan time.Time)

        timer := new(SimTimer)
        timer.C = ch
        timer.ch = ch
        timer.mu = new(sync.Mutex)
        timer.stopped = false
        timer.deadline = deadline

        addTimerChan <- timer

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

        removeTimerChan <- timer
        timer.stopped = true

        return true
}

// fire fires the timer if its deadline has been reached.
// The timer's deadline is reached if the current time now
// is after or equal to the deadline.
func (timer *SimTimer) fire(now time.Time) bool {
        timer.mu.Lock()
        defer timer.mu.Unlock()

        if timer.stopped {
                return false
        }

        passedDeadline := now.After(timer.deadline) || now.Equal(timer.deadline)

        if !passedDeadline {
                return false
        }

        timer.stopped = true
        timer.ch <- now

        return true
}
