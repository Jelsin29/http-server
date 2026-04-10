# Specification: Parse HTTP Request Line and Headers

## Purpose

Turn raw TCP bytes into a minimal HTTP request representation so the server understands what the client sent before writing a response.

## Domain: request

### Requirement: Parse Request Line

The system MUST parse the HTTP request line into method, target, and version.

#### Scenario: Parse a valid origin-form request line

- GIVEN input `GET /hello HTTP/1.1\r\n`
- WHEN the parser reads the first line
- THEN method is `GET`
- AND target is `/hello`
- AND version is `HTTP/1.1`

#### Scenario: Reject malformed request line

- GIVEN input with fewer or more than three request-line parts
- WHEN the parser reads the first line
- THEN parsing fails with a descriptive error

### Requirement: Parse Header Fields

The system MUST parse header fields until the blank line that terminates the header section.

#### Scenario: Parse multiple headers

- GIVEN a valid request line followed by headers and a blank line
- WHEN the parser reads the header section
- THEN each `name: value` pair is stored in the request headers
- AND surrounding optional whitespace is trimmed from header values

#### Scenario: Reject malformed header line

- GIVEN a header line without a colon separator
- WHEN the parser reads the line
- THEN parsing fails with a descriptive error

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

## Domain: testing

### Requirement: Meaningful Parser Coverage

The system MUST have tests for parser success and parser failure cases.

#### Scenario: Parser unit tests cover request line and headers

- GIVEN table-driven parser tests
- WHEN the test suite runs
- THEN it covers valid request parsing
- AND malformed request line handling
- AND malformed header handling

#### Scenario: Integration test uses a real TCP client

- GIVEN the server is started on a test port
- WHEN a client sends a minimal HTTP/1.1 request over `net.Dial`
- THEN the client receives the hardcoded response
- AND the connection closes after the response
