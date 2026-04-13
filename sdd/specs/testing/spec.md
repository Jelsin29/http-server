# Specification: Testing

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

### Requirement: Cover Concurrency and Error Mapping

The system MUST have tests that prove concurrent handling and status-specific error responses.

#### Scenario: Integration test proves concurrent request handling

- GIVEN a test double or synchronization point that blocks one connection mid-request
- WHEN a second real client request is sent over TCP
- THEN the second client completes before the first one is released

#### Scenario: Tests cover 400, 404, and 500 branches

- GIVEN unit and integration tests for the server response path
- WHEN the suite runs
- THEN it verifies malformed requests return `400 Bad Request`
- AND missing files and traversal attempts return `404 Not Found`
- AND injected internal failures return `500 Internal Server Error`
