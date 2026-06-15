package generator

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

func TestGoType(t *testing.T) {
	g := NewGenerator("", "", "", "", "")

	tests := []struct {
		name     string
		schema   *Schema
		required bool
		expected string
	}{
		{"nil schema", nil, false, "interface{}"},
		{"string", &Schema{Type: "string"}, false, "string"},
		{"integer", &Schema{Type: "integer"}, false, "int"},
		{"int64", &Schema{Type: "integer", Format: "int64"}, false, "int64"},
		{"number", &Schema{Type: "number"}, false, "float64"},
		{"boolean", &Schema{Type: "boolean"}, false, "bool"},
		{"date-time string", &Schema{Type: "string", Format: "date-time"}, false, "string"},
		{"array of string", &Schema{Type: "array", Items: &Schema{Type: "string"}, Required: []string{"items"}}, false, "[]string"},
		{"array no items", &Schema{Type: "array"}, false, "[]interface{}"},
		{"object", &Schema{Type: "object"}, false, "map[string]interface{}"},
		{"object with additionalProperties", &Schema{Type: "object", AdditionalProperties: &Schema{Type: "string"}}, false, "map[string]string"},
		{"ref", &Schema{Ref: "#/components/schemas/Pet"}, false, "Pet"},
		{"unknown type", &Schema{Type: "unknown"}, false, "interface{}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := g.goType(tt.schema, tt.required)
			if got != tt.expected {
				t.Errorf("goType(%v) = %q, want %q", tt.schema, got, tt.expected)
			}
		})
	}
}

func TestTSType(t *testing.T) {
	g := NewGenerator("", "", "", "", "")

	tests := []struct {
		name     string
		schema   *Schema
		required bool
		expected string
	}{
		{"nil schema", nil, false, "any"},
		{"string", &Schema{Type: "string"}, false, "string"},
		{"integer", &Schema{Type: "integer"}, false, "number"},
		{"number", &Schema{Type: "number"}, false, "number"},
		{"boolean", &Schema{Type: "boolean"}, false, "boolean"},
		{"array of string", &Schema{Type: "array", Items: &Schema{Type: "string"}}, false, "Array<string>"},
		{"array no items", &Schema{Type: "array"}, false, "Array<any>"},
		{"object", &Schema{Type: "object"}, false, "Record<string, any>"},
		{"object with additionalProperties", &Schema{Type: "object", AdditionalProperties: &Schema{Type: "string"}}, false, "Record<string, string>"},
		{"ref", &Schema{Ref: "#/components/schemas/Pet"}, false, "Pet"},
		{"unknown type", &Schema{Type: "unknown"}, false, "any"},
		{"enum string", &Schema{Type: "string", Enum: []interface{}{"active", "inactive"}}, false, "'active' | 'inactive'"},
		{"enum number", &Schema{Type: "integer", Enum: []interface{}{1, 2, 3}}, false, "1 | 2 | 3"},
		{"enum boolean", &Schema{Type: "boolean", Enum: []interface{}{true, false}}, false, "true | false"},
		{"enum mixed", &Schema{Type: "string", Enum: []interface{}{"admin", "user", "guest"}}, false, "'admin' | 'user' | 'guest'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := g.tsType(tt.schema, tt.required)
			if got != tt.expected {
				t.Errorf("tsType(%v) = %q, want %q", tt.schema, got, tt.expected)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := NewConfig()
	if cfg.ModuleName != "github.com/fred29910/gowasm" {
		t.Errorf("unexpected ModuleName: %s", cfg.ModuleName)
	}
	if cfg.OutputModule != "generated-sdk" {
		t.Errorf("unexpected OutputModule: %s", cfg.OutputModule)
	}
	if cfg.Package != "generated" {
		t.Errorf("unexpected Package: %s", cfg.Package)
	}
	if cfg.RuntimeImport != "github.com/fred29910/gowasm/pkg/runtime" {
		t.Errorf("unexpected RuntimeImport: %s", cfg.RuntimeImport)
	}
}

func TestNewGeneratorOverrides(t *testing.T) {
	g := NewGenerator("my/mod", "my/output", "mypkg", "/some/path", "my/runtime")
	if g.config.ModuleName != "my/mod" {
		t.Errorf("ModuleName not overridden: %s", g.config.ModuleName)
	}
	if g.config.OutputModule != "my/output" {
		t.Errorf("OutputModule not overridden: %s", g.config.OutputModule)
	}
	if g.config.Package != "mypkg" {
		t.Errorf("Package not overridden: %s", g.config.Package)
	}
	if g.config.RuntimePath != "/some/path" {
		t.Errorf("RuntimePath not overridden: %s", g.config.RuntimePath)
	}
	if g.config.RuntimeImport != "my/runtime" {
		t.Errorf("RuntimeImport not overridden: %s", g.config.RuntimeImport)
	}
}

func TestNewGeneratorFromConfig(t *testing.T) {
	cfg := &Config{
		ModuleName:    "custom/mod",
		OutputModule:  "custom/output",
		Package:       "custompkg",
		RuntimePath:   "/custom/path",
		RuntimeImport: "custom/runtime",
	}
	g := NewGeneratorFromConfig(cfg)
	if g.config.ModuleName != "custom/mod" {
		t.Errorf("unexpected ModuleName: %s", g.config.ModuleName)
	}

	// nil config should use defaults
	g2 := NewGeneratorFromConfig(nil)
	if g2.config.ModuleName != "github.com/fred29910/gowasm" {
		t.Errorf("nil config should use defaults")
	}
}

func TestBuildOperationMultiResponse(t *testing.T) {
	g := NewGenerator("", "", "", "", "")

	op := &Operation{
		OperationID: "getUsers",
		Summary:     "Get users",
		Method:      "GET",
		Path:        "/users",
		Responses: map[string]Response{
			"200": {
				Description: "Successful response",
				Content: map[string]Media{
					"application/json": {
						Schema: &Schema{
							Type: "array",
							Items: &Schema{
								Type: "object",
								Properties: map[string]Schema{
									"id":   {Type: "integer"},
									"name": {Type: "string"},
								},
							},
						},
					},
				},
			},
			"400": {
				Description: "Bad request",
				Content: map[string]Media{
					"application/json": {
						Schema: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"error": {Type: "string"},
							},
						},
					},
				},
			},
			"404": {
				Description: "Not found",
				Content: map[string]Media{
					"application/json": {
						Schema: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"message": {Type: "string"},
							},
						},
					},
				},
			},
		},
	}

	genOp := g.buildOperation(op)

	// Should have three responses
	if len(genOp.Responses) != 3 {
		t.Errorf("expected 3 responses, got %d", len(genOp.Responses))
	}

	// Find each response by code
	var resp200, resp400, resp404 *GeneratedResponse
	for i := range genOp.Responses {
		switch genOp.Responses[i].Code {
		case "200":
			resp200 = &genOp.Responses[i]
		case "400":
			resp400 = &genOp.Responses[i]
		case "404":
			resp404 = &genOp.Responses[i]
		}
	}

	// Verify 200 response exists and is primary
	if resp200 == nil {
		t.Fatal("200 response not found")
	}
	if !resp200.Primary {
		t.Error("200 response should be marked as primary")
	}
	if resp200.Description != "Successful response" {
		t.Errorf("200 description = %q, want %q", resp200.Description, "Successful response")
	}

	// Verify 400 response exists and is not primary
	if resp400 == nil {
		t.Fatal("400 response not found")
	}
	if resp400.Primary {
		t.Error("400 response should not be marked as primary")
	}
	if resp400.Description != "Bad request" {
		t.Errorf("400 description = %q, want %q", resp400.Description, "Bad request")
	}

	// Verify 404 response exists and is not primary
	if resp404 == nil {
		t.Fatal("404 response not found")
	}
	if resp404.Primary {
		t.Error("404 response should not be marked as primary")
	}
	if resp404.Description != "Not found" {
		t.Errorf("404 description = %q, want %q", resp404.Description, "Not found")
	}

	// Verify struct names
	if resp200.StructName != "GetUsersResponse" {
		t.Errorf("200 struct name = %q, want %q", resp200.StructName, "GetUsersResponse")
	}
	if resp400.StructName != "GetUsers400Response" {
		t.Errorf("400 struct name = %q, want %q", resp400.StructName, "GetUsers400Response")
	}
	if resp404.StructName != "GetUsers404Response" {
		t.Errorf("404 struct name = %q, want %q", resp404.StructName, "GetUsers404Response")
	}

	// Verify ResponseType is set to the primary response type
	if genOp.ResponseType == "" {
		t.Error("ResponseType should not be empty")
	}
}

