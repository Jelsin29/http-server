# Design: Milestone 1 - TCP Listener with Hardcoded HTTP Response

## Technical Approach

Minimal echo server implementing TCP socket lifecycle: `net.Listen` → infinite `Accept` loop → write hardcoded "HTTP/1.1 200 OK\r\n\r\n" → `conn.Close()`. No request reading, no parsing, no concurrency. Total ~20 lines across 3 packages following Go idioms from go-systems skill.

## Architecture Decisions

### Decision: Project Structure

**Choice**: `cmd/main.go` + `internal/server` + `internal/response`
**Alternatives considered**: Single file in root, flat structure with all code in cmd/
**Rationale**: Follows Go conventions (cmd/ for entry, internal/ for private packages). Separates concerns: server handles TCP lifecycle, response handles HTTP formatting. Scalable for future milestones.

### Decision: Connection Management

**Choice**: `defer conn.Close()` immediately after `ln.Accept()`
**Alternatives considered**: Explicit close after write, close in error handler
**Rationale**: Guarantees cleanup even if write fails. Prevents connection leaks. Idiomatic Go pattern.

### Decision: Hardcoded Response

**Choice**: Constant string in `response` package, no body
**Alternatives considered**: Response builder struct, dynamic status codes
**Rationale**: Milestone 1 is minimal — no flexibility needed. Response builder added in milestone 2 when parsing begins. Single `conn.Write()` call atomic for 18 bytes.

### Decision: Error Handling

**Choice**: Log Accept errors and continue, fatal on Listen error
**Alternatives considered**: Stop server on Accept error, panic on all errors
**Rationale**: Accept errors are transient (one bad connection shouldn't crash server). Listen errors are fatal (port in use = cannot start). Follows Go error philosophy: "log and continue" for recoverable, "fatal" for unrecoverable.

### Decision: Port Configuration

**Choice**: Hardcoded `:8080` for now, configurable in milestone 4
**Alternatives considered**: Environment variable, command-line flag
**Rationale**: Milestone 1 = minimal. Config adds complexity. Tests use port 0 (OS assigns). Config deferred to milestone 4 (production concerns).

## Data Flow

```
Client                           Server (internal/server)
  │                                   │
  │──── TCP Connect ──────────────────→│ Accept()
  │                                   │ defer conn.Close()
  │                                   │
  │                 ┌─────────────── connection established
  │                 │               │
  │←──── Write("HTTP/1.1 200 OK\r\n\r\n") ← response.Hardcoded()
  │                                   │
  │←──── TCP Close ──────────────────│ (deferred close)
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `go.mod` | Create | Go module `github.com/jelsin/http-server` |
| `cmd/main.go` | Create | Entry point, calls `server.Start(":8080")` |
| `internal/server/server.go` | Create | TCP listener + accept loop |
| `internal/server/server_test.go` | Create | Integration test with `net.Dial` |
| `internal/response/builder.go` | Create | Hardcoded response constant |
| `Makefile` | Create | Build/test/lint targets per go-systems skill |

## Interfaces / Contracts

```go
// internal/server/server.go
package server

import "net"

func Start(addr string) error {
    ln, err := net.Listen("tcp", addr)
    if err != nil {
        return fmt.Errorf("starting server on %s: %w", addr, err)
    }
    defer ln.Close()
    
    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Printf("accept error: %v", err)
            continue
        }
        go handleConnection(conn) // Milestone 4: goroutine per connection
    }
}

// Milestone 1: sequential (no goroutine yet)
func handleConnection(conn net.Conn) {
    defer conn.Close()
    conn.Write([]byte(response.HardcodedOK()))
}
```

```go
// internal/response/builder.go
package response

func HardcodedOK() string {
    return "HTTP/1.1 200 OK\r\n\r\n"
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Integration | Server accepts connection and sends 200 OK | `net.Dial` client connects, reads response, verifies exact string |
| Integration | Server handles port in use | Start server on :8080, attempt second start, verify error contains "address already in use" |
| Integration | Connection cleanup | Start server, connect, close connection, verify no goroutine leak (future: `runtime.NumGoroutine()`) |

Table-driven tests per go-systems skill:

```go
func TestServer(t *testing.T) {
    tests := []struct {
        name string
        port string
        want string
    }{
        {"default port", ":8080", "HTTP/1.1 200 OK\r\n\r\n"},
        {"port 0 (OS assigns)", ":0", "HTTP/1.1 200 OK\r\n\r\n"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Start server on tt.port, connect, read, verify tt.want
        })
    }
}
```

## Migration / Rollout

No migration required. Fresh project, no existing data.

## Open Questions

None. Design is minimal and clear.