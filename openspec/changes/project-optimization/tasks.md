## 1. TypeScript Enum Support

- [x] 1.1 Extend Schema type in openapi.go to capture enum values
- [x] 1.2 Update tsType function in generator.go to generate union types for enums
- [x] 1.3 Update TypeScript template (sdk.ts.tmpl) to render enum types
- [x] 1.4 Add tests for enum type generation in generator_test.go

## 2. Go Naming Conventions

- [x] 2.1 Update ToGoName function in types.go to handle acronyms (ID, URL, HTTP, JSON, API, UUID, JWT)
- [x] 2.2 Ensure generated struct fields use proper naming
- [x] 2.3 Update Go template (sdk.go.tmpl) if needed
- [x] 2.4 Add tests for naming conventions in types_test.go

## 3. Multi-Response Generation

- [x] 3.1 Extend GenerationModel to include all response codes with schemas
- [x] 3.2 Update buildOperations to capture error responses (4xx, 5xx)
- [x] 3.3 Update Go template to generate typed error response structs
- [x] 3.4 Update TypeScript template to generate typed error response interfaces
- [x] 3.5 Add tests for multi-response generation

## 4. Request Validation

- [x] 4.1 Add Validate() method generation for request structs in Go template
- [x] 4.2 Generate validation logic for required fields
- [x] 4.3 Generate validation logic for enum values
- [x] 4.4 Generate validation logic for format constraints (email, uuid, date-time)
- [x] 4.5 Add CLI flag to enable/disable validation generation
- [x] 4.6 Add tests for validation generation

## 5. Custom Template Support

- [x] 5.1 Add --go-template and --ts-template CLI flags in main.go
- [x] 5.2 Update Generator to load templates from file paths
- [x] 5.3 Add fallback to embedded templates when file not provided
- [x] 5.4 Document template variables in README
- [x] 5.5 Add example custom templates

## 6. CLI Enhancements

- [x] 6.1 Add --dry-run flag to show what would be generated
- [x] 6.2 Add --verbose flag for detailed progress output
- [x] 6.3 Add --output json flag for machine-readable output
- [x] 6.4 Improve error messages with context (file, line, suggestion)
- [x] 6.5 Add progress reporting for large specs

## 7. Oxlint Configuration Update

- [x] 7.1 Update oxlintrc.json to allow generated patterns (enum unions, etc.)
- [x] 7.2 Add rules to catch common issues in generated code
- [x] 7.3 Test oxlint on generated output passes

## 8. Testing Improvements

- [x] 8.1 Add tests for enum handling in openapi_test.go
- [x] 8.2 Add tests for Go naming in types_test.go
- [x] 8.3 Add tests for multi-response in generator_test.go
- [x] 8.4 Add tests for validation generation
- [x] 8.5 Add integration test with petstore example

## 9. Documentation Updates

- [x] 9.1 Update README with new features (enums, validation, custom templates)
- [x] 9.2 Add CLI reference for new flags
- [x] 9.3 Add migration guide for naming convention changes
- [x] 9.4 Update examples if needed

## 10. Verification

- [x] 10.1 Run full test suite (go test ./...)
- [x] 10.2 Generate petstore SDK and verify output
- [x] 10.3 Run oxlint on generated TypeScript
- [x] 10.4 Test WASM build with generated code
- [x] 10.5 Test dry-run and verbose modes
- [x] 10.6 Test custom template override