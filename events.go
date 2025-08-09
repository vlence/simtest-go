package simtest

import (
	"sort"
	"time"

	"github.com/vlence/gossert"
)

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
