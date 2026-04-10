# Design: Milestone 2 - Parse HTTP Request Line and Headers

## Technical Approach

Introduce a small `internal/request` package responsible for turning a byte stream into a `Request` struct. The server remains the TCP owner: it accepts a connection, hands the socket reader to the parser, and only writes the response once parsing succeeds.

## Architecture Decisions

### Decision: Dedicated `request` Package

**Choice**: Put parsing code in `internal/request`
**Alternatives considered**: keep parsing inline in `internal/server`, use `net/http` request parsing helpers
**Rationale**: Request parsing is its own concept. Keeping it separate prevents `server.go` from turning into protocol soup and leaves room for routing/file serving later.

### Decision: Tiny `Request` Struct

**Choice**: `Request{Method, Target, Version, Headers}`
**Alternatives considered**: full-body support, canonical header type, raw byte slices everywhere
**Rationale**: Milestone 2 only needs start-line and headers. Anything beyond that is architecture cosplay.

### Decision: Buffered Line-Oriented Parsing

**Choice**: `bufio.Reader` + `ReadString('\n')`
**Alternatives considered**: manual byte-by-byte scanner, `bufio.Scanner`
**Rationale**: HTTP/1.1 headers are line oriented. `Scanner` has token size limits and hides too much. `ReadString` keeps the implementation explicit without being painful.

### Decision: Strict-enough Parsing

**Choice**: Require three request-line parts and `name: value` headers; trim `\r\n` and optional surrounding spaces
**Alternatives considered**: fully lenient whitespace parsing, complete RFC grammar support
**Rationale**: This is a learning project, not nginx. The parser should teach the shape of the protocol, not disappear under edge-case machinery.

## Data Model

```go
type Request struct {
    Method  string
    Target  string
    Version string
    Headers map[string]string
}
```

Header names are stored exactly as received for now. Canonicalization can come later if route matching or semantics need it.

## Data Flow

```text
Client bytes
   │
   ▼
server.handleConnection(conn)
   │
   ├─ request.Parse(conn)
   │    ├─ read request line
   │    ├─ split method / target / version
   │    ├─ read headers until blank line
   │    └─ return Request or error
   │
   ├─ on parse error: log + close
   └─ on success: conn.Write(response.HardcodedOK())
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/request/request.go` | Create | Request struct |
| `internal/request/parser.go` | Create | Request parser |
| `internal/request/parser_test.go` | Create | Table-driven parser tests |
| `internal/server/server.go` | Update | Parse request before responding; log parse/write errors |
| `internal/server/server_test.go` | Update | Real request/response integration tests |
| `README.md` | Update | Learning notes for milestone 2 |

## Interfaces / Contracts

```go
package request

import "io"

type Request struct {
    Method  string
    Target  string
    Version string
    Headers map[string]string
}

func Parse(r io.Reader) (*Request, error)
```

```go
package server

func handleConnection(conn net.Conn) {
    defer conn.Close()

    if _, err := request.Parse(conn); err != nil {
        log.Printf("parsing request from %s: %v", conn.RemoteAddr(), err)
        return
    }

    if _, err := conn.Write([]byte(response.HardcodedOK())); err != nil {
        log.Printf("writing response to %s: %v", conn.RemoteAddr(), err)
    }
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Request line + header parsing | `strings.NewReader`, table-driven cases |
| Unit | Malformed input | Missing version, bad header, early EOF |
| Integration | Valid TCP request gets 200 response | `net.Dial`, send raw HTTP request, read response |
| Integration | Malformed request closes connection | `net.Dial`, send bad bytes, expect EOF/no response |

## Open Questions

1. Should milestone 2 read request bodies when `Content-Length` is present?
   - **Decision for now**: no, keep scope to header-only requests.
2. Should malformed requests get a `400 Bad Request` response now?
   - **Decision for now**: no, log + close; formal error responses belong in milestone 4.
