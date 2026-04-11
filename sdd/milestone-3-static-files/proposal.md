# Proposal: Milestone 3 - Serve Static Files with MIME Types

## Intent

Turn the parsed request target into something useful by mapping `GET` paths to files on disk and returning real HTTP responses with `Content-Type` and `Content-Length`. Milestone 2 proved request parsing; milestone 3 proves basic server-side routing and response composition.

## Scope

### In Scope
- Serve files from a dedicated `public/` directory for `GET` requests
- Map `/` to a default `index.html`
- Infer a small MIME set for `.html`, `.css`, `.js`, `.txt`, and fallback binary content
- Return `200 OK` with file bytes for hits and `404 Not Found` for missing files
- Add tests for existing files, missing files, and MIME selection

### Out of Scope
- Directory listings or auto-index beyond `/ -> index.html`
- Range requests, caching headers, keep-alive, or chunked bodies
- Concurrent handlers and richer error taxonomy beyond milestone needs

## Approach

Add a tiny static-file layer between parsed request and socket write. `internal/server/server.go` will resolve the request target against `public/`, reject path escape attempts, then delegate to `internal/response` for full HTTP response bytes. Keep MIME detection as an explicit map instead of pulling in heavy helpers.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/server/server.go` | Modified | Route parsed `GET` targets to file lookup |
| `internal/response/` | Modified | Build status line, headers, and body |
| `internal/static/` | New | Path resolution + MIME lookup |
| `internal/server/server_test.go` | Modified | End-to-end static file coverage |
| `public/` | New | Sample assets served by the server |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Path traversal via `..` segments | Medium | Clean paths and enforce `public/` prefix |
| Response builder grows messy fast | Medium | Keep file loading separate from response formatting |
| MIME coverage too broad too early | Low | Support only milestone file types |

## Rollback Plan

Remove `internal/static/` and `public/`, revert server flow to hardcoded 200 after parse, and drop milestone-3 tests/docs.

## Dependencies

- Go standard library only: `os`, `path/filepath`, `strings`, `bytes`

## Success Criteria

- [ ] `GET /` returns `public/index.html` with `200 OK`
- [ ] Existing file responses include correct `Content-Type` and `Content-Length`
- [ ] Missing files return `404 Not Found`
- [ ] Traversal attempts do not escape `public/`
- [ ] `go test -race -v ./...` passes