func TestBuildOperationNoPrimaryResponse(t *testing.T) {
	g := NewGenerator("", "", "", "", "")

	// Operation with no 2xx response - first response should be primary
	op := &Operation{
		OperationID: "deleteUser",
		Summary:     "Delete user",
		Method:      "DELETE",
		Path:        "/users/{id}",
		Parameters: []Parameter{
			{Name: "id", In: "path", Required: true, Schema: &Schema{Type: "integer"}},
		},
		Responses: map[string]Response{
			"400": {
				Description: "Bad request",
			},
			"404": {
				Description: "Not found",
			},
		},
	}

	genOp := g.buildOperation(op)

	// Should have two responses
	if len(genOp.Responses) != 2 {
		t.Errorf("expected 2 responses, got %d", len(genOp.Responses))
	}

	// First response (400) should be primary as fallback
	var primaryFound bool
	for _, resp := range genOp.Responses {
		if resp.Primary {
			primaryFound = true
			if resp.Code != "400" {
				t.Errorf("primary response code = %q, want %q", resp.Code, "400")
			}
		}
	}
	if !primaryFound {
		t.Error("no primary response found")
	}
}

func TestBuildOperationResponseTypeMapping(t *testing.T) {
	g := NewGenerator("", "", "", "", "")

	op := &Operation{
		OperationID: "getUser",
		Summary:     "Get user",
		Method:      "GET",
		Path:        "/users/{id}",
		Parameters: []Parameter{
			{Name: "id", In: "path", Required: true, Schema: &Schema{Type: "integer"}},
		},
		Responses: map[string]Response{
			"200": {
				Description: "Success",
				Content: map[string]Media{
					"application/json": {
						Schema: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"id":   {Type: "integer"},
								"name": {Type: "string"},
							},
						},
					},
				},
			},
			"404": {
				Description: "Not found",
				Content: map[string]Media{
					"application/json": {
						Schema: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"error": {Type: "string"},
							},
						},
					},
				},
			},
		},
	}

	genOp := g.buildOperation(op)

	// Verify GoType and TSType mapping for each response
	for _, resp := range genOp.Responses {
		switch resp.Code {
		case "200":
			if resp.GoType != "map[string]interface{}" {
				t.Errorf("200 GoType = %q, want %q", resp.GoType, "map[string]interface{}")
			}
			if resp.TSType != "Record<string, any>" {
				t.Errorf("200 TSType = %q, want %q", resp.TSType, "Record<string, any>")
			}
		case "404":
			if resp.GoType != "map[string]interface{}" {
				t.Errorf("404 GoType = %q, want %q", resp.GoType, "map[string]interface{}")
			}
			if resp.TSType != "Record<string, any>" {
				t.Errorf("404 TSType = %q, want %q", resp.TSType, "Record<string, any>")
			}
		}
	}
}

