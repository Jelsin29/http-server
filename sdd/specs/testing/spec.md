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
