package simtest

import (
	"time"

	"github.com/vlence/gossert"
)

// A timer. Timers may be called just once or multiple times
// at regular intervals. They can be stopped or reset.
type Timer interface {
	// Stops this timer and returns true. If this timer
	// is already stopped then it returns false.
	Stop() bool

	// Resets this timer to be called after the given
	// duration instead and returns true. If the timer
	// is already stopped then this function returns
	// false. Resetting a stopped timer will not restart
	// it.
	Reset(time.Duration) bool
}

type TimerCallback func(time.Time)

type Clock interface {
	// Now returns the current time.
	Now() time.Time

	// AfterFunc calls the given function with the current
	// after the given duration of time has passed. The given
	// function is run in a separate goroutine.
	AfterFunc(time.Duration, TimerCallback) Timer

	// TickFunc calls the given function after every interval
	// of the given duration. The given function is run in a
	// separate goroutine.
	TickFunc(time.Duration, TimerCallback) Timer

	// Sleep blocks this goroutine until the given amount
	// of time has passed.
	Sleep(time.Duration)
}

type SimClock struct {
	now      time.Time
	tickRate time.Duration

	tick        chan struct{}
	stop        chan struct{}
	currentTime chan time.Time

	newTickTimer  chan newTimerRequest
	newAfterTimer chan newTimerRequest
	newSleepTimer chan newTimerRequest

	stopTimer  chan stopTimerRequest
	resetTimer chan resetTimerRequest

	nextHandle chan simTimerHandle
}

type simTimer struct {
	gen       uint32
	ticksLeft time.Duration
	callback  TimerCallback
	repeat    time.Duration
}

type newTimerRequest struct {
	d        time.Duration
	callback TimerCallback
}

type stopTimerRequest struct {
	simTimerHandle
	ok bool
}

type resetTimerRequest struct {
	simTimerHandle
	ok       bool
	tickRate time.Duration
}

type simTimerAndHandle struct {
	*simTimer
	simTimerHandle
}

func NewSimClock(now time.Time, tickRate time.Duration) *SimClock {
	clock := &SimClock{
		now:      now,
		tickRate: tickRate,

		tick:        make(chan struct{}),
		stop:        make(chan struct{}),
		currentTime: make(chan time.Time),

		stopTimer:  make(chan stopTimerRequest),
		resetTimer: make(chan resetTimerRequest),

		newTickTimer:  make(chan newTimerRequest),
		newAfterTimer: make(chan newTimerRequest),
		newSleepTimer: make(chan newTimerRequest),

		nextHandle: make(chan simTimerHandle),
	}

	go clock.start()

	return clock
}

// Now returns the current time according to this clock.
func (clock *SimClock) Now() time.Time {
	return <-clock.currentTime
}

// AfterFunc executes callback after d amount of time has passed according to this
// clock.
func (clock *SimClock) AfterFunc(d time.Duration, callback TimerCallback) Timer {
	clock.newAfterTimer <- newTimerRequest{d, callback}
	return <-clock.nextHandle
}

// TickFunc executes callback at regular intervals of d amount of time.
func (clock *SimClock) TickFunc(d time.Duration, callback TimerCallback) Timer {
	clock.newTickTimer <- newTimerRequest{d, callback}
	return <-clock.nextHandle
}

// Sleep will block the current goroutine for d amount of time.
func (clock *SimClock) Sleep(d time.Duration) {
	done := make(chan struct{})

	clock.newSleepTimer <- newTimerRequest{
		d,
		func(time.Time) {
			done <- struct{}{}
		},
	}
	<-clock.nextHandle

	<-done
}

// Tick moves the clock forward by the configured tick rate
// and schedules callbacks of timers that have expired.
func (clock *SimClock) Tick() {
	clock.tick <- struct{}{}
	<-clock.tick
}

// Stop stops this clock. It is incorrect to use a clock after it has been stopped.
// All operations on a stopped clock will be blocked forever.
func (clock *SimClock) Stop() {
	clock.stop <- struct{}{}
}

