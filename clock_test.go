package simtest

import (
	"testing"
	"time"
)

func TestSimClockDoesNotProgressIfNotTicked(t *testing.T) {
        epoch := time.Now()
        clock := NewSimClock(epoch)
        defer clock.stop()
        
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
        tickSize := 1 * time.Millisecond
        numTicks := 1

        testSimClockProgressesAfterTick(t, tickSize, numTicks)
}

func FuzzSimClockProgressesAfterTick(f *testing.F) {
        f.Add(int64(1 * time.Millisecond), 1)

        f.Fuzz(func(t *testing.T, a int64, numTicks int) {
                tickSize := time.Duration(a)

                if numTicks <= 0 {
                        t.Skip("number of ticks is not positive")
                }

                if tickSize <= 0 {
                        t.Skip("tick size is not positive")
                }

                testSimClockProgressesAfterTick(t, tickSize, numTicks)
        })
}

func TestTimeMovesForwardWhenClockSleeps(t *testing.T) {
        tickSize := 1 * time.Millisecond
        sleepDuration := 1 * time.Second

        testTimeMovesForwardWhenClockSleeps(t, tickSize, sleepDuration)
}

func FuzzTimeMovesForwardWhenClockSleeps(f *testing.F) {
        f.Add(int64(1 * time.Microsecond), int64(1 * time.Second))
        f.Add(int64(1 * time.Microsecond), int64(1 * time.Millisecond))
        f.Add(int64(1 * time.Millisecond), int64(1 * time.Second))
        f.Add(int64(1 * time.Millisecond), int64(1 * time.Minute))

        f.Fuzz(func(t *testing.T, a ,b int64) {
                tickSize := time.Duration(a)
                sleepDuration := time.Duration(b)

                if tickSize <= 0 {
                        t.Skip("tick size is not positive")
                }
                
                if sleepDuration < 0 {
                        t.Skip("sleep duration is negative")
                }

                testTimeMovesForwardWhenClockSleeps(t, tickSize, sleepDuration)
        })
}

func TestTimerFiredWhileClockSleeping(t *testing.T) {
        testTimerFiredWhileClockSleeping(t, 10 * time.Second, 1 * time.Second, 1 * time.Millisecond)
}

func TestTimerFiredAtDeadline(t *testing.T) {
        tickSize := 100 * time.Millisecond
        timerDuration := 1 * time.Minute

        testTimerFiredAtDeadline(t, tickSize, timerDuration)
}

func FuzzTimerFiredAtDeadline(f *testing.F) {
        f.Add(int64(1 * time.Second), int64(1 * time.Minute))
        f.Add(int64(1 * time.Millisecond), int64(1 * time.Second))
        f.Add(int64(100 * time.Second), int64(1 * time.Hour))
        f.Add(int64(1 * time.Microsecond), int64(1 * time.Second))

        f.Fuzz(func(t *testing.T, a int64, b int64) {
                tickSize := time.Duration(a)
                timerDuration := time.Duration(b)

                if tickSize <= 0 {
                        t.Skip("tick size is not a positive number")
                }

                if timerDuration < 0 {
                        t.Skip("timer duration is a negative number")
                }

                testTimerFiredAtDeadline(t, tickSize, timerDuration)
        })
}

func TestStoppedTimerIsNotFired(t *testing.T) {
        now := time.Now()
        clock := NewSimClock(now)
        defer clock.stop()

        tickSize := 100 * time.Millisecond
        timerDuration := 1 * time.Minute
        deadline := now.Add(timerDuration)

        timer, ch := clock.NewTimer(timerDuration)
        timer.Stop()

        fired := false
        for now.Before(deadline) {
                select {
                case <- ch:
                        fired = true
                default:
                        now = clock.Tick(tickSize)
                }
        }

        if fired {
                t.Errorf("stopped timer fired")
        }
}

