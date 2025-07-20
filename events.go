package simtest

import (
	"sort"
	"time"
)

type eventType uint

const (
        sleep eventType = iota
        timer
        ticker
)

// An event is a sleep event, a timer event or a ticker event.
// ticker events are special because they are added back to the heap
// with a new deadline until the ticker is stopped.
type event struct {
        // The channel to send the current time to when this event is fired.
        ch chan time.Time

        // The type of this event.
        typ eventType

        // The duration after which this event is fired.
        dur time.Duration

        // The time when this event should be fired.
        deadline time.Time
}

// newEvent returns a new event of type typ that should be fired after
// d amount of time has passed since now.
func newEvent(now time.Time, d time.Duration, typ eventType) *event {
        return &event{
                ch: make(chan time.Time),
                dur: d,
                typ: typ,
                deadline: now.Add(d),
        }
}

// newSleepEvent returns a new sleep event that fires after d amount
// of time has passed since now.
func newSleepEvent(now time.Time, d time.Duration) *event {
        return newEvent(now, d, sleep)
}

// newTimerEvent returns a new timer event that fires after d amount
// of time has passed since now.
func newTimerEvent(now time.Time, d time.Duration) *event {
        return newEvent(now, d, timer)
}

// newTickerEvent returns a new ticker event that fires after d amount
// of time has passed since now.
func newTickerEvent(now time.Time, d time.Duration) *event {
        return newEvent(now, d, ticker)
}

// A sorted list of events that need to be fired. The event with
// the earliest deadline is returned by Pop. Call Push to add new
// events to the list. List is sorted at every Push.
type events []*event

// newEvents returns a new events list.
func newEvents() events {
        return make(events, 0)
}

// Len returns the number of events that are pending to be fired.
func (e events) Len() int {
        return len(e)
}

// Less returns true if the event at index i has an earlier deadline
// than the event at index j.
func (e events) Less(i, j int) bool {
        return e[i].deadline.Before(e[j].deadline)
}

// Swap swaps the events at indexes i and j.
func (e events) Swap(i, j int) {
        e[i], e[j] = e[j], e[i]
}

// Push adds the new event ev to the list of events and sorts it.
// If ev has an earlier deadline than all the current events then
// ev will be fired next.
func (e *events) Push(ev *event) {
        *e = append(*e, ev)
        sort.Sort(*e)
}

// Pop returns the next event that should be fired and removes it
// from the list. It returns nil if there are no events i.e. if
// e.Len() == 0.
func (e *events) Pop() *event {
        if len(*e) == 0 {
                return nil
        }

        ev := (*e)[0]

        *e = (*e)[1:]

        return ev
}

// Peek returns the next event that needs to be fired without
// removing it from the list. It returns nil if there are no
// events i.e. if e.Len() == 0.
func (e *events) Peek() *event {
        if len(*e) == 0 {
                return nil
        }

        return (*e)[0]
}
