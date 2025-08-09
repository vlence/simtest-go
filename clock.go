package simtest

import (
        "sort"
        "sync"
        "time"

        "github.com/vlence/gossert"
)

// A Clock returns the current time and can create timers and tickers. An
// application should use the same clock everywhere.
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

// SimClock is a simulated clock. The clock moves forward
// in time only when Tick is called. Every simulated clock
// runs a goroutine in the background that manages callbacks
// like sleeps, timers, and tickers. Some of its methods are
// not safe for use within the same goroutine. Read the
// documentation of Tick, NewTimer, NewTicker, and Sleep.
type SimClock struct {
        // Use this mutex to sync reads and writes to now.
        nowMu *sync.RWMutex

        // Use this mutex to sync reads and writes to stopped.
        stopMu *sync.RWMutex

        // The current time of the clock. Call Now to get this value
        // in a thread safe manner.
        now time.Time

        // This is true if this clock has been stopped. Call isStopped
        // to get this value in a thread safe manner.
        stopped bool

        // Send the current time to the event manager goroutine on every
        // tick.
        tickCh chan time.Time

        // Send true to this channel to stop the event manager goroutine.
        stopCh chan bool

        // Send event updates to this channel to update events.
        eventUpdateCh chan *eventUpdate

        // Send new events to this channel to register new events.
        registerEventCh chan *event
}

// NewSimClock returns a SimClock whose current time is now.
func NewSimClock(now time.Time) *SimClock {
        clock := new(SimClock)
        clock.nowMu = new(sync.RWMutex)
        clock.stopMu = new(sync.RWMutex)
        clock.now = now
        clock.stopCh = make(chan bool)
        clock.tickCh = make(chan time.Time)
        clock.eventUpdateCh = make(chan *eventUpdate)
        clock.registerEventCh = make(chan *event)

        go clock.eventManager()

        return clock
}

// Now returns the current time. To move the time forward call Tick.
func (clock *SimClock) Now() time.Time {
        gossert.Ok(nil != clock, "simclock: clock is nil")

        clock.nowMu.RLock()
        defer clock.nowMu.RUnlock()

        return clock.now
}

// NewTimer creates and returns a simulated timer. The simulated timer fires
// after d amount of time has passed i.e. it fires if Tick has been called
// enough times to simulate d amount of time passing. Care must be taken to
// not create a simulated timer and wait for it to expire in the same
// goroutine that calls Tick because it can deadlock. NewTimer panics
// if the clock has been stopped.
func (clock *SimClock) NewTimer(d time.Duration) (Timer, <-chan time.Time) {
        gossert.Ok(nil != clock, "simclock: clock is nil")
        gossert.Ok(!clock.isStopped(), "simclock: creating new timer using stopped clock")

        clock.nowMu.RLock()
        defer clock.nowMu.RUnlock()

        timer := &simTimer{
                event{
                        d:       d,
                        ch:      make(chan time.Time),
                        when:    clock.now.Add(d),
                        repeat:  false,
                        stopped: false,
                },
                clock,
        }

        clock.registerEventCh <- &timer.event

        return timer, timer.ch
}

// NewTicker returns a *SimTicker. NewTicker panics if the clock has been stopped.
func (clock *SimClock) NewTicker(d time.Duration) Ticker {
        gossert.Ok(nil != clock, "simclock: clock is nil")
        gossert.Ok(!clock.isStopped(), "simclock: creating new ticker using stopped clock")

        ch := make(chan time.Time)
        ticker := newSimTicker(d, (<-chan time.Time)(ch), clock)

        return ticker
}

// Sleep simulates blocking this goroutine for d amount of time. This
// function will return once Tick has been called enough times to
// simulate d amount of time passing. Care must be taken to ensure Sleep
// and Tick are not called from the same goroutine. If they're called
// from the same goroutine the program will deadlock. Sleep will panic
// if the clock has been stopped.
func (clock *SimClock) Sleep(d time.Duration) {
        gossert.Ok(nil != clock, "simclock: clock is nil")
        gossert.Ok(!clock.isStopped(), "simclock: sleeping a stopped clock")

        ev := &event{
                d:       d,
                ch:      make(chan time.Time),
                when:    clock.Now().Add(d),
                repeat:  false,
                stopped: false,
        }

        clock.registerEventCh <- ev

        <-ev.ch
}

// Tick moves the clock's time forward by tickSize and returns
// the current time. All registered events that need to be fired
// next will be fired. This method will block the goroutine and
// can lead to a deadlock in certain situations. For example if
// you call Sleep and Tick in the same goroutine, that goroutine
// will deadlock. Tick will panic if the clock has been stopped.
func (clock *SimClock) Tick(tickSize time.Duration) time.Time {
        gossert.Ok(nil != clock, "simclock: clock is nil")
        gossert.Ok(!clock.isStopped(), "simclock: ticking a stopped clock")

        clock.nowMu.Lock()
        defer clock.nowMu.Unlock()

        now := clock.now.Add(tickSize)
        clock.now = now

        clock.tickCh <- now

        return now
}

