## ADDED Requirements

### Requirement: Generated request validation
The system SHALL generate Go validation methods for request structs when validation generation is enabled.

#### Scenario: Required fields
- **WHEN** a request struct contains fields marked `required` in the OpenAPI schema
- **THEN** the generated `Validate() error` method returns an error when a required field has a zero value

#### Scenario: Enum values
- **WHEN** a request field has an OpenAPI enum constraint
- **THEN** the generated validation method rejects values that are not in the allowed enum set

#### Scenario: Format constraints
- **WHEN** a request field has an OpenAPI format constraint such as `email`, `uuid`, or `date-time`
- **THEN** the generated validation method checks the field value against the corresponding format rule

#### Scenario: Validation disabled
- **WHEN** validation generation is disabled by CLI configuration
- **THEN** the generated request structs do not include validation methods
