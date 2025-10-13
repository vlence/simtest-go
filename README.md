# Deterministic Simulation Testing in Go

This library provides some primitives for writing tests that deterministically simulate real world
faults for e.g.
- Corrupted data
- High latency

The usefulness of this library doesn't come from merely injecting these faults but being
able to reproduce them thereby being deterministic.

`simtest` is heavily inspired by TigerBeetle's VOPR.

If you're new to the idea of deterministic simulation testing I recommend you watch this video
of [ThePrimeagen talking with Joran Greef](https://www.youtube.com/watch?v=sC1B3d9C_sI).

## The Main Idea

Determinism means that we can predict the behaviour of a system given its initial conditions. In
terms of software it means we can predict the output of 