// TestValidationEnabled verifies that validation methods are generated when validation is enabled.
func TestValidationEnabled(t *testing.T) {
	g := NewGeneratorFromConfig(&Config{
		ModuleName:   "test/mod",
		OutputModule: "test/output",
		Package:      "testpkg",
		Validation:   true,
	})

	doc := &OpenAPI{
		Components: Components{
			Schemas: map[string]Schema{
				"CreateUserRequest": {
					Type: "object",
					Required: []string{"name", "email"},
					Properties: map[string]Schema{
						"name":  {Type: "string"},
						"email": {Type: "string", Format: "email"},
					},
				},
			},
		},
	}

	model := g.buildModel(doc)
	if !model.Validation {
		t.Error("expected Validation to be true")
	}

	tmpl, err := template.New("sdk").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}).Parse(goClientTmpl)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	data := goTemplateData{GenerationModel: model}
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "func (s CreateUserRequest) Validate() error") {
		t.Error("expected Validate() method to be generated for CreateUserRequest")
	}
	if !strings.Contains(output, `return fmt.Errorf("name is required")`) {
		t.Error("expected required field validation for name")
	}
	if !strings.Contains(output, `return fmt.Errorf("email is required")`) {
		t.Error("expected required field validation for email")
	}
	if !strings.Contains(output, `isValidEmail(s.Email)`) {
		t.Error("expected email format validation")
	}
}

// TestValidationDisabled verifies that validation methods are NOT generated when validation is disabled.
func TestValidationDisabled(t *testing.T) {
	g := NewGeneratorFromConfig(&Config{
		ModuleName:   "test/mod",
		OutputModule: "test/output",
		Package:      "testpkg",
		Validation:   false,
	})

	doc := &OpenAPI{
		Components: Components{
			Schemas: map[string]Schema{
				"CreateUserRequest": {
					Type: "object",
					Required: []string{"name"},
					Properties: map[string]Schema{
						"name":  {Type: "string"},
						"email": {Type: "string"},
					},
				},
			},
		},
	}

	model := g.buildModel(doc)
	if model.Validation {
		t.Error("expected Validation to be false")
	}

	tmpl, err := template.New("sdk").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}).Parse(goClientTmpl)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	data := goTemplateData{GenerationModel: model}
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "func (s CreateUserRequest) Validate() error") {
		t.Error("Validate() method should NOT be generated when validation is disabled")
	}
}

// TestValidationRequiredFields verifies that required field validation logic is correct.
func TestValidationRequiredFields(t *testing.T) {
	g := NewGeneratorFromConfig(&Config{
		ModuleName:   "test/mod",
		OutputModule: "test/output",
		Package:      "testpkg",
		Validation:   true,
	})

	doc := &OpenAPI{
		Components: Components{
			Schemas: map[string]Schema{
				"OrderRequest": {
					Type: "object",
					Required: []string{"item", "quantity"},
					Properties: map[string]Schema{
						"item":     {Type: "string"},
						"quantity": {Type: "integer"},
					},
				},
			},
		},
	}

	model := g.buildModel(doc)
	schema := model.Schemas[0]

	// Verify required fields are tracked
	if !schema.Properties[0].Required {
		t.Error("expected 'item' to be required")
	}
	if !schema.Properties[1].Required {
		t.Error("expected 'quantity' to be required")
	}

	tmpl, err := template.New("sdk").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}).Parse(goClientTmpl)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	data := goTemplateData{GenerationModel: model}
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `return fmt.Errorf("item is required")`) {
		t.Error("expected required field validation for item")
	}
	if !strings.Contains(output, `return fmt.Errorf("quantity is required")`) {
		t.Error("expected required field validation for quantity")
	}
	if !strings.Contains(output, `if s.Item == ""`) {
		t.Error("expected string zero-value check for item")
	}
	if !strings.Contains(output, `if s.Quantity == 0`) {
		t.Error("expected int zero-value check for quantity")
	}
}

// TestValidationEnumFields verifies that enum validation logic is correct.
func TestValidationEnumFields(t *testing.T) {
	g := NewGeneratorFromConfig(&Config{
		ModuleName:   "test/mod",
		OutputModule: "test/output",
		Package:      "testpkg",
		Validation:   true,
	})

	doc := &OpenAPI{
		Components: Components{
			Schemas: map[string]Schema{
				"StatusUpdate": {
					Type: "object",
					Required: []string{"status"},
					Properties: map[string]Schema{
						"status": {
							Type: "string",
							Enum: []interface{}{"active", "inactive", "pending"},
						},
					},
				},
			},
		},
	}

	model := g.buildModel(doc)

	tmpl, err := template.New("sdk").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}).Parse(goClientTmpl)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	data := goTemplateData{GenerationModel: model}
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "isValidEnum(s.Status") {
		t.Error("expected enum validation for status field")
	}
	if !strings.Contains(output, `"active"`) || !strings.Contains(output, `"inactive"`) || !strings.Contains(output, `"pending"`) {
		t.Error("expected all enum values to be present in validation")
	}
}

// TestValidationRequestParams verifies validation for request parameters.
func TestValidationRequestParams(t *testing.T) {
	g := NewGeneratorFromConfig(&Config{
		ModuleName:   "test/mod",
		OutputModule: "test/output",
		Package:      "testpkg",
		Validation:   true,
	})

	op := &Operation{
		OperationID: "createOrder",
		Summary:     "Create order",
		Method:      "POST",
		Path:        "/orders",
		Parameters: []Parameter{
			{Name: "id", In: "path", Required: true, Schema: &Schema{Type: "string"}},
			{Name: "category", In: "query", Required: false, Schema: &Schema{
				Type: "string",
				Enum: []interface{}{"electronics", "clothing"},
			}},
		},
		RequestBody: &RequestBody{
			Required: true,
			Content: map[string]Media{
				"application/json": {
					Schema: &Schema{
						Type: "object",
						Required: []string{"product"},
						Properties: map[string]Schema{
							"product": {Type: "string"},
						},
					},
				},
			},
		},
		Responses: map[string]Response{
			"200": {Description: "Success"},
		},
	}

	genOp := g.buildOperation(op)

	if len(genOp.PathParams) != 1 || genOp.PathParams[0].GoName != "ID" {
		t.Errorf("expected path param ID, got %v", genOp.PathParams)
	}
	if !genOp.PathParams[0].Required {
		t.Error("expected path param to be required")
	}

	if len(genOp.QueryParams) != 1 || genOp.QueryParams[0].GoName != "Category" {
		t.Errorf("expected query param Category, got %v", genOp.QueryParams)
	}
	if genOp.QueryParams[0].Required {
		t.Error("expected query param to be optional")
	}
	if len(genOp.QueryParams[0].EnumValues) != 2 {
		t.Errorf("expected 2 enum values, got %d", len(genOp.QueryParams[0].EnumValues))
	}
}