// updateEvent sends an event update request to the background goroutine that manages
// the events. updateEvent will panic if the clock has been stopped.
func (clock *SimClock) updateEvent(d time.Duration, event *event) {
        gossert.Ok(nil != clock, "simclock: clock is nil")

        gossert.Ok(!clock.isStopped(), "simclock: updating event of stopped clock")
        gossert.Ok(event != nil, "simclock: trying to update nil event")

        clock.eventUpdateCh <- &eventUpdate{d, event}
}

// eventManager manages events. It accepts new events to be registered,
// updates them and fires them once they have expired. eventManager
// must be run as a separate goroutine.
func (clock *SimClock) eventManager() {
        gossert.Ok(nil != clock, "simclock: clock is nil")

        events := make(callbacks, 0)

        for {
                select {
                case newEvent := <-clock.registerEventCh:
                        events.Register(newEvent)

                case update := <-clock.eventUpdateCh:
                        d, event := update.d, update.event
                        gossert.Ok(event != nil, "simclock: trying to update nil event")

                        if d < 0 {
                                event.stopped = true
                                break
                        }

                        event.d = d
                        event.when = clock.Now().Add(d)
                        sort.Sort(events)

                case now := <-clock.tickCh:
                        // execute callbacks that have expired.
                        for event := events.peek(); event != nil && !now.Before(event.when); event = events.peek() {
                                gossert.Ok(event == events.next(), "simclock: peeked callback is not same as popped callback")

                                if event.stopped {
                                        continue
                                }

                                event.ch <- now

                                if !event.repeat {
                                        event.stopped = true
                                        continue
                                }

                                event.when = event.when.Add(event.d)
                                events.Register(event)
                        }

                case <-clock.stopCh:
                        for event := events.next(); event != nil; event = events.next() {
                                event.stopped = true
                                close(event.ch)
                        }
                        return
                }
        }
}

// isStopped returns true if this clock has been stopped.
func (clock *SimClock) isStopped() bool {
        gossert.Ok(nil != clock, "simclock: clock is nil")

        clock.stopMu.RLock()
        defer clock.stopMu.RUnlock()

        return clock.stopped
}

// stop stops this clock. Channels of all pending events are closed.
func (clock *SimClock) stop() bool {
        gossert.Ok(nil != clock, "simclock: clock is nil")

        clock.stopMu.Lock()
        defer clock.stopMu.Unlock()

        if clock.stopped {
                return false
        }

        clock.stopCh <- true

        close(clock.tickCh)
        close(clock.stopCh)
        close(clock.registerEventCh)
        close(clock.eventUpdateCh)

        clock.stopped = true

        return true
}

// An event is some future event. Events have a channel to which the
// current time is sent when the event happens.
type event struct {
        // The amount of time until the event happens from the moment it was created.
        d time.Duration

        // The current time will be sent to this channel when the event occurs.
        ch chan time.Time

        // The time when the event will occur.
        when time.Time

        // Whether the event repeats every d duration.
        repeat bool

        // Whether this event has been stopped/canceled.
        stopped bool
}

// An eventUpdate is an update for an event. Simulated clocks take event
// updates and update the state of registered events.
type eventUpdate struct {
        // Set to -1 to stop event. 0 does nothing. Any positive number
        // will reset the event duration.
        d time.Duration

        // The event to apply the update to.
        event *event
}

// List of registered events that need to be fired when they occur.
// An event is registered if it is in the list.
type callbacks []*event

// Len returns the number of registered events.
func (cbs callbacks) Len() int {
        return len(cbs)
}

// Less returns true if the event at index i needs to be fired before
// the event at index j.
func (cbs callbacks) Less(i, j int) bool {
        return cbs[i].when.Before(cbs[j].when)
}

// Swap swaps the events at indexes i and j.
func (cbs callbacks) Swap(i, j int) {
        cbs[i], cbs[j] = cbs[j], cbs[i]
}

// Register registers the event ev. ev is guaranteed to
// fire in the next tick if its deadline comes before
// the other registered events.
func (cbs *callbacks) Register(ev *event) {
        gossert.Ok(nil != cbs, "simclock: cannot register new event on nil events list")

        *cbs = append(*cbs, ev)
        sort.Sort(cbs)
}

// next returns the next event that should be fired and removes it
// from the list. It returns nil if there are no events i.e. if
// e.Len() == 0.
func (cbs *callbacks) next() *event {
        gossert.Ok(nil != cbs, "simclock: cannot get next event from nil events list")

        if len(*cbs) == 0 {
                return nil
        }

        ev := (*cbs)[0]

        *cbs = (*cbs)[1:]

        return ev
}

// peek returns the next event that needs to be fired without
// removing it from the list. It returns nil if there are no
// events i.e. if e.Len() == 0.
func (cbs *callbacks) peek() *event {
        gossert.Ok(nil != cbs, "simclock: cannot peek next event from nil events list")
        if len(*cbs) == 0 {
                return nil
        }

        return (*cbs)[0]
}
