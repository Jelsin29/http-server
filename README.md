# http-server — learning journal

started this because i had no idea what happens when you type `curl http://localhost:8080` into a terminal. like, what's actually happening at the TCP level? frameworks hide all ofthis stuff.

so i'm building an HTTP/1.1 server from scratch using only `net` package — no `net/http`, no frameworks. raw socket programming.

## what i built (milestone 1)

a minimal TCP server that accepts connections and responds with "HTTP/1.1 200 OK\r\n\r\n". that's it. no request parsing, no routing, no concurrency. just:

1.Listen on a port (`net.Listen("tcp", ":8080")`)
2.Accept connections in a loop (`ln.Accept()`)
3.Write hardcoded response and close (`conn.Write()` + `conn.Close()`)

## what broke me

**the `:0` port thing** — i wrote tests using`:0`thinking "OS will assign a port, perfect for testing." but then the test tries to connect to `:0`which... doesn't work. you can't connect to `:0`. you have to know the actual port the OS assigned.

the fix was to use fixed high ports (18081, 18082) for tests. feels hacky but whatever, it works. i'll figure out the `:0` thing later.

**defer `conn.Close()` placement** — originally i had the `defer` inside a goroutine (because i was thinking about concurrency). but milestone 1 is sequential, so the `defer` should be right after `Accept()`, before handling the connection. moved it, tests passed.

## what actually broke (real bugs)

- test on `:0` failed with "connection refused"— fixed by using fixed ports
- spent way too long wondering why the accept loop wasn't returning... it's not supposed to return, it's an infinite loop (`for { ... }`). tests run it in a goroutine.

## usage

```bash
make          # builds to bin/http-server
make run      # runs the server on :8080
make test     # runs tests with race detector
make lint     # runs go vet + golangci-lint
```

test it:

```bash
make run &
curl http://localhost:8080
# HTTP/1.1 200 OK
#
echo "hello" | nc localhost 8080
# HTTP/1.1 200 OK
#
```

## what's next

milestone 2 is parsing the HTTP request line and headers. right now the server just blindly responds "200 OK"without even looking at what the client sent. that's... not how HTTP works.

milestone 3 is serving static files with MIME types. milestone 4 is concurrency (goroutine per connection) and proper error responses (404, 400, 500).

## what i'd do differently

i should've started with the test first. i wrote the response package function-first, then realized i had no way to test it. TDD exists for a reason.

also, the test for port-in-use is kinda weak — it just checks if error contains "bind" or "address already in use". could be more robust, but it's fine for milestone 1.

## references

- [RFC 7230 - HTTP/1.1 Message Syntax and Routing](https://datatracker.ietf.org/doc/html/rfc7230)
- [Go net package docs](https://pkg.go.dev/net)
- [Russ Cox on TCP sockets](https://swtch.com/~rsc/regexp/) (this is about regex but he explains networking concepts really well)