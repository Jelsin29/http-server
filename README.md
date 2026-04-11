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

## what broke me

**the `:0` port thing** — i wrote tests using `:0` thinking "OS will assign a port, perfect for testing." but then the test tries to connect to `:0` which... doesn't work. you can't connect to `:0`. you have to know the actual port the OS assigned.

the fix was to use fixed high ports (18081, 18082) for tests. feels hacky but whatever, it works. i'll figure out the `:0` thing later.

**defer `conn.Close()` placement** — originally i had the `defer` inside a goroutine (because i was thinking about concurrency). but milestone 1 is sequential, so the `defer` should be right after `Accept()`, before handling the connection. moved it, tests passed.

**HTTP is just lines until it suddenly isn't** — parsing the request line felt easy for about five minutes. then headers show up and you realize the protocol is basically: read lines, split carefully, stop exactly on the blank line, don't get lazy with whitespace around `:` unless you want weird bugs. it's amazing how fast "just parse some text" turns into protocol design.

**relative paths are a trap** — i had the static server looking up `public/` as a relative path and thought i was done because the direct handler tests passed. integration said otherwise. turns out the process working directory matters, and tying file resolution to cwd is one of those bugs that looks harmless until it wastes half an hour. fixed it by resolving the document root from the server package location instead of assuming where the process started.

## what actually broke (real bugs)

- test on `:0` failed with "connection refused"— first tried to bind to `:0` and connect to `:0`, which doesn't work. you can listen on `:0` (OS assigns port) but you can't connect to it. had to use fixed ports for tests.
- spent way too long wondering why the accept loop wasn't returning... it's not supposed to return, it's an infinite loop (`for { ... }`). tests run it in a goroutine.
- forgot to add `defer conn.Close()` inside the connection handler initially, was leaking connections. remembered goroutines are milestone 4, this is milestone 1, moved it to sequential handling.
- server originally answered before reading anything, which meant tests weren't proving real HTTP behavior. milestone 2 fixed that by forcing the handler to parse request line + headers first.
- i was also ignoring write errors like a clown. added tests with a fake `net.Conn` so i could prove parse failures and write failures still close the connection.
- static file serving initially returned `404` in integration even for `/` because i resolved `public/` relative to cwd. the unit-ish handler test accidentally hid the bug because it reused the same broken resolution path. the real TCP test caught it.

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

milestone 4 is concurrency (goroutine per connection) and broader error responses (`400`, `404`, maybe `500` when file reads blow up for real). right now the server is still single-connection-at-a-time, which is fine for learning, but absolutely not okay for a real workload.

## what i'd do differently

i should've started milestone 1 with a more realistic integration test. reading a response without sending a request proved the TCP write, sure, but it dodged the actual HTTP part. milestone 2 corrected that.

also, body handling is still missing. header-only requests are fine. the moment someone sends `Content-Length: 20` and actually includes a body, i'm back in the weeds.

## references

- [RFC 7230 - HTTP/1.1 Message Syntax and Routing](https://datatracker.ietf.org/doc/html/rfc7230)
- [Go net package docs](https://pkg.go.dev/net)
- [Russ Cox on TCP sockets](https://swtch.com/~rsc/regexp/) (this is about regex but he explains networking concepts really well)
