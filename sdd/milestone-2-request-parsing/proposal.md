# Proposal: Milestone 2 - Parse HTTP Request Line and Headers

## Intent

Teach the next layer above raw TCP by reading an HTTP/1.1 request from the connection, parsing the request line plus headers, and only then sending the hardcoded 200 OK response. Milestone 1 proved the socket lifecycle; milestone 2 proves how HTTP messages are delimited on top of that byte stream.

## Scope

### In Scope
- Parse request line: method, target, version
- Parse headers until the empty line that ends the header section
- Represent parsed request as a small Go struct in `internal/request`
- Wire server handler to read and parse a request before writing the response
- Add tests for valid requests and malformed request lines / headers
- Keep response body empty and status code fixed at 200 OK for valid requests

### Out of Scope
- Routing by path (milestone 3+)
- Static file serving (milestone 3)
- Concurrent connection handling (milestone 4)
- Chunked transfer encoding or streaming bodies
- Full HTTP compliance for every request-target form
- Rich error responses (400/404/500) sent to clients

## Approach

Add a `request` package that parses raw bytes using `bufio.Reader`.

```text
internal/request/parser.go
  ├─ Parse(reader io.Reader) (*Request, error)
  ├─ parseRequestLine(line string)
  └─ parseHeaders(reader *bufio.Reader)

internal/server/server.go
  └─ handleConnection(conn net.Conn)
      ├─ request.Parse(conn)
      └─ write response.HardcodedOK()
```

The parser stays intentionally small: strict request-line split, header name/value split on `:`, trim optional whitespace, stop on blank line. This is enough for `curl`, `nc`, and hand-written test clients without dragging the project into full RFC edge-case hell too early.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/request/` | New | Request struct + parser |
| `internal/server/server.go` | Update | Read request before responding |
| `internal/server/server_test.go` | Update | End-to-end request parsing coverage |
| `README.md` | Update | Learning notes for parsing milestone |
| `sdd/milestone-2-request-parsing/` | New | Proposal/spec/design/tasks |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|-------------|
| Idle client blocks handler forever | Medium | Keep milestone tests focused on complete requests; add a seam for future read deadlines if needed |
| Over-engineering parser into mini `net/http` | High | Restrict scope to request line + headers only |
| Unread request bytes still cause reset on close | Medium | For milestone 2 only support header-only requests cleanly; body handling deferred |

## Rollback Plan

Remove `internal/request`, revert server handler to immediate hardcoded response, delete milestone-2 SDD docs and related tests.

## Dependencies

- Go standard library only: `bufio`, `fmt`, `io`, `strings`
- No `net/http`

## Success Criteria

- [ ] Valid `GET / HTTP/1.1` request parses into method/target/version + headers
- [ ] Server reads a real HTTP request from `net.Dial` client before responding
- [ ] Malformed request line returns parser error
- [ ] Malformed header line returns parser error
- [ ] `go test -race -v ./...` passes
- [ ] Existing milestone-1 behavior for valid requests still works
