package simtest

import (
	"math/rand/v2"
	"sort"
	"testing"
	"time"
)

func TestEventsListPopsEventsInAscendingOrder(t *testing.T) {
        var e *event

        now := time.Now()
        sleepEvent := newSleepEvent(now, 1 * time.Minute)
        timerEvent := newTimerEvent(now, 2 * time.Minute)
        tickerEvent := newTickerEvent(now, 3 * time.Minute)

        evts := newEvents()
        evts.Push(tickerEvent)
        evts.Push(sleepEvent)
        evts.Push(timerEvent)

        if e = evts.Pop(); e.deadline != sleepEvent.deadline {
                t.Errorf("sleep event not popped first, got %#v", e)
        }

        if e = evts.Pop(); e.deadline != timerEvent.deadline {
                t.Errorf("timer event not popped second, got %#v", e)
        }

        if e = evts.Pop(); e.deadline != tickerEvent.deadline {
                t.Errorf("ticker event not popped third, got %#v", e)
        }

        t.Logf("events len %d", evts.Len())
}

func FuzzEventsListPopsEventsInAscendingOrder(f *testing.F) {
        f.Add(uint(2))
        f.Add(uint(100))
        f.Add(uint(1))
        f.Add(uint(0))
        f.Add(uint(1000))

        f.Fuzz(func(t *testing.T, numEvents uint) {
                if numEvents == 0 {
                        t.SkipNow()
                }

                now := time.Now()
                actual := make(events, 0)
                events := newEvents()

                for range numEvents {
                        ch := make(chan time.Time)
                        typ := eventType(rand.UintN(uint(ticker)))
                        dur := time.Duration(rand.Uint64N(1000)) * time.Millisecond
                        deadline := now.Add(dur)
                        
                        event := &event{ch, typ, dur, deadline}
                        events.Push(event)
                        actual = append(actual, event)
                }

                sort.Sort(actual)

                prevEvent := events[0]
                for i, actualEvent := range actual {
                        event := events.Pop()
                        
                        if !actualEvent.deadline.Equal(event.deadline) {
                                t.Errorf("%d: popped deadline %s but expected %s", i, event.deadline, actualEvent.deadline)
                        }

                        if !prevEvent.deadline.Before(event.deadline) && !prevEvent.deadline.Equal(event.deadline) {
                                t.Errorf("%d: popped deadline %s is not after previous deadline %s", i, event.deadline, prevEvent.deadline)
                        }
                }
        })
}
