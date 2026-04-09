# Tasks: Milestone 1 - TCP Listener with Hardcoded HTTP Response

## Phase 1: Infrastructure

- [x] 1.1 Create `go.mod` with module `github.com/jelsin/http-server` and Go 1.21+
- [x] 1.2 Create `Makefile` with targets: `build`, `test`, `lint`, `run`, `clean` (follow go-systems skill template)

## Phase 2: Core Implementation

- [x] 2.1 Create `internal/response/builder.go` with `HardcodedOK() string` returning "HTTP/1.1 200 OK\r\n\r\n"
- [x] 2.2 Create `internal/server/server.go` with `Start(addr string) error` — TCP listener on `net.Listen("tcp", addr)`
- [x] 2.3 Add accept loop to `server.go`: infinite `ln.Accept()`, log errors and continue, defer `conn.Close()` immediately after accept
- [x] 2.4 Add connection handler to `server.go`: write `response.HardcodedOK()` to connection, return (connection closes via defer)
- [x] 2.5 Create `cmd/main.go` calling `server.Start(":8080")`, handle error with `log.Fatal`
- [x] 2.6 Add error handling: fatal on Listen failure (port in use), log on Accept failure (transient)

## Phase 3: Testing

- [x] 3.1 Create `internal/server/server_test.go` with table-driven test: connect via `net.Dial`, read response, verify exact "HTTP/1.1 200 OK\r\n\r\n"
- [x] 3.2 Add test case: port:0 (OS assigns arbitrary) to verify server works on any port
- [x] 3.3 Add test case: server handles connection cleanup (no leaks) — verify connection closes after response
- [x] 3.4 Add test case: `Start(":8080")` fails with clear error when port already in use

## Phase 4: Verification

- [x] 4.1 Run `make test` — all tests pass
- [x] 4.2 Run `make lint` (go vet) — no errors
- [x] 4.3 Test manually: `make run`, `curl http://localhost:8080`, verify "200 OK" response
- [x] 4.4 Test manually: `nc localhost 8080`, send arbitrary bytes, verify "HTTP/1.1 200 OK\r\n\r\n" response
- [x] 4.5 Verify `git status` shows clean (all changes committed per commit discipline)