// TestContains verifies the contains helper function used for required field detection.
func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		list     []string
		item     string
		expected bool
	}{
		{"found", []string{"a", "b", "c"}, "b", true},
		{"not found", []string{"a", "b", "c"}, "d", false},
		{"empty list", []string{}, "a", false},
		{"nil list", nil, "a", false},
		{"empty item in list", []string{""}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.list, tt.item)
			if got != tt.expected {
				t.Errorf("contains(%v, %q) = %v, want %v", tt.list, tt.item, got, tt.expected)
			}
		})
	}
}

// TestBuildOperationMultipleErrorResponses verifies operations with multiple error responses (400, 404, 500).
func TestBuildOperationMultipleErrorResponses(t *testing.T) {
	g := NewGenerator("", "", "", "", "")

	op := &Operation{
		OperationID: "getUserProfile",
		Summary:     "Get user profile",
		Method:      "GET",
		Path:        "/users/{id}/profile",
		Parameters: []Parameter{
			{Name: "id", In: "path", Required: true, Schema: &Schema{Type: "integer"}},
		},
		Responses: map[string]Response{
			"200": {
				Description: "Success",
				Content: map[string]Media{
					"application/json": {
						Schema: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"id":   {Type: "integer"},
								"name": {Type: "string"},
							},
						},
					},
				},
			},
			"400": {
				Description: "Bad request - invalid ID",
				Content: map[string]Media{
					"application/json": {
						Schema: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"error": {Type: "string"},
							},
						},
					},
				},
			},
			"404": {
				Description: "User not found",
				Content: map[string]Media{
					"application/json": {
						Schema: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"message": {Type: "string"},
							},
						},
					},
				},
			},
			"500": {
				Description: "Internal server error",
				Content: map[string]Media{
					"application/json": {
						Schema: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"code":    {Type: "integer"},
								"message": {Type: "string"},
							},
						},
					},
				},
			},
		},
	}

	genOp := g.buildOperation(op)

	// Should have four responses
	if len(genOp.Responses) != 4 {
		t.Errorf("expected 4 responses, got %d", len(genOp.Responses))
	}

	// Find each response by code
	var responsesByCode = make(map[string]*GeneratedResponse)
	for i := range genOp.Responses {
		responsesByCode[genOp.Responses[i].Code] = &genOp.Responses[i]
	}

	// Verify all error codes exist
	for _, code := range []string{"400", "404", "500"} {
		if resp, ok := responsesByCode[code]; !ok {
			t.Errorf("response %s not found", code)
		} else if resp.Primary {
			t.Errorf("response %s should not be primary", code)
		}
	}

	// Verify 200 is primary
	if resp200, ok := responsesByCode["200"]; !ok {
		t.Error("200 response not found")
	} else if !resp200.Primary {
		t.Error("200 response should be marked as primary")
	} else if resp200.GoType != "map[string]interface{}" {
		t.Errorf("200 GoType = %q, want %q", resp200.GoType, "map[string]interface{}")
	}

	// Verify struct names for error responses
	expectedStructNames := map[string]string{
		"400": "GetUserProfile400Response",
		"404": "GetUserProfile404Response",
		"500": "GetUserProfile500Response",
	}
	for code, expectedName := range expectedStructNames {
		if resp, ok := responsesByCode[code]; !ok {
			t.Errorf("response %s not found", code)
		} else if resp.StructName != expectedName {
			t.Errorf("response %s struct name = %q, want %q", code, resp.StructName, expectedName)
		}
	}

	// Verify ResponseType is set to the primary response type
	if genOp.ResponseType != "map[string]interface{}" {
		t.Errorf("ResponseType = %q, want %q", genOp.ResponseType, "map[string]interface{}")
	}
}

