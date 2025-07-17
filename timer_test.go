package simtest

import (
	"testing"
	"time"
)

func TestNewTimerHasExpectedDeadline(t *testing.T) {
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

func TestTimerIsFiredAfterClockTicks(t *testing.T) {
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
                        t.Logf("timer fired at %s", now)
                        if now.Before(expectedDeadline) {
                                t.Errorf("timer fired too early, expected %s", expectedDeadline)
                        }
                        return
                default:
                        clock.Tick(tickSize)
                }
        }

        t.Errorf("timer wasn't fired")
}
