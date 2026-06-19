## ADDED Requirements

### Requirement: Dry-run generation
The system SHALL support a dry-run generation mode that analyzes the OpenAPI input and reports the files that would be generated without writing output files.

#### Scenario: Dry-run without writing files
- **WHEN** the user runs generation with dry-run enabled
- **THEN** the command prints the planned output file list and exits without creating or modifying generated files

#### Scenario: Dry-run with invalid spec
- **WHEN** the OpenAPI input is invalid during a dry-run
- **THEN** the command reports the validation error instead of producing an output file list
