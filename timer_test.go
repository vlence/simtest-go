package simtest

import (
	"math/rand/v2"
	"testing"
	"time"
)

func TestTimerHasExpectedDeadline(t *testing.T) {
        epoch := time.Now()
        clock := NewSimClock(epoch)
        defer clock.Stop()

        d := 1 * time.Second
        expectedDeadline := epoch.Add(d)

        tt, _ := clock.NewTimer(d)
        timer, _ := tt.(*SimTimer)

        if !expectedDeadline.Equal(timer.deadline) {
                t.Errorf("timer's deadline %s does not match expected deadline %s", timer.deadline, expectedDeadline)
        }
}

func TestTimerIsFired(t *testing.T) {
        epoch := time.Now()
        clock := NewSimClock(epoch)
        defer clock.Stop()

        dur := 1 * time.Second
        _, ch := clock.NewTimer(dur)
        expectedDeadline := epoch.Add(dur)

        tickSize := 100 * time.Millisecond
        for range 100 {
                select {
                case now := <-ch:
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
        _, ch := clock.NewTimer(dur)

        fired := false
        tickSize := 100 * time.Millisecond
        for range 100 {
                select {
                case <-ch:
                        if fired {
                                t.Errorf("timer fired twice")
                        }

                        fired = true
                default:
                        clock.Tick(tickSize)
                }
        }

        if !fired {
                t.Errorf("timer not fired")
        }
}

func FuzzTestTimerFiredOnlyOnce(f *testing.F) {
        minResolution := int64(time.Microsecond)
        maxResolution := int64(time.Second)

        for range 10 {
                mul := rand.Int64N(9) + 1
                res := rand.Int64N(maxResolution - minResolution) + minResolution
                tickSize := mul * res

                f.Add(tickSize)
        }

        f.Fuzz(func(t *testing.T, a int64) {
                epoch := time.Now()
                clock := NewSimClock(epoch)
                defer clock.Stop()

                tickSize := time.Duration(a)
                durMul := rand.Int64N(8) + 2 // duration is 2-10x the tickSize
                dur := time.Duration(durMul * a)

                minTicks := dur / tickSize // min number of ticks that will be required to fire the timer
                maxTicks := minTicks + 100 // max iterations we are willing to wait

                _, ch := clock.NewTimer(dur)

                fired := false
                for range maxTicks {
                        select {
                        case <-ch:
                                if fired {
                                        t.Errorf("timer fired twice")
                                }

                                fired = true
                        default:
                                clock.Tick(tickSize)
                        }
                }

                if !fired {
                        t.Errorf("timer not fired")
                }
        })
}
