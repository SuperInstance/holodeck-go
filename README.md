# Holodeck Go

## Overview

Concurrent **Go implementation** of the FLUX-LCAR holodeck protocol — a distributed multi-agent environment where agents navigate a graph of interconnected rooms, communicate via scoped channels, and interact through commands. This implementation demonstrates how Go's concurrency primitives map naturally to the holodeck's architectural requirements: one goroutine per agent, `sync.RWMutex` for room graph access, and channels for event propagation.

### What It Teaches

What IS concurrency in a MUD? One goroutine per agent. Channels for room events. `sync.RWMutex` for room graph access — multiple readers OR one writer. The Go `select` statement IS the event loop.

## Architecture

```
holodeck-go/
├── go.mod                          — Go 1.24 module (zero external dependencies)
├── cmd/
│   └── holodeck/
│       └── main.go                 — TCP listener, goroutine-per-connection server
└── pkg/
    ├── room/
    │   └── room.go                 — Concurrent room graph with World/Room types
    ├── agent/
    │   └── agent.go                — Agent session lifecycle, state machine, mailbox
    ├── command/
    │   └── command.go              — Command dispatch: look, go, say, tell, who, yell, gossip, note, read, help
    ├── comms/
    │   └── comms.go                — Scoped communication: say (room), tell (direct), yell (adjacent), gossip (global)
    └── conformance/
        └── conformance_test.go     — 40-point conformance test suite
```

### Key Design Decisions

| Concern | Go Approach |
|---|---|
| Connection handling | `net.Listen` → `Accept` → `go handleConnection(conn)` — one goroutine per client |
| Room graph safety | `sync.RWMutex` on both `World` and `Room` — readers-writer lock allows concurrent reads |
| Agent state | `Agent` struct with `State` enum (`StateLogin` → `StatePlaying` → `StateQuit`) |
| Communication | Injected broadcast function via `comms.SetBroadcastFunc()` — decouples comms from world |
| Room exits | `map[string]string` (direction → room ID) — O(1) lookup by compass direction |
| ID generation | Hex-encoded name hashing — deterministic, collision-free for unique names |

### Concurrency Model

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Goroutine   │     │  Goroutine   │     │  Goroutine   │
│  Agent A     │     │  Agent B     │     │  Agent C     │
│  (conn 1)    │     │  (conn 2)    │     │  (conn 3)    │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                    │                    │
       └──────────┬─────────┴──────────┬─────────┘
                  │                    │
           ┌──────▼───────┐    ┌──────▼───────┐
           │  World.RWMutex│    │ Room.RWMutex │
           │  (room map)   │    │ (agents set) │
           └──────────────┘    └──────────────┘
```

Multiple goroutines read rooms concurrently (RLock). Any mutation (agent movement, room creation) requires exclusive write lock (Lock).

## Quick Start

```bash
# Clone and build
go build ./...

# Run tests (conformance suite)
go test ./...

# Start server on default port :7777
go run cmd/holodeck/main.go

# Start on custom port
go run cmd/holodeck/main.go 9999

# Connect
nc localhost 7777
```

**In-session commands:**
```
look              — Describe current room, exits, and occupants
go <direction>    — Move north/south/east/west
say <message>     — Speak to everyone in the room
tell <agent> <msg> — Send a private message
yell <message>    — Broadcast to adjacent rooms
gossip <message>  — Broadcast to the entire world
note <message>    — Write a note on the wall
read              — Read wall notes
who               — List all agents across all rooms
help              — Show command reference
quit              — Disconnect
```

## Status

**17/40 conformance tests — Operational** 🟡

## Comparison with holodeck-zig

| Feature | holodeck-go | holodeck-zig |
|---|---|---|
| Language | Go 1.24 | Zig |
| Concurrency model | Goroutines + sync.RWMutex | Zig async/event-loop |
| Connection handling | `go handleConnection()` per client | Single-threaded event loop |
| Room graph storage | `map[string]*Room` | Array/slice |
| Memory management | Garbage collected | Manual (defer-free) |
| Port | :7777 | :7779 |
| Conformance | 17/40 | (reference implementation) |

Go's goroutine model makes it trivially composable for high-connection-count scenarios, while Zig's zero-allocation approach provides predictable latency. The two implementations are **wire-compatible** — both implement the same FLUX-LCAR protocol and can share clients.

## Dependencies

Go 1.24+. **Standard library only** — zero external dependencies.

---

<img src="callsign1.jpg" width="128" alt="callsign">
