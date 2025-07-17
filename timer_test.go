package simtest

import (
	"testing"
	"time"
)

func TestTimerHasExpectedDeadline(t *testing.T) {
        epoch := time.Now()
        clock := NewSimClock(epoch)
        defer clock.Stop()

        d := 1 * time.Second
        expectedDeadline := epoch.Add(d)

        timer, _ := clock.NewTimer(d).(*SimTimer)

        if !expectedDeadline.Equal(timer.deadline) {
                t.Errorf("timer's deadline %s does not match expected deadline %s", timer.deadline, expectedDeadline)
        }
}

func TestTimerIsFired(t *testing.T) {
        epoch := time.Now()
        clock := NewSimClock(epoch)
        defer clock.Stop()

        dur := 1 * time.Second
        timer, _ := clock.NewTimer(dur).(*SimTimer)
        expectedDeadline := epoch.Add(dur)

        tickSize := 100 * time.Millisecond
        for range 100 {
                select {
                case now := <-timer.C:
                        if now.Before(expectedDeadline) {
                                t.Errorf("timer fired too early; fired at %s but should have been fired after %s", now, expectedDeadline)
                        }
                        return
                default:
                        clock.Tick(tickSize)
                }
        }

        t.Errorf("timer wasn't fired")
}

func TestTimerFiredOnlyOnce(t *testing.T) {
        epoch := time.Now()
        clock := NewSimClock(epoch)
        defer clock.Stop()

        dur := 1 * time.Second
        timer, _ := clock.NewTimer(dur).(*SimTimer)

        fired := false
        tickSize := 100 * time.Millisecond
        for range 100 {
                select {
                case <-timer.C:
                        if fired {
                                t.Errorf("timer fired twice")
                        }

                        fired = true
                default:
                        clock.Tick(tickSize)
                }
        }
}
