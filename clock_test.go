package simtest

import (
	"testing"
	"time"
)

func TestClockTicking(t *testing.T) {
        tickRate := 100 * time.Nanosecond
        numTicks := 10

        testClockTicking(t, tickRate, numTicks)
}

func FuzzClockTicking(f *testing.F) {
        f.Add(uint64(1 * time.Nanosecond), 5)
        f.Add(uint64(1 * time.Nanosecond), 10)
        f.Add(uint64(1 * time.Nanosecond), 100)

        f.Add(uint64(100 * time.Nanosecond), 5)
        f.Add(uint64(100 * time.Nanosecond), 10)
        f.Add(uint64(100 * time.Nanosecond), 100)

        f.Add(uint64(1 * time.Microsecond), 5)
        f.Add(uint64(1 * time.Microsecond), 10)
        f.Add(uint64(1 * time.Microsecond), 100)

        f.Add(uint64(100 * time.Microsecond), 5)
        f.Add(uint64(100 * time.Microsecond), 10)
        f.Add(uint64(100 * time.Microsecond), 100)

        f.Add(uint64(1 * time.Millisecond), 5)
        f.Add(uint64(1 * time.Millisecond), 10)
        f.Add(uint64(1 * time.Millisecond), 100)

        f.Add(uint64(100 * time.Millisecond), 5)
        f.Add(uint64(100 * time.Millisecond), 10)
        f.Add(uint64(100 * time.Millisecond), 100)

        f.Fuzz(func(t *testing.T, a uint64, b int) {
                testClockTicking(t, time.Duration(a), b)
        })
}

func testClockTicking(t *testing.T, tickRate time.Duration, numTicks int) {
        now := time.Now()
        expected := now.Add(tickRate)

        clock := NewSimClock(now, tickRate)
        defer clock.Stop()

        for range numTicks {
                clock.Tick()
                now = clock.Now()

                if now != expected {
                        t.Fatalf("expected %s but got %s", expected, now)
                }

                expected = now.Add(tickRate)
        }
}

func TestClockSleep(t *testing.T) {
        now := time.Now()
        tickRate := 100 * time.Nanosecond
        sleepDuration := 1 * time.Millisecond

        expected := now.Add(sleepDuration)
        
        done := make(chan struct{})
        clock := NewSimClock(now, tickRate)
        defer clock.Stop()

        go func(clock *SimClock, d time.Duration, done chan struct{}) {
                clock.Sleep(d)
                done <- struct{}{}
        }(clock, sleepDuration, done)

        for {
                select {
                case <-done:
                        now = clock.Now()

                        if now.Before(expected) {
                                t.Fatalf("expected %s but got %s", expected, now)
                        }
                        
                        return
                default:
                        clock.Tick()
                }
        }
}

func TestClockAfterFunc(t *testing.T) {
        now := time.Now()
        tickRate := 100 * time.Nanosecond

        delay := 1 * time.Millisecond
        expected := now.Add(delay)

        clock := NewSimClock(now, tickRate)
        defer clock.Stop()

        done := make(chan time.Time)
        clock.AfterFunc(delay, func(now time.Time) {
                go func() {
                        done <- now
                }()
        })

        for {
                select {
                case now = <-done:
                        if now.Before(expected) {
                                t.Fatalf("expected %s but got %s", expected, now)
                        }

                        return
                default:
                        clock.Tick()
                }
        }
}
