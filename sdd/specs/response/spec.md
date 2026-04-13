# Specification: Response

## Domain: response

### Requirement: Provide Consistent Error Message Formatting

The response layer MUST format success and error messages through one shared contract.

#### Scenario: 400 response includes text body and headers

- GIVEN the server emits a bad-request response
- WHEN the response bytes are built
- THEN the message includes `Content-Type: text/plain; charset=utf-8`
- AND `Content-Length` matches the body bytes

#### Scenario: 500 response uses generic body

- GIVEN the server emits an internal-server-error response
- WHEN the response bytes are built
- THEN the body is a generic message like `internal server error`
- AND no raw Go error text appears in the client response
