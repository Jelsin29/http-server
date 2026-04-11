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
