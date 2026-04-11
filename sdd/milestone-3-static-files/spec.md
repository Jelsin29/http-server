# Specification: Milestone 3 - Serve Static Files with MIME Types

## Purpose

Turn parsed `GET` requests into real file responses so the server can serve a tiny website from disk instead of always returning the same empty `200 OK`.

## Domain: static

### Requirement: Resolve Request Targets Inside Public Directory

The system MUST map request targets to files under a dedicated `public/` directory without allowing path traversal outside that root.

#### Scenario: Map root path to index file

- GIVEN the request target is `/`
- WHEN the server resolves the file path
- THEN it maps to `public/index.html`

#### Scenario: Resolve nested asset path

- GIVEN the request target is `/assets/app.css`
- WHEN the server resolves the file path
- THEN it maps to `public/assets/app.css`

#### Scenario: Reject traversal attempt

- GIVEN the request target contains `..` segments that would escape `public/`
- WHEN the server resolves the file path
- THEN resolution fails
- AND the server does not read files outside `public/`

### Requirement: Infer Content Type from File Extension

The system MUST choose a `Content-Type` header from a small explicit MIME map.

#### Scenario: Return HTML content type

- GIVEN the resolved file ends with `.html`
- WHEN the response is built
- THEN `Content-Type` is `text/html; charset=utf-8`

#### Scenario: Return CSS content type

- GIVEN the resolved file ends with `.css`
- WHEN the response is built
- THEN `Content-Type` is `text/css; charset=utf-8`

#### Scenario: Return JavaScript content type

- GIVEN the resolved file ends with `.js`
- WHEN the response is built
- THEN `Content-Type` is `application/javascript; charset=utf-8`

#### Scenario: Return plain text content type

- GIVEN the resolved file ends with `.txt`
- WHEN the response is built
- THEN `Content-Type` is `text/plain; charset=utf-8`

#### Scenario: Fallback for unknown extension

- GIVEN the resolved file has an unsupported or missing extension
- WHEN the response is built
- THEN `Content-Type` is `application/octet-stream`

## Domain: server

### Requirement: Serve Existing Static Files

The server MUST return file contents for `GET` requests that resolve to existing files under `public/`.

#### Scenario: Existing file returns 200 response

- GIVEN the server is running and `public/index.html` exists
- WHEN a client sends `GET / HTTP/1.1`
- THEN the response status line is `HTTP/1.1 200 OK`
- AND the response body is the file contents
- AND the response includes `Content-Type` and `Content-Length`

#### Scenario: Existing nested asset returns 200 response

- GIVEN the server is running and `public/assets/app.css` exists
- WHEN a client sends `GET /assets/app.css HTTP/1.1`
- THEN the response status line is `HTTP/1.1 200 OK`
- AND the response body is the CSS file contents
- AND `Content-Type` matches the file extension

### Requirement: Return Not Found for Missing Files

The server MUST return a `404 Not Found` response when a `GET` target does not resolve to an existing file.

#### Scenario: Missing file returns 404

- GIVEN the server is running and `public/missing.txt` does not exist
- WHEN a client sends `GET /missing.txt HTTP/1.1`
- THEN the response status line is `HTTP/1.1 404 Not Found`
- AND the response body explains the file was not found

### Requirement: Keep Parse Failures from Sending Static Responses

The server MUST keep malformed request handling from milestone 2.

#### Scenario: Malformed request still closes connection

- GIVEN the server accepts a malformed request
- WHEN request parsing fails before file resolution
- THEN the connection closes
- AND no static file response is written

## Domain: testing

### Requirement: Cover Static File Serving Behavior

The system MUST have unit and integration tests for file resolution, MIME selection, and full TCP responses.

#### Scenario: Unit tests cover path resolution and MIME lookup

- GIVEN table-driven tests for the static package
- WHEN the test suite runs
- THEN it covers root mapping
- AND nested asset mapping
- AND traversal rejection
- AND MIME selection for supported and fallback extensions

#### Scenario: Integration tests cover hit and miss responses

- GIVEN the server is started on a test port with sample files in `public/`
- WHEN a real TCP client requests an existing file and a missing file
- THEN the suite verifies both `200 OK` and `404 Not Found` responses
- AND it verifies `Content-Length` matches the body bytes
