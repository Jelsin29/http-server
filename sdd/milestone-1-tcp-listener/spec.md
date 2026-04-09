# Specification: TCP Server with Hardcoded Response

## Purpose

Minimal TCP listener that accepts connections and responds with hardcoded HTTP/1.1 200 OK. Foundation for HTTP server teaching socket lifecycle before parsing complexity.

## Domain: server

### Requirement: Accept TCP Connections

The system MUST accept incoming TCP connections on a configurable port (default: 8080).

#### Scenario: Server starts and accepts connection

- GIVEN server is not running
- WHEN server starts on port 8080
- THEN server listens for incoming connections
- AND accepts the first connection

#### Scenario: Server handles port already in use

- GIVEN port 8080 is already bound by another process
- WHEN server attempts to start
- THEN server returns clear error message containing "address already in use"
- AND exits with non-zero status

### Requirement: Clean Connection Lifecycle

The system MUST close every accepted connection after writing the response.

#### Scenario: Connection closes after response

- GIVEN server is accepting connections
- WHEN a client connects
- THEN server writes HTTP response
- AND server closes the connection

#### Scenario: Connection cleanup on write failure

- GIVEN server accepts connection
- WHEN write operation fails
- THEN connection MUST still be closed via defer
- AND error is logged

## Domain: response

### Requirement: Hardcoded HTTP Response

The system MUST respond with exactly "HTTP/1.1 200 OK\r\n\r\n" for every connection, regardless of request content.

#### Scenario: Client receives correct HTTP response

- GIVEN server is running
- WHEN client connects and sends any data (or nothing)
- THEN client receives "HTTP/1.1 200 OK\r\n\r\n"
- AND connection is closed by server

#### Scenario: Response format correctness

- GIVEN server sends response
- WHEN response is captured
- THEN response starts with "HTTP/1.1 200 OK"
- AND ends with double CRLF ("\r\n\r\n")
- AND has no body

### Requirement: Response Fits in Single TCP Write

The system SHOULD send response in a single `conn.Write()` call without buffering.

#### Scenario: Minimal response sent atomically

- GIVEN server accepts connection
- WHEN response is written
- THEN all 18 bytes sent in single Write call
- AND no partial write handling required

## Domain: testing

### Requirement: Integration Test Coverage

The system MUST have integration tests verifying end-to-end behavior using `net.Dial`.

#### Scenario: Test verifies connection and response

- GIVEN server is started on test port
- WHEN test client connects using net.Dial
- THEN test receives "HTTP/1.1 200 OK\r\n\r\n"
- AND test verifies connection closes properly

#### Scenario: Test covers port binding failure

- GIVEN port is in use
- WHEN server attempts to bind
- THEN test verifies error returned
- AND test verifies error message contains "address already in use"