func (clock *SimClock) start() {
	timers := make([]simTimer, 0)

	// buffered channel containing the most recently freed timer
	freedTimer := make(chan uint32, 1)

	freeTimers := make([]uint32, 0)  // list of all timers that can be reused
	inUseTimers := make([]uint32, 0) // list of all timers currently in use

	nextSimTimerAndHandle := clock.nextSimTimerAndHandle(&timers, &freeTimers)

	for {
		select {
		case req := <-clock.newAfterTimer:
			timer := nextSimTimerAndHandle.simTimer

			timer.ticksLeft = req.d / clock.tickRate
			timer.callback = req.callback

			inUseTimers = append(inUseTimers, nextSimTimerAndHandle.idx)
			nextSimTimerAndHandle = clock.nextSimTimerAndHandle(&timers, &freeTimers)

			clock.nextHandle <- nextSimTimerAndHandle.simTimerHandle

		case req := <-clock.newTickTimer:
			timer := nextSimTimerAndHandle.simTimer

			timer.ticksLeft = req.d / clock.tickRate
			timer.callback = req.callback
			timer.repeat = req.d

			inUseTimers = append(inUseTimers, nextSimTimerAndHandle.idx)
			nextSimTimerAndHandle = clock.nextSimTimerAndHandle(&timers, &freeTimers)

			clock.nextHandle <- nextSimTimerAndHandle.simTimerHandle

		case req := <-clock.newSleepTimer:
			timer := nextSimTimerAndHandle.simTimer

			timer.ticksLeft = req.d / clock.tickRate
			timer.callback = req.callback

			inUseTimers = append(inUseTimers, nextSimTimerAndHandle.idx)
			nextSimTimerAndHandle = clock.nextSimTimerAndHandle(&timers, &freeTimers)

			clock.nextHandle <- nextSimTimerAndHandle.simTimerHandle

		case clock.currentTime <- clock.now:

		case <-clock.tick:
			clock.now = clock.now.Add(clock.tickRate)

			for _, idx := range inUseTimers {
				timer := &timers[idx]

				gossert.Ok(timer.ticksLeft >= 0, "clock: timer has negative ticks left")
				if timer.ticksLeft > 0 {
					timer.ticksLeft--
					continue
				}

				gossert.Ok(timer.callback != nil, "clock: timer callback is nil")
				timer.callback(clock.now)

				if timer.repeat > 0 {
					timer.ticksLeft = timer.repeat / clock.tickRate
					continue
				}

				freedTimer <- idx
			}

			clock.tick <- struct{}{}

		case req := <-clock.stopTimer:
			timer := &timers[req.idx]

			if timer.gen != req.gen {
				req.ok = false
			} else {
				req.ok = true
			}

			freedTimer <- req.idx
			clock.stopTimer <- req

		case req := <-clock.resetTimer:
			timer := &timers[req.idx]

			if timer.gen != req.gen {
				req.ok = false
			} else {
				timer.ticksLeft = req.tickRate / clock.tickRate

				if timer.repeat > 0 {
					timer.repeat = req.tickRate
				}

				req.ok = true
			}

			freedTimer <- req.idx
			clock.resetTimer <- req

		case idx := <-freedTimer:
			timer := &timers[idx]

			timer.gen++
			timer.repeat = 0
			timer.callback = nil
			timer.ticksLeft = 0

			freeTimers = append(freeTimers, idx)

			for i := range inUseTimers {
				if inUseTimers[i] == idx {
					for j := len(inUseTimers) - 1; i < j; i++ {
						inUseTimers[i] = inUseTimers[i+1]
					}

					inUseTimers = inUseTimers[:len(inUseTimers)-1]

					break
				}
			}

		case <-clock.stop:
			return
		}
	}
}

func (clock *SimClock) nextSimTimerAndHandle(timersPtr *[]simTimer, freeTimersPtr *[]uint32) simTimerAndHandle {
	var timer *simTimer
	var timerHandle simTimerHandle

	timers := *timersPtr
	freeTimers := *freeTimersPtr

	timerHandle.clock = clock

	if len(freeTimers) > 0 {
		lastIndex := len(freeTimers) - 1
		timerIndex := freeTimers[lastIndex]

		timer = &timers[timerIndex]
		timer.gen++
		timer.repeat = 0
		timer.callback = nil

		timerHandle.idx = timerIndex
		timerHandle.gen = timer.gen

		freeTimers = freeTimers[:lastIndex]
	} else {
		lastIndex := len(timers)

		timers = append(timers, simTimer{})

		timer = &timers[lastIndex]

		timerHandle.idx = uint32(lastIndex)
		timerHandle.gen = timer.gen
	}

	*timersPtr = timers
	*freeTimersPtr = freeTimers

	return simTimerAndHandle{timer, timerHandle}
}

type simTimerHandle struct {
	idx   uint32
	gen   uint32
	clock *SimClock
}

func (handle simTimerHandle) Stop() bool {
	req := stopTimerRequest{handle, false}

	handle.clock.stopTimer <- req
	req = <-handle.clock.stopTimer

	return req.ok
}

func (handle simTimerHandle) Reset(d time.Duration) bool {
	req := resetTimerRequest{handle, false, d}

	handle.clock.resetTimer <- req
	req = <-handle.clock.resetTimer

	return req.ok
}
