# Holodeck Go

Concurrent Go implementation of the FLUX-LCAR holodeck protocol.

## What It Teaches

What IS concurrency in a MUD? One goroutine per agent. Channels for room events. `sync.RWMutex` for room graph access — multiple readers OR one writer. The Go `select` statement IS the event loop.

## Build

```bash
go build ./...     # Build
go test ./...      # Run tests
```

## Status

**17/40 conformance tests — Operational** 🟡

## Architecture

```
cmd/holodeck/main.go       — TCP listener, goroutine per connection
pkg/room/room.go           — concurrent room graph with sync.RWMutex
pkg/agent/agent.go         — agent session, command dispatch
pkg/comms/comms.go         — say/tell/yell/gossip channels
pkg/conformance/           — conformance test suite
```

## Run

```bash
go run cmd/holodeck/main.go  # Starts on :7779
```

## Dependencies

Go 1.24+. Standard library only.