// TestBuildOperationNoSuccessResponse verifies operations with only error responses (no 200).
func TestBuildOperationNoSuccessResponse(t *testing.T) {
	g := NewGenerator("", "", "", "", "")

	op := &Operation{
		OperationID: "validateToken",
		Summary:     "Validate token",
		Method:      "POST",
		Path:        "/auth/validate",
		RequestBody: &RequestBody{
			Required: true,
			Content: map[string]Media{
				"application/json": {
					Schema: &Schema{
						Type: "object",
						Properties: map[string]Schema{
							"token": {Type: "string"},
						},
					},
				},
			},
		},
		Responses: map[string]Response{
			"401": {
				Description: "Unauthorized - invalid token",
				Content: map[string]Media{
					"application/json": {
						Schema: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"error": {Type: "string"},
							},
						},
					},
				},
			},
			"403": {
				Description: "Forbidden - insufficient permissions",
				Content: map[string]Media{
					"application/json": {
						Schema: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"message": {Type: "string"},
							},
						},
					},
				},
			},
			"500": {
				Description: "Internal server error",
				Content: map[string]Media{
					"application/json": {
						Schema: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"code":    {Type: "integer"},
								"message": {Type: "string"},
							},
						},
					},
				},
			},
		},
	}

	genOp := g.buildOperation(op)

	// Should have three responses
	if len(genOp.Responses) != 3 {
		t.Errorf("expected 3 responses, got %d", len(genOp.Responses))
	}

	// Find each response by code
	var responsesByCode = make(map[string]*GeneratedResponse)
	for i := range genOp.Responses {
		responsesByCode[genOp.Responses[i].Code] = &genOp.Responses[i]
	}

	// Exactly one response should be primary (fallback: first from map iteration)
	var primaryCount int
	var primaryCode string
	for _, resp := range genOp.Responses {
		if resp.Primary {
			primaryCount++
			primaryCode = resp.Code
		}
	}
	if primaryCount != 1 {
		t.Errorf("expected exactly 1 primary response, got %d", primaryCount)
	}

	// Verify ResponseType is set from the primary error response
	if genOp.ResponseType != "map[string]interface{}" {
		t.Errorf("ResponseType = %q, want %q", genOp.ResponseType, "map[string]interface{}")
	}

	// The primary response gets StructName = "ValidateTokenResponse" (no code suffix).
	// The other responses get StructName = "ValidateToken{code}Response".
	for _, resp := range genOp.Responses {
		if resp.Primary {
			if resp.StructName != "ValidateTokenResponse" {
				t.Errorf("primary response %s struct name = %q, want %q", resp.Code, resp.StructName, "ValidateTokenResponse")
			}
		} else {
			expected := "ValidateToken" + resp.Code + "Response"
			if resp.StructName != expected {
				t.Errorf("response %s struct name = %q, want %q", resp.Code, resp.StructName, expected)
			}
		}
	}
	_ = primaryCode

	// Verify all responses have proper descriptions
	expectedDescriptions := map[string]string{
		"401": "Unauthorized - invalid token",
		"403": "Forbidden - insufficient permissions",
		"500": "Internal server error",
	}
	for code, expectedDesc := range expectedDescriptions {
		if resp, ok := responsesByCode[code]; !ok {
			t.Errorf("response %s not found", code)
		} else if resp.Description != expectedDesc {
			t.Errorf("response %s description = %q, want %q", code, resp.Description, expectedDesc)
		}
	}
}

// TestBuildOperationResponseWithSchema verifies operations where responses have schemas.
func TestBuildOperationResponseWithSchema(t *testing.T) {
	g := NewGenerator("", "", "", "", "")

	op := &Operation{
		OperationID: "getProduct",
		Summary:     "Get product",
		Method:      "GET",
		Path:        "/products/{id}",
		Parameters: []Parameter{
			{Name: "id", In: "path", Required: true, Schema: &Schema{Type: "integer"}},
		},
		Responses: map[string]Response{
			"200": {
				Description: "Success",
				Content: map[string]Media{
					"application/json": {
						Schema: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"id":    {Type: "integer"},
								"name":  {Type: "string"},
								"price": {Type: "number"},
							},
						},
					},
				},
			},
			"404": {
				Description: "Product not found",
				Content: map[string]Media{
					"application/json": {
						Schema: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"error": {Type: "string"},
							},
						},
					},
				},
			},
		},
	}

	genOp := g.buildOperation(op)

	// Verify 200 response has schema-derived type
	var resp200 *GeneratedResponse
	for i := range genOp.Responses {
		if genOp.Responses[i].Code == "200" {
			resp200 = &genOp.Responses[i]
			break
		}
	}
	if resp200 == nil {
		t.Fatal("200 response not found")
	}
	if resp200.GoType != "map[string]interface{}" {
		t.Errorf("200 GoType = %q, want %q", resp200.GoType, "map[string]interface{}")
	}
	if resp200.TSType != "Record<string, any>" {
		t.Errorf("200 TSType = %q, want %q", resp200.TSType, "Record<string, any>")
	}

	// Verify 404 response has schema-derived type
	var resp404 *GeneratedResponse
	for i := range genOp.Responses {
		if genOp.Responses[i].Code == "404" {
			resp404 = &genOp.Responses[i]
			break
		}
	}
	if resp404 == nil {
		t.Fatal("404 response not found")
	}
	if resp404.GoType != "map[string]interface{}" {
		t.Errorf("404 GoType = %q, want %q", resp404.GoType, "map[string]interface{}")
	}
	if resp404.TSType != "Record<string, any>" {
		t.Errorf("404 TSType = %q, want %q", resp404.TSType, "Record<string, any>")
	}
}

// TestBuildOperationResponseWithoutSchema verifies operations where responses have no schemas.
func TestBuildOperationResponseWithoutSchema(t *testing.T) {
	g := NewGenerator("", "", "", "", "")

	op := &Operation{
		OperationID: "deleteItem",
		Summary:     "Delete item",
		Method:      "DELETE",
		Path:        "/items/{id}",
		Parameters: []Parameter{
			{Name: "id", In: "path", Required: true, Schema: &Schema{Type: "integer"}},
		},
		Responses: map[string]Response{
			"204": {
				Description: "No content - successfully deleted",
				// No Content field - empty response
			},
			"404": {
				Description: "Item not found",
				// No Content field - empty response
			},
			"500": {
				Description: "Internal server error",
				// No Content field - empty response
			},
		},
	}

	genOp := g.buildOperation(op)

	// Should have three responses
	if len(genOp.Responses) != 3 {
		t.Errorf("expected 3 responses, got %d", len(genOp.Responses))
	}

	// Find each response by code
	var responsesByCode = make(map[string]*GeneratedResponse)
	for i := range genOp.Responses {
		responsesByCode[genOp.Responses[i].Code] = &genOp.Responses[i]
	}

	// Verify all responses have default types (interface{} / any) when no schema is provided
	for _, code := range []string{"204", "404", "500"} {
		if resp, ok := responsesByCode[code]; !ok {
			t.Errorf("response %s not found", code)
		} else {
			if resp.GoType != "interface{}" {
				t.Errorf("response %s GoType = %q, want %q", code, resp.GoType, "interface{}")
			}
			if resp.TSType != "any" {
				t.Errorf("response %s TSType = %q, want %q", code, resp.TSType, "any")
			}
		}
	}

	// Verify 204 is primary (first 2xx)
	if resp204, ok := responsesByCode["204"]; !ok {
		t.Error("204 response not found")
	} else if !resp204.Primary {
		t.Error("204 response should be marked as primary")
	}

	// Verify ResponseType is set from the primary response
	if genOp.ResponseType != "interface{}" {
		t.Errorf("ResponseType = %q, want %q", genOp.ResponseType, "interface{}")
	}

	// Verify struct names
	expectedStructNames := map[string]string{
		"204": "DeleteItemResponse",
		"404": "DeleteItem404Response",
		"500": "DeleteItem500Response",
	}
	for code, expectedName := range expectedStructNames {
		if resp, ok := responsesByCode[code]; !ok {
			t.Errorf("response %s not found", code)
		} else if resp.StructName != expectedName {
			t.Errorf("response %s struct name = %q, want %q", code, resp.StructName, expectedName)
		}
	}
}

