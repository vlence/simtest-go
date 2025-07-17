package simtest

import (
	"testing"
	"time"
)

func TestSimClockDoesNotProgressIfNotTicked(t *testing.T) {
        epoch := time.Now()
        clock := NewSimClock(epoch)
        defer clock.Stop()
        
        // Loop for a while to be sure.
        // I mean we can just read the code and know that time shouldn't
        // pass but can we really be that sure? Stranger things have
        // happened.
        for range 1_000_000 {
                currentTime := time.Now()
                timeHasPassed := currentTime.After(epoch)
                currentSimTime := clock.Now()
                clockTimeHasChanged := !currentSimTime.Equal(epoch)

                if timeHasPassed && clockTimeHasChanged {
                        t.Errorf("simclock time changed from %s to %s", epoch, currentSimTime)
                }
        }
}

func TestSimClockProgressesAfterTick(t *testing.T) {
        epoch := time.Now()
        clock := NewSimClock(epoch)
        defer clock.Stop()

        tickSize := 1 * time.Millisecond
        currentTime := clock.Tick(tickSize)
        expectedTime := epoch.Add(tickSize)

        if !currentTime.Equal(expectedTime) {
                t.Errorf("expected time %s after tick (size=%s) but got %s", expectedTime, tickSize, currentTime)
        }
}

func TestSimClockTickReturnsCurrentTime(t *testing.T) {
        epoch := time.Now()
        clock := NewSimClock(epoch)
        defer clock.Stop()

        tickSize := 1 * time.Millisecond
        timeReturnedByTick := clock.Tick(tickSize)
        currentTime := clock.Now()

        if !timeReturnedByTick.Equal(currentTime) {
                t.Errorf("time returned by Tick %s not equal to time returned by Now %s", timeReturnedByTick, currentTime)
        }
}
