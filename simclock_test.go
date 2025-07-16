package simtest

import (
	"testing"
	"time"
)

func TestSimClockDoesNotProgressUntilTicked(t *testing.T) {
        now := time.Now()
        clock := NewSimClock(now)
        
        for range 1000 {
                if now != clock.Now() {
                        t.Errorf("simclock time has changed without ticking")
                }
        }
        

        tickSize := 1 * time.Millisecond
        expectedTime := now.Add(tickSize)

        clock.Tick(tickSize)
        newNow := clock.Now()
        if !newNow.Equal(expectedTime) {
                t.Errorf("simclock progressed with %d tick size instead of %d", newNow.Sub(now), tickSize)
        }
}