// TestValidationRequiredStringField verifies that required string fields generate empty string checks.
func TestValidationRequiredStringField(t *testing.T) {
	g := NewGeneratorFromConfig(&Config{
		ModuleName:   "test/mod",
		OutputModule: "test/output",
		Package:      "testpkg",
		Validation:   true,
	})

	doc := &OpenAPI{
		Components: Components{
			Schemas: map[string]Schema{
				"StringRequired": {
					Type: "object",
					Required: []string{"name"},
					Properties: map[string]Schema{
						"name": {Type: "string"},
					},
				},
			},
		},
	}

	model := g.buildModel(doc)
	schema := model.Schemas[0]

	if !schema.Properties[0].Required {
		t.Error("expected 'name' to be required")
	}
	if schema.Properties[0].GoType != "string" {
		t.Errorf("expected GoType 'string', got %q", schema.Properties[0].GoType)
	}

	tmpl, err := template.New("sdk").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}).Parse(goClientTmpl)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	data := goTemplateData{GenerationModel: model}
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `if s.Name == ""`) {
		t.Error("expected empty string zero-value check for name")
	}
	if !strings.Contains(output, `return fmt.Errorf("name is required")`) {
		t.Error("expected required field validation error for name")
	}
}

// TestValidationRequiredIntField verifies that required integer fields generate zero-value checks.
func TestValidationRequiredIntField(t *testing.T) {
	g := NewGeneratorFromConfig(&Config{
		ModuleName:   "test/mod",
		OutputModule: "test/output",
		Package:      "testpkg",
		Validation:   true,
	})

	doc := &OpenAPI{
		Components: Components{
			Schemas: map[string]Schema{
				"IntRequired": {
					Type: "object",
					Required: []string{"count"},
					Properties: map[string]Schema{
						"count": {Type: "integer"},
					},
				},
			},
		},
	}

	model := g.buildModel(doc)
	schema := model.Schemas[0]

	if !schema.Properties[0].Required {
		t.Error("expected 'count' to be required")
	}
	if schema.Properties[0].GoType != "int" {
		t.Errorf("expected GoType 'int', got %q", schema.Properties[0].GoType)
	}

	tmpl, err := template.New("sdk").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}).Parse(goClientTmpl)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	data := goTemplateData{GenerationModel: model}
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `if s.Count == 0`) {
		t.Error("expected int zero-value check for count")
	}
	if !strings.Contains(output, `return fmt.Errorf("count is required")`) {
		t.Error("expected required field validation error for count")
	}
}

// TestValidationRequiredBoolField verifies that required boolean fields generate false-value checks.
func TestValidationRequiredBoolField(t *testing.T) {
	g := NewGeneratorFromConfig(&Config{
		ModuleName:   "test/mod",
		OutputModule: "test/output",
		Package:      "testpkg",
		Validation:   true,
	})

	doc := &OpenAPI{
		Components: Components{
			Schemas: map[string]Schema{
				"BoolRequired": {
					Type: "object",
					Required: []string{"active"},
					Properties: map[string]Schema{
						"active": {Type: "boolean"},
					},
				},
			},
		},
	}

	model := g.buildModel(doc)
	schema := model.Schemas[0]

	if !schema.Properties[0].Required {
		t.Error("expected 'active' to be required")
	}
	if schema.Properties[0].GoType != "bool" {
		t.Errorf("expected GoType 'bool', got %q", schema.Properties[0].GoType)
	}

	tmpl, err := template.New("sdk").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}).Parse(goClientTmpl)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	data := goTemplateData{GenerationModel: model}
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `if s.Active == false`) {
		t.Error("expected bool zero-value check for active")
	}
	if !strings.Contains(output, `return fmt.Errorf("active is required")`) {
		t.Error("expected required field validation error for active")
	}
}

// TestValidationEnumField verifies that enum validation is generated for fields with allowed values.
func TestValidationEnumField(t *testing.T) {
	g := NewGeneratorFromConfig(&Config{
		ModuleName:   "test/mod",
		OutputModule: "test/output",
		Package:      "testpkg",
		Validation:   true,
	})

	doc := &OpenAPI{
		Components: Components{
			Schemas: map[string]Schema{
				"RoleRequest": {
					Type: "object",
					Required: []string{"role"},
					Properties: map[string]Schema{
						"role": {
							Type: "string",
							Enum: []interface{}{"admin", "user", "guest"},
						},
					},
				},
			},
		},
	}

	model := g.buildModel(doc)
	schema := model.Schemas[0]

	if len(schema.Properties[0].EnumValues) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(schema.Properties[0].EnumValues))
	}

	tmpl, err := template.New("sdk").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}).Parse(goClientTmpl)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	data := goTemplateData{GenerationModel: model}
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `isValidEnum(s.Role`) {
		t.Error("expected enum validation for role field")
	}
	if !strings.Contains(output, `"admin"`) {
		t.Error("expected 'admin' in enum values")
	}
	if !strings.Contains(output, `"user"`) {
		t.Error("expected 'user' in enum values")
	}
	if !strings.Contains(output, `"guest"`) {
		t.Error("expected 'guest' in enum values")
	}
	if !strings.Contains(output, `role must be one of the allowed values`) {
		t.Error("expected enum validation error message")
	}
}

