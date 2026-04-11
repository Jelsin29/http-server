# Proposal: Milestone 4 - Concurrency and Error Responses

## Intent

Milestone 3 can serve files, but it still handles one client at a time and silently drops some failures. This milestone makes the server feel more real: one goroutine per accepted connection, explicit `400` parse failures, existing `404` misses, and a `500` path for unexpected server-side errors.

## Scope

### In Scope
- Run each accepted connection in its own goroutine while keeping the accept loop alive
- Return `400 Bad Request` for malformed HTTP requests instead of closing with no response
- Keep `404 Not Found` for missing files / traversal attempts
- Return `500 Internal Server Error` when response building or file reads fail unexpectedly
- Add tests that prove concurrent handling and status-specific error bodies

### Out of Scope
- Keep-alive, pipelining, connection pooling, or worker pools
- Request body parsing, methods beyond current milestone needs, or structured logging

## Approach

Split connection handling into a small request/response pipeline that always tries to emit an HTTP response for known failures. `Start` will spawn `go handleConnection(conn)` after `Accept`. Response construction should move toward a helper that maps parse/static/build errors into `response.Message` values so status behavior stays consistent.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/server/server.go` | Modified | Goroutine-per-connection accept loop and error mapping |
| `internal/response/` | Modified | Shared helpers for 200/400/404/500 messages |
| `internal/static/` | Modified | Preserve typed errors for not-found vs internal failures |
| `internal/server/server_test.go` | Modified | Concurrency + error-response integration coverage |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Goroutine tests become flaky | Medium | Use blocking test doubles / synchronization instead of sleeps alone |
| Error mapping gets tangled | Medium | Centralize status generation in one helper |
| Internal errors leak details | Low | Send generic 500 body, keep specifics in logs only |

## Rollback Plan

Revert to sequential `handleConnection(conn)` calls, remove new error-mapping helpers, and restore malformed-request behavior plus milestone-4 tests.

## Dependencies

- Go standard library only: `net`, `sync`, `errors`, `log`

## Success Criteria

- [ ] Two clients can be served without one blocking the other's full request lifecycle
- [ ] Malformed requests receive `HTTP/1.1 400 Bad Request`
- [ ] Missing files and traversal attempts still receive `404 Not Found`
- [ ] Unexpected server-side failures return `500 Internal Server Error` without leaking internals
- [ ] `go test -race -v ./...` passes
