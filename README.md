# Deterministic Simulation Testing in Go

This library provides some primitives for writing tests that simulate real world faults for e.g.
- Corrupted data
- High latency

`simtest` is heavily inspired by TigerBeetle's VOPR.

If you're new to the idea of deterministic simulation testing I recommend you watch this video
of [ThePrimeagen talking with Joran Greef](https://www.youtube.com/watch?v=sC1B3d9C_sI).

## The Main Idea

Unit tests are great for testing the behaviour of your application. When you find a new bug you
can write another unit test to test for that bug's presence. If in the future you refactor or
add a new feature and the bug appears again the unit test will catch. However unit tests cannot
find new bugs.

Fuzzing can be used to find new bugs. In case you don't know what fuzzing is it just means
generating random data and observing how your application behaves. Fuzzing is a great way to find
how well our applications can handle data that it's not expecting. When new bugs are found by
fuzzing we can then create unit tests for them. 

To really put our application through the paces we need to be able to simulate that kind of faults
that can happen in the real world. In the real world tasks can take longer than expected, the
network is not reliable, disks fail, etc. One way to simulate these kind of faults is to use
`io.Reader` and `io.Writer` implementations that inject these faults. For example to simulate
high read latency one could wait for some time before starting the actual read operation.

```go
type HighLatencyReader {
    io.Reader
}

func NewHighLatencyReader(r io.Reader) {
    return &HighLatencyReader{r}
}

func (r *HighLatencyReader) Read(p []byte) (int, error) {
    time.Sleep(1 * time.Second)
    return r.Reader.Read(p)
}
```

There's however the problem of time as well. Consider our previous example of introducing latency
into a read operations. We are waiting for 1 second. If we are running this in our tests our test
will need to wait 1 second for the faulty read to complete. We need some way to make our application
think time has passed without actually waiting that long.

It's fairly common to think of things like the network and disk as dependencies and to mock them
during tests. What is uncommon, as far as I know, is to think of time itself as some kind of
dependency. If we can simulate time then we can make it go faster, call timers and ticks sooner,
sleep faster, etc. A simple way to do this is to have an interface for a clock that can tell the
current time and create timers and suchlike. The clock can tick at a rate of our choosing and we
can make time go faster. Our earlier example can be rewritten like this:

```go
type HighLatencyReader {
    io.Reader
    clock Clock
}

func NewHighLatencyReader(r io.Reader, clock *Clock) {
    return &HighLatencyReader{r, clock}
}

func (r *HighLatencyReader) Read(p []byte) (int, error) {
    clock.Sleep(1 * time.Second)
    return r.Reader.Read(p)
}
```
