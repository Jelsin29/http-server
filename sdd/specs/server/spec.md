# Specification: Server

## Domain: server

### Requirement: Read Request Before Responding

The server MUST read and parse a request before sending the hardcoded response.

#### Scenario: Server responds after parsing valid request

- GIVEN the server is running
- WHEN a client sends a complete HTTP request line and headers
- THEN the server parses the request successfully
- AND the client receives `HTTP/1.1 200 OK\r\n\r\n`

#### Scenario: Server closes malformed request connection

- GIVEN the server accepts a malformed request
- WHEN request parsing fails
- THEN the server closes the connection
- AND the error is logged with context

### Requirement: Handle Accepted Connections Concurrently

The server MUST keep the accept loop responsive by handling each accepted connection in its own goroutine.

#### Scenario: Slow connection does not block a second client

- GIVEN the server accepts one connection that stalls during request handling
- WHEN a second client connects and sends a complete request
- THEN the second client still receives its full HTTP response
- AND the accept loop keeps accepting connections

#### Scenario: Accept errors do not stop the listener

- GIVEN the listener encounters an `Accept` error
- WHEN the server logs the failure
- THEN the main accept loop continues
- AND later connections can still be handled

### Requirement: Return 400 for Malformed HTTP Requests

The server MUST return `HTTP/1.1 400 Bad Request` when request parsing fails after a connection is accepted.

#### Scenario: Malformed header returns 400 response

- GIVEN a client sends a request with an invalid header line
- WHEN request parsing fails
- THEN the response status line is `HTTP/1.1 400 Bad Request`
- AND the response body explains the request is bad

#### Scenario: Malformed request still closes cleanly

- GIVEN a client sends malformed request bytes
- WHEN the server returns `400 Bad Request`
- THEN the server closes the connection after writing the response

### Requirement: Preserve Existing 404 Behavior

The server MUST keep returning `404 Not Found` for missing files and blocked traversal attempts.

#### Scenario: Missing file still returns 404

- GIVEN the request target resolves inside `public/` but the file does not exist
- WHEN the server builds the response
- THEN the response status line is `HTTP/1.1 404 Not Found`
- AND the response body stays generic

#### Scenario: Traversal attempt still returns 404

- GIVEN the request target would escape `public/`
- WHEN static resolution fails with traversal protection
- THEN the response status line is `HTTP/1.1 404 Not Found`
- AND no file outside `public/` is read

### Requirement: Return 500 for Unexpected Server Failures

The server MUST return `HTTP/1.1 500 Internal Server Error` when response generation fails for reasons other than request syntax or missing/traversal file targets.

#### Scenario: Static read failure returns 500

- GIVEN the request is syntactically valid
- AND static asset loading fails for an unexpected filesystem reason
- WHEN the server maps that error to an HTTP response
- THEN the response status line is `HTTP/1.1 500 Internal Server Error`
- AND the body does not expose internal error details

#### Scenario: Response write setup failure returns 500 body

- GIVEN request parsing succeeds
- AND response construction hits an unexpected server-side failure before `conn.Write`
- WHEN the failure is handled
- THEN the client receives `HTTP/1.1 500 Internal Server Error`
- AND the server logs contextual details separately
