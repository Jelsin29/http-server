# Tasks: Milestone 3 - Serve Static Files with MIME Types

## Phase 1: SDD Artifacts

- [x] 1.1 Write proposal for milestone 3 scope
- [x] 1.2 Write specification for static file behavior
- [x] 1.3 Write technical design for static serving and response formatting

## Phase 2: Static Asset Loading

- [x] 2.1 Create `internal/static/static.go` with safe target resolution under `public/`
- [x] 2.2 Map `/` to `public/index.html`
- [x] 2.3 Reject `..` traversal attempts that escape `public/`
- [x] 2.4 Read file bytes from disk and surface missing-file errors cleanly
- [x] 2.5 Add explicit MIME lookup for `.html`, `.css`, `.js`, `.txt`, with binary fallback

## Phase 3: Response Formatting

- [x] 3.1 Replace `response.HardcodedOK()` with a general HTTP response builder
- [x] 3.2 Include `Content-Type` and `Content-Length` headers in file responses
- [x] 3.3 Support at least `200 OK` and `404 Not Found` response output

## Phase 4: Server Integration

- [x] 4.1 Update `internal/server/server.go` to load static assets after request parsing
- [x] 4.2 Return `200 OK` with file body for existing assets
- [x] 4.3 Return `404 Not Found` for missing assets
- [x] 4.4 Preserve malformed-request logging and early close behavior from milestone 2

## Phase 5: Sample Assets and Tests

- [x] 5.1 Create `public/index.html` for root requests
- [x] 5.2 Create `public/assets/app.css` for nested asset and MIME coverage
- [x] 5.3 Add table-driven tests for static path resolution and MIME selection
- [x] 5.4 Add response builder tests for headers, CRLF separator, and body bytes
- [x] 5.5 Update integration tests to assert static file hits and missing-file responses over real TCP

## Phase 6: Verification

- [x] 6.1 Run `go test -race -v ./...`
- [x] 6.2 Update README learning journal with milestone 3 notes
- [x] 6.3 Verify `git status` reflects only intended milestone 3 artifacts before implementation work continues
