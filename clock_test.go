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

func TestClockReturnsTimeItWasStoppedAt(t *testing.T) {
        epoch := time.Now()
        clock := NewSimClock(epoch)
        tickSize := 1 * time.Millisecond
        now := epoch

        for range 100 {
                now = clock.Tick(tickSize)
        }

        stoppedAt := clock.Stop()

        if !stoppedAt.Equal(now) {
                t.Errorf("expected clock to stop at %s but it stopped at %s", now, stoppedAt)
        }
}

func TestStoppedClockDoesNotCreateTimer(t *testing.T) {
        epoch := time.Now()
        clock := NewSimClock(epoch)
        clock.Stop()

        timer, ch := clock.NewTimer(1 * time.Second)

        if timer != nil {
                t.Errorf("stopped clock returned a timer")
        }

        if ch != nil {
                t.Errorf("stopped clock returned a timer channel")
        }
}

func TestStoppedClockDoesNotTick(t *testing.T) {
        epoch := time.Now()
        clock := NewSimClock(epoch)
        clock.Stop()
        tickSize := 1 * time.Millisecond

        now := clock.Tick(tickSize)

        if !epoch.Equal(now) {
                t.Errorf("stopped clock ticked")
        }

        clock = NewSimClock(epoch)

        for range 100 {
                now = clock.Tick(tickSize)
        }

        stoppedAt := clock.Stop()

        now = clock.Tick(tickSize)

        if !stoppedAt.Equal(now) {
                t.Errorf("stopped clock ticked")
        }
}

func TestTimeMovesForwardWhenClockSleeps(t *testing.T) {
        epoch := time.Now()
        clock := NewSimClock(epoch)
        tickSize := 1 * time.Millisecond
        sleepDuration := 1 * time.Second
        expectedTimeAfterSleeping := epoch.Add(sleepDuration)

        done := false
        doneCh := make(chan bool, 1)
        go func() {
                clock.Sleep(sleepDuration)
                doneCh <- true
        }()

        for !done {
                select {
                case <- doneCh:
                        done = true
                default:
                        clock.Tick(tickSize)
                }
        }

        now := clock.Now()

        if now.Before(expectedTimeAfterSleeping) {
                t.Errorf("expected time to be after %s but current time is %s", expectedTimeAfterSleeping, now)
        }
}

func TestTimersAreFiredWhileSleeping(t *testing.T) {
        testTimersAreFiredWhileSleeping(t, 10 * time.Second, 1 * time.Second, 1 * time.Millisecond)
}

func FuzzTimersAreFiredWhileSleeping(f *testing.F) {
        f.Add(int64(10 * time.Second), int64(1 * time.Second), int64(1 * time.Millisecond))

        f.Fuzz(func(t *testing.T, sleepDuration, timerDuration, tickSize int64) {
                if sleepDuration <= timerDuration {
                        // timer duration needs to be lesser than sleep duration
                        t.SkipNow()
                }

                testTimersAreFiredWhileSleeping(t, time.Duration(sleepDuration), time.Duration(timerDuration), time.Duration(tickSize))
        })
}

func testTimersAreFiredWhileSleeping(t *testing.T, sleepDuration, timerDuration, tickSize time.Duration) {
        epoch := time.Now()
        clock := NewSimClock(epoch)

        fired := false
        slept := false
        sleptChannel := make(chan bool, 1)
        _, firedChannel := clock.NewTimer(timerDuration)

        go func() {
                clock.Sleep(sleepDuration)
                sleptChannel <- true
        }()

        for !slept {
                select {
                case <-sleptChannel:
                        slept = true
                case <-firedChannel:
                        fired = true
                        if slept {
                                t.Errorf("timer did not fired while clock was still sleeping")
                        }
                default:
                        clock.Tick(tickSize)
                }
        }

        if slept && !fired {
                t.Errorf("clock finished sleeping but timer did not fire")
        }
}
