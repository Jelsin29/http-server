# Design: Milestone 4 - Concurrency and Error Responses

## Technical Approach

Keep the existing parse -> static load -> response build pipeline, but stop treating failures as "log and close" for everything. `Start` should launch `go handleConnection(conn)` so one slow client does not serialize the whole server. Inside `handleConnection`, request parse errors should map to a `400` response, static traversal and missing-file errors should stay `404`, and unexpected failures should map to a generic `500` response.

## Architecture Decisions

### Decision: Goroutine Per Accepted Connection

**Choice**: Spawn `go handleConnection(conn)` immediately after each successful `Accept`.
**Alternatives considered**: keep sequential handling, add a worker pool now.
**Rationale**: The learning goal is understanding the first concurrency step, not designing throughput controls yet. One goroutine per connection is the clearest change from milestone 3 and matches the proposal scope.

### Decision: Centralize Status Mapping in the Server Layer

**Choice**: Add one helper in `internal/server` that turns parse/static/build failures into `response.Message` values.
**Alternatives considered**: inline conditionals inside `handleConnection`, move error mapping into `static` or `response`.
**Rationale**: The server owns request lifecycle decisions. `static` should expose typed failures, and `response` should stay a formatter. Centralizing mapping avoids sprinkling `400`/`404`/`500` branches across the handler.

### Decision: Keep Client-Facing Error Bodies Generic

**Choice**: Return small plain-text bodies such as `bad request`, `not found`, and `internal server error`.
**Alternatives considered**: empty bodies, returning raw Go error text, JSON error payloads.
**Rationale**: This project is still learning raw HTTP fundamentals. Plain text keeps wire behavior visible, and generic `500` text avoids leaking server internals.

### Decision: Add Injectable Seams for Failure and Concurrency Tests

**Choice**: Introduce narrow package-level seams or helpers so tests can block static loading and force internal failures without depending on flaky timing or impossible filesystem states.
**Alternatives considered**: only sleep-based integration tests, trying to provoke real OS-level read failures against checked-in files.
**Rationale**: The new behavior is mostly about coordination and error mapping. Deterministic tests matter more than pretending the OS will reliably reproduce the exact failure we want.

## Data Flow

```text
listener.Accept()
   │
   ├─ error -> log and continue
   │
   └─ conn
      │
      └─ go handleConnection(conn)
            │
            ├─ request.Parse(conn)
            │    ├─ success -> continue
            │    └─ error -> write 400 response
            │
            ├─ buildResponse(req)
            │    ├─ static missing/traversal -> 404 message
            │    ├─ unexpected failure -> 500 message
            │    └─ success -> 200 message
            │
            ├─ conn.Write(response bytes)
            └─ defer conn.Close()
```

## Proposed Data Model

```go
package response

type Message struct {
    StatusCode int
    Reason     string
    Headers    map[string]string
    Body       []byte
}

func Build(msg Message) []byte
```

```go
package server

func handleConnection(conn net.Conn)
func buildResponse(req *request.Request) ([]byte, error)
func errorResponse(statusCode int, reason string, body string) []byte
func mapRequestError(err error) []byte
func mapServerError(err error) []byte
```

Exact helper names can change, but the contract should be simple: one place decides which HTTP status corresponds to which failure class, and one place constructs the standard headers/body for text responses.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/server/server.go` | Update | Run handlers in goroutines and map parse/static/internal failures to 400/404/500 |
| `internal/server/server_test.go` | Update | Add deterministic concurrency coverage plus 400/500 integration and handler tests |
| `internal/response/builder.go` | Update | Reuse existing message builder for shared error helpers if needed |
| `internal/response/builder_test.go` | Update | Cover new error-message formatting helpers if added there |
| `internal/static/static.go` | Update | Preserve typed failures cleanly for missing vs traversal vs unexpected read errors |
| `README.md` | Update | Record what changed and what broke during concurrency/error handling work |

## Interfaces / Contracts

```go
package static

var ErrTraversal = errors.New("path escapes public root")

type Asset struct {
    Body        []byte
    ContentType string
}

func Load(root string, target string) (*Asset, error)
```

`Load` should keep wrapping filesystem failures with context while still allowing `errors.Is(err, os.ErrNotExist)` and `errors.Is(err, ErrTraversal)` to work. Any other error becomes an internal failure from the server's point of view.

```go
package server

func Start(addr string) error
func handleConnection(conn net.Conn)
func buildResponse(req *request.Request) ([]byte, error)
```

`handleConnection` owns the full connection lifecycle: parse, classify outcome, write exactly one response when possible, then close. Unexpected failures are logged with context but clients only see generic text.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Parse failure -> 400 | stub connection sends malformed request, assert status/body and clean close |
| Unit | Missing/traversal -> 404 | keep direct `buildResponse` / handler coverage for existing typed errors |
| Unit | Unexpected internal failure -> 500 | inject failure seam around static load or response path |
| Unit | Error message formatting | assert `Content-Length`, `Content-Type`, CRLF separator, and generic body text |
| Integration | Concurrent handling | block first request in a test seam, prove second TCP client completes first |
| Integration | Malformed request over TCP | send bad request bytes, assert `HTTP/1.1 400 Bad Request` |
| Integration | Internal failure over TCP | inject deterministic server-side failure and assert `500 Internal Server Error` |

## Open Questions

1. Should unsupported HTTP methods start returning `405 Method Not Allowed` now?
   - **Decision for now**: no, keep this milestone focused on syntax errors, file misses, and internal failures.
2. Should the server attempt recovery on partial `conn.Write` results?
   - **Decision for now**: no, keep current single-write behavior and just log write failures. Reliable partial-write handling can be a later milestone.
