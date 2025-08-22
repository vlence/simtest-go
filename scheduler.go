package simtest

import "runtime"

// Use this to schedule tasks and yield in production.
var GoRuntime Scheduler = &goRuntime{}

// Scheduler can schedule tasks and 
type Scheduler interface {
	// Go schedules the given task.
	Go(task func())

	// Yield pauses the current task and allows another task to execute.
	Yield()
}

type goRuntime struct{}

// Go runs task as a goroutine.
func (*goRuntime) Go(task func()) {
	go task()
}

// Yield calls runtime.Gosched.
func (*goRuntime) Yield() {
	runtime.Gosched()
}

// Status of a simulated task.
type taskStatus int

const (
	taskQueued taskStatus = iota
	taskRunning
	taskBlocked
	taskStopped
)

// A simulated task. Tasks are just functions that are waiting
// to be executed, waiting for some event or have finished. A
// panic inside the task is not recovered so that it can be
// captured during testing.
type simTask struct {
	run func()
	yield chan bool
	status taskStatus
}

// A simulated scheduler. Implements the Scheduler interface.
type SimScheduler struct {
	clock *SimClock

	yield chan chan bool
	newTask chan *simTask

	nextTask int
	taskDone chan bool
	runningTasks []*simTask
	blockedTasks []*simTask
}

// NewSimScheduler returns a new simulated scheduler. firstTask is executed
// in the first call to Tick.
func NewSimScheduler(firstTask func(), clock *SimClock) *SimScheduler {
	scheduler := new(SimScheduler)
	scheduler.clock = clock
	scheduler.yield = make(chan chan bool)
	scheduler.newTask = make(chan *simTask)
	scheduler.nextTask = 0
	scheduler.taskDone = make(chan bool)
	scheduler.runningTasks = []*simTask{scheduler.newSimTask(firstTask)}
	scheduler.blockedTasks = []*simTask{}

	return scheduler
}

func (scheduler *SimScheduler) newSimTask(fn func()) *simTask {
	return &simTask{
		func() {
			fn()

			scheduler.taskDone <- true
		},
		nil,
		taskQueued,
	}
}

// Go creates a new simulated task for the given function
// and schedules it. Do not call Go in the same goroutine
// as Tick; you will deadlock.
func (scheduler *SimScheduler) Go(fn func()) {
	scheduler.newTask <- scheduler.newSimTask(fn)
}

// Yield blocks until the current task executed again. Do
// not call Yield in the same goroutine as Tick; you will
// deadlock.
func (scheduler *SimScheduler) Yield() {
	wait := make(chan bool)
	scheduler.yield <- wait
	<-wait
}

// Tick moves this scheduler forward by picking the next task
// from the running queue and executing it. If all tasks in
// the running queue have been executed then the blocked tasks
// are slated to be executed next.
func (scheduler *SimScheduler) Tick() {
	task := scheduler.runningTasks[scheduler.nextTask]

	// todo: assert that task is not waiting for yield, clock and io simultaneously

	if taskQueued == task.status {
		go task.run()
	}

	task.status = taskRunning

	if task.yield != nil {
		task.yield <- true
		task.yield = nil
	}

	// todo: fire clock event

	// todo: fulfil io request

	// todo: wait for clock and io event
	// todo: tick the clock
	select {
	case newTask := <-scheduler.newTask:
		newTask.status = taskQueued
		scheduler.blockedTasks = append(scheduler.blockedTasks, newTask)
	case yield := <-scheduler.yield:
		task.yield = yield
		scheduler.blockedTasks = append(scheduler.blockedTasks, task)
	case <-scheduler.taskDone:
		task.status = taskStopped
	}

	scheduler.nextTask++

	if len(scheduler.runningTasks) == scheduler.nextTask {
		scheduler.runningTasks, scheduler.blockedTasks = scheduler.blockedTasks, scheduler.runningTasks[:0]
	}
}