func FuzzTimerFiredWhileClockSleeping(f *testing.F) {
        f.Add(int64(10 * time.Second), int64(1 * time.Second), int64(1 * time.Millisecond))

        f.Fuzz(func(t *testing.T, sleepDuration, timerDuration, tickSize int64) {
                if sleepDuration <= timerDuration {
                        // timer duration needs to be lesser than sleep duration
                        t.SkipNow()
                }

                testTimerFiredWhileClockSleeping(t, time.Duration(sleepDuration), time.Duration(timerDuration), time.Duration(tickSize))
        })
}

func testSimClockProgressesAfterTick(t *testing.T, tickSize time.Duration, numTicks int) {
        epoch := time.Now()
        clock := NewSimClock(epoch)
        defer clock.stop()

        now := clock.Now()
        for range numTicks {
                actual := clock.Tick(tickSize)
                expected := now.Add(tickSize)

                if !actual.Equal(expected) {
                        t.Errorf("expected %s after tick (size=%s) but got %s", expected, tickSize, actual)
                }

                actual = clock.Now()

                if !actual.Equal(expected) {
                        t.Errorf("expected %s but got %s from Now", expected, actual)
                }

                now = actual
        }

}

func testTimeMovesForwardWhenClockSleeps(t *testing.T, tickSize, sleepDuration time.Duration) {
        epoch := time.Now()
        clock := NewSimClock(epoch)
        defer clock.stop()

        done := false
        doneCh := make(chan bool, 1)

        go func() {
                start := clock.Now()
                clock.Sleep(sleepDuration)
                end := clock.Now()
                dur := end.Sub(start)

                if dur < sleepDuration {
                        t.Errorf("expected to sleep for %s but actually slept for %s", sleepDuration, dur)
                }

                doneCh <- true
        }()

        for !done {
                select {
                case done = <- doneCh:
                default:
                        clock.Tick(tickSize)
                        // t.Logf("tick %s", clock.Tick(tickSize))
                }
        }

        now := clock.Now()

        if now.Sub(epoch) < sleepDuration {
                t.Errorf("expected at least %s to have passed but %s has passed", sleepDuration, now.Sub(epoch))
        }
}

func testTimerFiredWhileClockSleeping(t *testing.T, sleepDuration, timerDuration, tickSize time.Duration) {
        epoch := time.Now()
        clock := NewSimClock(epoch)
        defer clock.stop()

        fired := false
        slept := false
        sleptChannel := make(chan bool)
        firedChannel := make(chan bool)

        sleeperReady := false
        sleeperReadyCh := make(chan bool, 1)

        timerReady := false
        timerReadyCh := make(chan bool, 1)

        go func() {
                sleeperReadyCh <- true
                clock.Sleep(sleepDuration)

                sleptChannel <- true
        }()

        go func() {
                _, ch := clock.NewTimer(timerDuration)
                timerReadyCh <- true

                <-ch
                firedChannel <- true
        }()

        for !slept && !fired {
                select {
                case sleeperReady = <-sleeperReadyCh:
                case timerReady = <-timerReadyCh:
                case slept = <-sleptChannel:
                case fired = <-firedChannel:
                        if slept {
                                t.Errorf("timer did not fire while clock was still sleeping")
                        }
                default:
                        if sleeperReady && timerReady {
                                clock.Tick(tickSize)
                        }
                }
        }
}

func testTimerFiredAtDeadline(t *testing.T, tickSize, timerDuration time.Duration) {
        now := time.Now()
        clock := NewSimClock(now)
        defer clock.stop()

        doneCh := make(chan bool, 1)
        
        go func() {
                now := clock.Now()
                _, ch := clock.NewTimer(timerDuration)

                firedAt := <-ch
                actualDuration := firedAt.Sub(now)
                delta := actualDuration - timerDuration

                if actualDuration < timerDuration {
                        t.Errorf("expected timer to fire after %s but fired after %s, delta %s", timerDuration, actualDuration, delta)
                }

                doneCh <- true
        }()

        done := false
        for !done {
                select {
                case done = <-doneCh:
                default:
                        now = clock.Tick(tickSize)
                }
        }
}
