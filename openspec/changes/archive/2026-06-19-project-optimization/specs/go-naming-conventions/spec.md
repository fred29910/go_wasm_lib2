## ADDED Requirements

### Requirement: Go acronym-safe naming
The system SHALL convert generated Go identifiers so common acronyms are rendered in uppercase, including `ID`, `URL`, `HTTP`, `JSON`, `API`, `UUID`, and `JWT`.

#### Scenario: Field name with id segment
- **WHEN** an OpenAPI property name contains the segment `id`
- **THEN** the generated Go field name uses `ID` instead of `Id`

#### Scenario: Field name with url segment
- **WHEN** an OpenAPI property name contains the segment `url`
- **THEN** the generated Go field name uses `URL` instead of `Url`

#### Scenario: Operation and type names
- **WHEN** an OpenAPI operation ID or schema name contains recognized acronym segments
- **THEN** the generated Go type or function name preserves acronym casing while remaining valid Go identifiers
