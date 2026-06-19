## ADDED Requirements

### Requirement: Enhanced OpenAPI code generation model
The system SHALL extend the internal OpenAPI code generation model so downstream templates can access enum values, all documented response schemas, and validation metadata.

#### Scenario: Enum metadata captured
- **WHEN** the parser reads an OpenAPI schema containing enum values
- **THEN** the internal model preserves those enum values for both Go and TypeScript generation

#### Scenario: All response schemas captured
- **WHEN** an OpenAPI operation documents multiple response status codes
- **THEN** the internal model stores each response code and schema for template rendering

#### Scenario: Validation metadata captured
- **WHEN** an OpenAPI schema defines required fields, enum constraints, or format constraints
- **THEN** the internal model exposes that metadata to the validation generation path
