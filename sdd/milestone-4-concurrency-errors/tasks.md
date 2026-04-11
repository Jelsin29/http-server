# Tasks: Milestone 4 - Concurrency and Error Responses

## Phase 1: SDD Artifacts

- [x] 1.1 Write proposal for milestone 4 scope
- [x] 1.2 Write specification for concurrency and explicit error responses
- [x] 1.3 Write technical design for goroutine handling and error mapping

## Phase 2: Error Response Plumbing

- [ ] 2.1 Add a shared helper for small text HTTP responses with correct `Content-Type` and `Content-Length`
- [ ] 2.2 Return `400 Bad Request` when `request.Parse` fails instead of closing with no response
- [ ] 2.3 Keep `404 Not Found` for missing files and traversal attempts
- [ ] 2.4 Return `500 Internal Server Error` for unexpected static or response-building failures
- [ ] 2.5 Ensure client-facing `500` bodies stay generic while server logs keep detailed context

## Phase 3: Concurrency

- [ ] 3.1 Update `Start` to launch `go handleConnection(conn)` after each successful `Accept`
- [ ] 3.2 Keep connection ownership local to the handler with `defer conn.Close()`
- [ ] 3.3 Confirm accept-loop errors are logged and do not stop future accepts

## Phase 4: Test Coverage

- [ ] 4.1 Add handler/integration coverage proving malformed requests receive `400 Bad Request`
- [ ] 4.2 Add deterministic test coverage for concurrent handling without relying on raw sleeps alone
- [ ] 4.3 Add coverage proving missing files and traversal attempts still return `404 Not Found`
- [ ] 4.4 Add injected-failure coverage proving unexpected server errors return `500 Internal Server Error`
- [ ] 4.5 Run `go test -race -v ./...` with the new concurrency tests

## Phase 5: Journal and Verification

- [ ] 5.1 Update `README.md` with milestone 4 notes, mistakes, and what changed mentally
- [ ] 5.2 Verify `git status` reflects only intended milestone 4 artifacts and implementation changes
