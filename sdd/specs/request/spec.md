# Specification: Request Parsing

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
