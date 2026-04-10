# Tasks: Milestone 2 - Parse HTTP Request Line and Headers

## Phase 1: SDD Artifacts

- [x] 1.1 Write proposal for milestone 2 scope
- [x] 1.2 Write specification for request parsing behavior
- [x] 1.3 Write technical design for parser + server integration

## Phase 2: Request Parser

- [x] 2.1 Create `internal/request/request.go` with minimal `Request` struct
- [x] 2.2 Create `internal/request/parser.go` with `Parse(io.Reader) (*Request, error)`
- [x] 2.3 Parse request line into method, target, version
- [x] 2.4 Parse headers until blank line, trimming optional whitespace
- [x] 2.5 Return descriptive errors for malformed request line or header

## Phase 3: Server Integration

- [x] 3.1 Update `internal/server/server.go` to parse request before responding
- [x] 3.2 Log parse failures with connection context
- [x] 3.3 Log response write failures with connection context

## Phase 4: Testing

- [x] 4.1 Add table-driven parser tests for valid requests
- [x] 4.2 Add parser tests for malformed request line and malformed header
- [x] 4.3 Update integration tests to send real HTTP request bytes over TCP
- [x] 4.4 Add integration test ensuring malformed request closes without 200 OK

## Phase 5: Verification

- [x] 5.1 Run `go test -race -v ./...`
- [ ] 5.2 Update README learning journal with milestone 2 notes
- [ ] 5.3 Verify `git status` is clean after atomic commits