// TestValidationFormatEmail verifies that email format validation is generated for fields with email format.
func TestValidationFormatEmail(t *testing.T) {
	g := NewGeneratorFromConfig(&Config{
		ModuleName:   "test/mod",
		OutputModule: "test/output",
		Package:      "testpkg",
		Validation:   true,
	})

	doc := &OpenAPI{
		Components: Components{
			Schemas: map[string]Schema{
				"ContactRequest": {
					Type: "object",
					Required: []string{"email"},
					Properties: map[string]Schema{
						"email": {
							Type:   "string",
							Format: "email",
						},
					},
				},
			},
		},
	}

	model := g.buildModel(doc)
	schema := model.Schemas[0]

	if schema.Properties[0].Format != "email" {
		t.Errorf("expected format 'email', got %q", schema.Properties[0].Format)
	}

	tmpl, err := template.New("sdk").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}).Parse(goClientTmpl)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	data := goTemplateData{GenerationModel: model}
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `isValidEmail(s.Email)`) {
		t.Error("expected email format validation call")
	}
	if !strings.Contains(output, `email must be a valid email`) {
		t.Error("expected email format validation error message")
	}
}

// TestValidationFormatUUID verifies that UUID format validation is generated for fields with uuid format.
func TestValidationFormatUUID(t *testing.T) {
	g := NewGeneratorFromConfig(&Config{
		ModuleName:   "test/mod",
		OutputModule: "test/output",
		Package:      "testpkg",
		Validation:   true,
	})

	doc := &OpenAPI{
		Components: Components{
			Schemas: map[string]Schema{
				"ResourceRequest": {
					Type: "object",
					Required: []string{"id"},
					Properties: map[string]Schema{
						"id": {
							Type:   "string",
							Format: "uuid",
						},
					},
				},
			},
		},
	}

	model := g.buildModel(doc)
	schema := model.Schemas[0]

	if schema.Properties[0].Format != "uuid" {
		t.Errorf("expected format 'uuid', got %q", schema.Properties[0].Format)
	}

	tmpl, err := template.New("sdk").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}).Parse(goClientTmpl)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	data := goTemplateData{GenerationModel: model}
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `isValidUUID(s.ID)`) {
		t.Error("expected UUID format validation call")
	}
	if !strings.Contains(output, `id must be a valid uuid`) {
		t.Error("expected UUID format validation error message")
	}
}

// TestValidationFormatDateTime verifies that date-time format validation is generated for fields with date-time format.
func TestValidationFormatDateTime(t *testing.T) {
	g := NewGeneratorFromConfig(&Config{
		ModuleName:   "test/mod",
		OutputModule: "test/output",
		Package:      "testpkg",
		Validation:   true,
	})

	doc := &OpenAPI{
		Components: Components{
			Schemas: map[string]Schema{
				"EventRequest": {
					Type: "object",
					Required: []string{"startTime"},
					Properties: map[string]Schema{
						"startTime": {
							Type:   "string",
							Format: "date-time",
						},
					},
				},
			},
		},
	}

	model := g.buildModel(doc)
	schema := model.Schemas[0]

	if schema.Properties[0].Format != "date-time" {
		t.Errorf("expected format 'date-time', got %q", schema.Properties[0].Format)
	}

	tmpl, err := template.New("sdk").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}).Parse(goClientTmpl)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	data := goTemplateData{GenerationModel: model}
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `isValidDateTime(s.StartTime)`) {
		t.Error("expected date-time format validation call")
	}
	if !strings.Contains(output, `startTime must be a valid date-time`) {
		t.Error("expected date-time format validation error message")
	}
}

// TestGeneratedSchemaValidationFields verifies that schema properties track validation metadata.
func TestGeneratedSchemaValidationFields(t *testing.T) {
	g := NewGeneratorFromConfig(&Config{
		ModuleName:   "test/mod",
		OutputModule: "test/output",
		Package:      "testpkg",
		Validation:   true,
	})

	doc := &OpenAPI{
		Components: Components{
			Schemas: map[string]Schema{
				"TestSchema": {
					Type: "object",
					Required: []string{"requiredField"},
					Properties: map[string]Schema{
						"requiredField": {Type: "string"},
						"optionalField": {Type: "string"},
						"enumField": {
							Type: "string",
							Enum: []interface{}{"a", "b"},
						},
						"formattedField": {
							Type:   "string",
							Format: "email",
						},
					},
				},
			},
		},
	}

	model := g.buildModel(doc)
	schema := model.Schemas[0]

	// Find each property
	props := make(map[string]GeneratedProperty)
	for _, p := range schema.Properties {
		props[p.Name] = p
	}

	// Verify required field
	if !props["requiredField"].Required {
		t.Error("expected requiredField to be required")
	}

	// Verify optional field
	if props["optionalField"].Required {
		t.Error("expected optionalField to be optional")
	}

	// Verify enum field
	if len(props["enumField"].EnumValues) != 2 {
		t.Errorf("expected 2 enum values, got %d", len(props["enumField"].EnumValues))
	}

	// Verify format field
	if props["formattedField"].Format != "email" {
		t.Errorf("expected format 'email', got %q", props["formattedField"].Format)
	}
}

