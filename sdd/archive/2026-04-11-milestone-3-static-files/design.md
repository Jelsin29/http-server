# Design: Milestone 3 - Serve Static Files with MIME Types

## Technical Approach

Keep the request parser exactly where it is and insert a small static-file layer after parsing. `internal/server` will parse the request, resolve the target against `public/`, read the file bytes when present, and ask `internal/response` to format a full HTTP response with status line, headers, and body.

## Architecture Decisions

### Decision: Add a Focused `static` Package

**Choice**: Create `internal/static` for path resolution, safe file lookup, and MIME detection.
**Alternatives considered**: inline everything in `internal/server`, scatter MIME logic in `internal/response`.
**Rationale**: Path safety and MIME lookup are file-serving concerns, not socket concerns. Keeping them separate prevents `handleConnection` from turning into a pile of `filepath` edge cases.

### Decision: Use `public/` as Fixed Document Root

**Choice**: Serve only files under a repository-local `public/` directory.
**Alternatives considered**: configurable document root, serving from current working directory.
**Rationale**: Milestone 3 is about understanding the mechanics, not building deploy-time configuration. A fixed root keeps tests and behavior predictable.

### Decision: Explicit MIME Map

**Choice**: Maintain a tiny extension-to-content-type map for `.html`, `.css`, `.js`, and `.txt`, with `application/octet-stream` as fallback.
**Alternatives considered**: `mime.TypeByExtension`, broader MIME registry.
**Rationale**: The milestone explicitly wants MIME learning. A hand-written map makes the behavior visible and avoids pretending we need a complete content-negotiation system already.

### Decision: Introduce General Response Builder

**Choice**: Replace the single hardcoded response helper with a builder that can emit status line, headers, blank line, and body bytes.
**Alternatives considered**: keep string concatenation inline in the server, add separate helper per status code.
**Rationale**: Static files need `Content-Type`, `Content-Length`, and a body. That is the point where the current `HardcodedOK()` helper stops being honest.

## Data Flow

```text
Client bytes
   │
   ▼
server.handleConnection(conn)
   │
   ├─ request.Parse(conn)
   │    └─ return Request or error
   │
   ├─ static.Load("public", req.Target)
   │    ├─ normalize request target
   │    ├─ map / -> public/index.html
   │    ├─ reject escape outside public/
   │    ├─ read file bytes
   │    └─ return file payload + content type or not-found result
   │
   ├─ response.Build(status, headers, body)
   │
   └─ conn.Write(responseBytes)
```

## Proposed Data Model

```go
package static

type Asset struct {
    Body        []byte
    ContentType string
}

func Load(root string, target string) (*Asset, error)
```

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

The exact names can shift during implementation, but the contract stays simple: static lookup returns bytes plus MIME, and response formatting turns that into wire bytes.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/static/static.go` | Create | Safe target resolution, file loading, MIME lookup |
| `internal/static/static_test.go` | Create | Table-driven tests for path and MIME behavior |
| `internal/response/builder.go` | Update | Build arbitrary HTTP responses, not just hardcoded 200 |
| `internal/response/builder_test.go` | Update | Assert headers, blank line, and body formatting |
| `internal/server/server.go` | Update | Parse request, load static asset, write 200/404 responses |
| `internal/server/server_test.go` | Update | Real TCP tests for existing and missing files |
| `public/index.html` | Create | Default root document |
| `public/assets/app.css` | Create | Sample nested asset for MIME tests |
| `README.md` | Update | Learning notes for milestone 3 |

## Interfaces / Contracts

```go
package static

type Asset struct {
    Body        []byte
    ContentType string
}

func Load(root string, target string) (*Asset, error)
func ContentTypeFor(name string) string
```

`Load` returns a descriptive error when the target is invalid or the file cannot be read. The server will translate `os.ErrNotExist` into `404 Not Found` and treat other file errors as internal failures in milestone 4. For milestone 3, the main behavior target is clean hit/miss handling.

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

The builder is responsible for:

- writing `HTTP/1.1 <status> <reason>\r\n`
- writing headers in `Key: Value\r\n` format
- writing the required blank line
- appending body bytes unchanged

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Root and nested path mapping | table-driven `static.Load` cases |
| Unit | Traversal rejection | targets like `/../secret.txt` and `/assets/../../secret.txt` |
| Unit | MIME selection | direct `ContentTypeFor` tests |
| Unit | Response wire format | assert status line, headers, CRLF separator, and body |
| Integration | Existing file over TCP | `net.Dial`, send `GET /`, assert 200, headers, body |
| Integration | Missing file over TCP | `net.Dial`, send `GET /missing.txt`, assert 404 |
| Integration | Malformed request behavior | keep milestone 2 coverage to prove parser gate still holds |

## Open Questions

1. Should directory requests other than `/` auto-map to `index.html`?
   - **Decision for now**: no, only `/` gets special treatment.
2. Should unsupported methods like `POST` return `405 Method Not Allowed` now?
   - **Decision for now**: no, method-specific responses can wait for milestone 4 once error handling is broader.
