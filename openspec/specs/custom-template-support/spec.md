# custom-template-support Specification

## Purpose
TBD - created by archiving change project-optimization. Update Purpose after archive.
## Requirements
### Requirement: Custom template file support
The system SHALL allow users to provide custom Go and TypeScript template files for SDK generation.

#### Scenario: Custom Go template provided
- **WHEN** the user supplies a valid `--go-template` file path
- **THEN** the generator renders Go SDK output using that template instead of the embedded default

#### Scenario: Custom TypeScript template provided
- **WHEN** the user supplies a valid `--ts-template` file path
- **THEN** the generator renders TypeScript SDK output using that template instead of the embedded default

#### Scenario: Custom template missing
- **WHEN** the user supplies a template path that does not exist or cannot be read
- **THEN** generation fails with an actionable error that names the missing template path

#### Scenario: No custom template provided
- **WHEN** the user does not provide a custom template flag
- **THEN** the generator falls back to the embedded default template for that language

