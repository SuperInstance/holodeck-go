# Holodeck Go — Charter

## Mission
Implement the Holodeck Studio in Go, leveraging goroutines and channels for natural concurrency.

## What to Build
A concurrent MUD server where each agent gets a goroutine, rooms communicate via channels, and the oversight cycle runs as a ticker.

## Architecture
```
cmd/holodeck/main.go   — entry point, TCP listener
pkg/
  room/      — room graph, exits, descriptions, runtime
  agent/     — agent session, command dispatch
  command/   — command parser and handlers
  comms/     — say/tell/yell/gossip/note/mailbox channels
  live/      — live connections (HTTP, shell)
  combat/    — oversight ticks, evolving scripts
  manual/    — living manual, generations
  conformance/ — test suite
```

## The Deep Question
What IS concurrency in a MUD? One goroutine per agent, channels for room events?
What IS a room — a channel hub that fans out messages to subscribers?
What does natural concurrency teach us about spatial agent interaction?
