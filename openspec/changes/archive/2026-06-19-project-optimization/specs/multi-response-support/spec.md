## ADDED Requirements

### Requirement: Typed response generation for documented status codes
The system SHALL generate typed response models for every documented HTTP response code in an OpenAPI operation, including 4xx and 5xx responses with schemas.

#### Scenario: Multiple documented responses
- **WHEN** an operation documents response schemas for `200`, `400`, and `500`
- **THEN** the generated Go and TypeScript SDK exposes distinct typed response models for each documented status code

#### Scenario: Primary success response
- **WHEN** an operation documents a 2xx response schema
- **THEN** the generated SDK treats that response as the primary success response type

#### Scenario: Undocumented response fallback
- **WHEN** an operation returns a status code without a documented schema
- **THEN** the generated SDK preserves the existing fallback behavior using a generic response type