// TestIntegrationPetstore verifies the entire generation pipeline works correctly
// with the petstore OpenAPI spec.
func TestIntegrationPetstore(t *testing.T) {
	// Create a temporary directory for output
	outDir := t.TempDir()

	// Path to the petstore spec
	specPath := filepath.Join("..", "..", "examples", "petstore", "openapi.yaml")
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skipf("petstore spec not found at %s", specPath)
	}

	// Configure generator
	cfg := &Config{
		ModuleName:    "test/petstore",
		OutputModule:  "petstore-generated",
		Package:       "generated",
		RuntimePath:   filepath.Join("..", "..", "pkg", "runtime"),
		RuntimeImport: "github.com/fred29910/gowasm/pkg/runtime",
		Validation:    true,
	}
	g := NewGeneratorFromConfig(cfg)

	// Step 1: Parse and generate
	result, err := g.Generate(specPath, outDir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify generation result
	if len(result.Files) != 3 {
		t.Errorf("expected 3 generated files, got %d", len(result.Files))
	}

	// Step 2: Read generated files
	goContent, err := os.ReadFile(filepath.Join(outDir, "generated.go"))
	if err != nil {
		t.Fatalf("failed to read generated.go: %v", err)
	}
	tsContent, err := os.ReadFile(filepath.Join(outDir, "sdk.ts"))
	if err != nil {
		t.Fatalf("failed to read sdk.ts: %v", err)
	}

	goStr := string(goContent)
	tsStr := string(tsContent)

	// Step 3: Verify expected types in Go
	expectedGoTypes := []string{
		"type Pet struct",
		"type CreatePetRequest struct",
		"type CreatePetResponse struct",
		"type FindPetsByStatusRequest struct",
		"type FindPetsByStatusResponse struct",
		"type GetPetByIDRequest struct",
		"type GetPetByIDResponse struct",
	}
	for _, expected := range expectedGoTypes {
		if !strings.Contains(goStr, expected) {
			t.Errorf("generated Go code missing expected type: %s", expected)
		}
	}

	// Step 4: Verify expected types in TypeScript
	expectedTSTypes := []string{
		"export interface Pet",
		"export interface CreatePetRequest",
		"export interface CreatePetResponse",
		"export interface FindPetsByStatusRequest",
		"export interface FindPetsByStatusResponse",
		"export interface GetPetByIDRequest",
		"export interface GetPetByIDResponse",
	}
	for _, expected := range expectedTSTypes {
		if !strings.Contains(tsStr, expected) {
			t.Errorf("generated TypeScript code missing expected type: %s", expected)
		}
	}

	// Step 5: Verify expected functions in Go
	expectedGoFunctions := []string{
		"func CreatePetRequestToRequest(",
		"func CreatePetRequestCall(",
		"func FindPetsByStatusRequestToRequest(",
		"func FindPetsByStatusRequestCall(",
		"func GetPetByIDRequestToRequest(",
		"func GetPetByIDRequestCall(",
	}
	for _, expected := range expectedGoFunctions {
		if !strings.Contains(goStr, expected) {
			t.Errorf("generated Go code missing expected function: %s", expected)
		}
	}

	// Step 6: Verify expected functions in TypeScript
	expectedTSFunctions := []string{
		"export async function createPet(",
		"export async function findPetsByStatus(",
		"export async function getPetByID(",
	}
	for _, expected := range expectedTSFunctions {
		if !strings.Contains(tsStr, expected) {
			t.Errorf("generated TypeScript code missing expected function: %s", expected)
		}
	}

	// Step 7: Verify validation methods exist in Go
	expectedValidation := []string{
		"func (s Pet) Validate() error",
		"func (r GetPetByIDRequest) Validate() error",
		"name is required",
		"petId is required",
	}
	for _, expected := range expectedValidation {
		if !strings.Contains(goStr, expected) {
			t.Errorf("generated Go code missing expected validation: %s", expected)
		}
	}

	// Step 8: Verify Pet schema properties in Go
	if !strings.Contains(goStr, "ID int64") {
		t.Error("expected Pet.ID to be int64")
	}
	if !strings.Contains(goStr, "Name string") {
		t.Error("expected Pet.Name to be string")
	}
	if !strings.Contains(goStr, "Status string") {
		t.Error("expected Pet.Status to be string")
	}

	// Step 9: Verify enum values in TypeScript
	if !strings.Contains(tsStr, "'available' | 'pending' | 'sold'") {
		t.Error("expected Pet status enum in TypeScript")
	}

	// Step 10: Write a go.mod to the output directory so we can compile
	workspaceRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("failed to get workspace root: %v", err)
	}
	goModContent := `module test/petstore

go 1.25.1

require github.com/fred29910/gowasm v0.0.0

replace github.com/fred29910/gowasm => ` + workspaceRoot + `
`
	if err := os.WriteFile(filepath.Join(outDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Step 11: Verify Go code compiles
	buildCmd := exec.Command("go", "build", "./...")
	buildCmd.Dir = outDir
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Errorf("generated Go code failed to compile: %v\nOutput: %s", err, buildOutput)
	}

	// Step 12: Verify TypeScript passes oxlint (if available)
	if _, err := exec.LookPath("oxlint"); err == nil {
		// Write oxlint config for the generated code
		oxlintConfig := `{
  "plugins": ["typescript"],
  "env": { "browser": true, "es2022": true },
  "rules": {
    "typescript/no-unused-vars": "error",
    "no-empty": "error"
  }
}`
		if err := os.WriteFile(filepath.Join(outDir, "oxlintrc.json"), []byte(oxlintConfig), 0644); err != nil {
			t.Fatalf("failed to write oxlintrc.json: %v", err)
		}

		lintCmd := exec.Command("oxlint", "-c", "oxlintrc.json", "sdk.ts")
		lintCmd.Dir = outDir
		lintOutput, err := lintCmd.CombinedOutput()
		if err != nil {
			t.Errorf("generated TypeScript failed oxlint: %v\nOutput: %s", err, lintOutput)
		}
	} else {
		t.Log("oxlint not found, skipping TypeScript lint check")
	}
}
