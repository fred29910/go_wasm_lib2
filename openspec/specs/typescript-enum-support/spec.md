# typescript-enum-support Specification

## Purpose
TBD - created by archiving change project-optimization. Update Purpose after archive.
## Requirements
### Requirement: TypeScript enum union generation
The system SHALL generate a named TypeScript union type for every OpenAPI schema property that defines an `enum` array with scalar values.

#### Scenario: String enum field
- **WHEN** an OpenAPI schema defines a string property with enum values `available`, `pending`, and `sold`
- **THEN** the generated TypeScript declares a union type that allows exactly those string values for the property

#### Scenario: Numeric enum field
- **WHEN** an OpenAPI schema defines a numeric property with enum values
- **THEN** the generated TypeScript union type uses the exact numeric enum values and rejects values outside the enum set

#### Scenario: Empty or absent enum
- **WHEN** a schema property does not define `enum` or defines an empty `enum`
- **THEN** the generated TypeScript type falls back to the normal scalar type without emitting an invalid union

