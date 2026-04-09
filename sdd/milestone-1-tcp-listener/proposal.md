# Proposal: Milestone 1 - TCP Listener with Hardcoded HTTP Response

## Intent

Build the foundation of an HTTP server by implementing a minimal TCP listener that accepts connections and responds with hardcoded "HTTP/1.1 200 OK\r\n\r\n". This teaches the TCP socket lifecycle (Listen → Accept → Write → Close) before adding HTTP parsing complexity (milestone 2).

## Scope

### In Scope
- TCP listener on port 8080 using `net.Listen("tcp", ":8080")`
- Infinite accept loop handling incoming connections
- Hardcoded HTTP/1.1 200 OK response (no body)
- Proper connection cleanup with `defer conn.Close()`
- Integration test using `net.Dial` client

### Out of Scope
- Request parsing (milestone 2)
- Concurrent connections (milestone 4)
- Static file serving (milestone 3)
- HTTP error responses (400, 404, 500)
- Multiple routes or path handling

## Approach

Implement a minimal echo server structure:
```
cmd/main.go
  └─ internal/server/listener.go — Listen + Accept loop
      └─ internal/response/builder.go — Hardcoded "HTTP/1.1 200 OK\r\n\r\n"
```

Server accepts connection, immediately writes hardcoded response, closes connection. No request reading (that's milestone 2). Total implementation: ~20 lines of Go.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/main.go` | New | Entry point, starts server on :8080 |
| `internal/server/listener.go` | New | TCP listener + accept loop |
| `internal/response/builder.go` | New | Hardcoded HTTP response builder |
| `go.mod` | New | Go module initialization |
| `Makefile` | New | Build/test/lint targets |
| `internal/server/listener_test.go` | New | Integration test with net.Dial client |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|-------------|
| Connection leak if write fails before close | Low | `defer conn.Close()` immediately after Accept |
| Port 8080 already in use | Medium | Return clear error from Listen, test can use port 0 (OS assigns) |
| Partial writes (TCP fragmentation) | Low | Response is ~18 bytes, fits in single packet |

## Rollback Plan

Delete all files created (cmd/, internal/, go.mod, Makefile). No database migrations, no config changes, no shared state.

## Dependencies

- Go standard library: `net`, `io`, `fmt`
- No external dependencies

## Success Criteria

- [ ] Server starts and listens on port 8080
- [ ] `curl http://localhost:8080` returns "200 OK" (empty body)
- [ ] `nc localhost 8080` receives "HTTP/1.1 200 OK\r\n\r\n"
- [ ] Integration test passes: client connects, sends arbitrary bytes, receives 200 OK
- [ ] `go vet ./...` passes
- [ ] `go test -race -v ./...` passes
- [ ] Proper error handling on Listen failure (port in use)
- [ ] Connection cleanup verified (no goroutine leaks)