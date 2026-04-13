# http-server — learning journal

started this because i had no idea what happens when you type `curl http://localhost:8080` into a terminal. like, what's actually happening at the TCP level? frameworks hide all of this stuff.

so i'm building an HTTP/1.1 server from scratch using only `net` package — no `net/http`, no frameworks. raw socket programming.

## what i built so far

### milestone 1

a minimal TCP server that accepts connections and responds with "HTTP/1.1 200 OK\r\n\r\n". that's it. no request parsing, no routing, no concurrency. just:

1.Listen on a port (`net.Listen("tcp", ":8080")`)
2.Accept connections in a loop (`ln.Accept()`)
3.Write hardcoded response and close (`conn.Write()` + `conn.Close()`)

### milestone 2

now it actually reads an HTTP request before answering. not a full web server yet, but enough to parse:

- request line (`GET /hello HTTP/1.1`)
- headers until the blank line
- malformed header lines vs valid `name: value` pairs

the server still answers with the same dumb hardcoded `200 OK`, but at least now it understands the shape of what the client sent.

### milestone 3

now it serves actual files from `public/` instead of always lying with the same empty response. right now it can:

- map `/` to `public/index.html`
- serve nested assets like `/assets/app.css`
- set `Content-Type` from a tiny explicit MIME map
- set `Content-Length` correctly
- return `404 Not Found` with a small body when the file doesn't exist

### milestone 4

this was the "okay, now stop pretending this is a real server" milestone. i changed two big things:

- each accepted connection now gets its own goroutine, so one blocked request doesn't freeze the listener
- malformed requests return `400 Bad Request` instead of just getting the socket closed in silence
- missing files and traversal attempts still return `404 Not Found`
- unexpected server-side failures now return a generic `500 Internal Server Error`
- tests prove the second client can finish while the first one is deliberately blocked mid-request

## what broke me

**the `:0` port thing** — i wrote tests using `:0` thinking "OS will assign a port, perfect for testing." but then the test tries to connect to `:0` which... doesn't work. you can't connect to `:0`. you have to know the actual port the OS assigned.

the fix was to use fixed high ports (18081, 18082) for tests. feels hacky but whatever, it works. i'll figure out the `:0` thing later.

**defer `conn.Close()` placement** — originally i had the `defer` inside a goroutine (because i was thinking about concurrency). but milestone 1 is sequential, so the `defer` should be right after `Accept()`, before handling the connection. moved it, tests passed.

**HTTP is just lines until it suddenly isn't** — parsing the request line felt easy for about five minutes. then headers show up and you realize the protocol is basically: read lines, split carefully, stop exactly on the blank line, don't get lazy with whitespace around `:` unless you want weird bugs. it's amazing how fast "just parse some text" turns into protocol design.

**relative paths are a trap** — i had the static server looking up `public/` as a relative path and thought i was done because the direct handler tests passed. integration said otherwise. turns out the process working directory matters, and tying file resolution to cwd is one of those bugs that looks harmless until it wastes half an hour. fixed it by resolving the document root from the server package location instead of assuming where the process started.

**"concurrency" is not just sprinkling `go` everywhere** — i knew milestone 4 would be `go handleConnection(conn)`, but that only sounds easy if you ignore ownership. the handler has to own `defer conn.Close()`, the accept loop has to stay dumb, and the test has to prove real overlap instead of just sleeping and hoping. once i added a blocking seam around static loading, the concept finally clicked.

## what actually broke (real bugs)

- test on `:0` failed with "connection refused"— first tried to bind to `:0` and connect to `:0`, which doesn't work. you can listen on `:0` (OS assigns port) but you can't connect to it. had to use fixed ports for tests.
- spent way too long wondering why the accept loop wasn't returning... it's not supposed to return, it's an infinite loop (`for { ... }`). tests run it in a goroutine.
- forgot to add `defer conn.Close()` inside the connection handler initially, was leaking connections. remembered goroutines are milestone 4, this is milestone 1, moved it to sequential handling.
- server originally answered before reading anything, which meant tests weren't proving real HTTP behavior. milestone 2 fixed that by forcing the handler to parse request line + headers first.
- i was also ignoring write errors like a clown. added tests with a fake `net.Conn` so i could prove parse failures and write failures still close the connection.
- static file serving initially returned `404` in integration even for `/` because i resolved `public/` relative to cwd. the unit-ish handler test accidentally hid the bug because it reused the same broken resolution path. the real TCP test caught it.
- milestone 3 malformed requests just got disconnected with no HTTP response. that was lazy. milestone 4 fixed it with an explicit `400 Bad Request` body so the wire behavior actually says what happened.
- my first thought for the concurrency test was `time.Sleep(100 * time.Millisecond)` and hope the second request wins. that's garbage. i replaced it with a blocking loader seam so the test proves overlap deterministically.

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
# HTTP/1.1 404 Not Found
# not found

curl http://localhost:8080/assets/app.css
# HTTP/1.1 200 OK
# Content-Type: text/css; charset=utf-8
```

## what's next

if i kept going, the next pain would be request bodies and routing. right now this server can parse headers, serve files, and fail a lot more honestly than before, but it still has no idea how to read `Content-Length` bytes from a POST body or dispatch different handlers by method/path.

## what i'd do differently

i should've started milestone 1 with a more realistic integration test. reading a response without sending a request proved the TCP write, sure, but it dodged the actual HTTP part. milestone 2 corrected that.

also, body handling is still missing. header-only requests are fine. the moment someone sends `Content-Length: 20` and actually includes a body, i'm back in the weeds.

also, my test harness still starts real listeners on fixed high ports and leaves them running for the duration of the test process. it works, but it's definitely a learning-project scar, not a polished harness.

## references

- [RFC 7230 - HTTP/1.1 Message Syntax and Routing](https://datatracker.ietf.org/doc/html/rfc7230)
- [Go net package docs](https://pkg.go.dev/net)
- [Russ Cox on TCP sockets](https://swtch.com/~rsc/regexp/) (this is about regex but he explains networking concepts really well)
