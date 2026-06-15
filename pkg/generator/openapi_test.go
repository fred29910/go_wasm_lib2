package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestParseOpenAPI(t *testing.T) {
	// Use the example spec
	doc, err := ParseOpenAPI("../../examples/petstore/openapi.yaml")
	if err != nil {
		t.Fatalf("ParseOpenAPI failed: %v", err)
	}

	if doc.OpenAPI != "3.0.0" {
		t.Errorf("expected OpenAPI version 3.0.0, got %s", doc.OpenAPI)
	}

	if len(doc.Components.Schemas) != 1 {
		t.Errorf("expected 1 schema, got %d", len(doc.Components.Schemas))
	}

	if _, ok := doc.Components.Schemas["Pet"]; !ok {
		t.Error("expected Pet schema to exist")
	}

	ops := doc.Operations()
	if len(ops) != 3 {
		t.Errorf("expected 3 operations, got %d", len(ops))
	}

	if doc.DefaultServer() != "https://petstore3.swagger.io/api/v3" {
		t.Errorf("unexpected default server: %s", doc.DefaultServer())
	}
}

func TestParseOpenAPI_MissingFile(t *testing.T) {
	_, err := ParseOpenAPI("/nonexistent/path.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestParseOpenAPI_InvalidYAML(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "invalid.yaml")
	if err := os.WriteFile(tmpFile, []byte("not: valid: yaml: ["), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseOpenAPI(tmpFile)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestParseOpenAPI_MissingOpenAPIField(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "noversion.yaml")
	if err := os.WriteFile(tmpFile, []byte("info:\n  title: Test\n  version: \"1.0\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseOpenAPI(tmpFile)
	if err == nil {
		t.Error("expected error for missing openapi field")
	}
}

func TestParseOpenAPI_DuplicateOperationID(t *testing.T) {
	spec := `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths:
  /a:
    get:
      operationId: sameId
      responses:
        "200":
          description: OK
  /b:
    get:
      operationId: sameId
      responses:
        "200":
          description: OK
`
	tmpFile := filepath.Join(t.TempDir(), "dup.yaml")
	if err := os.WriteFile(tmpFile, []byte(spec), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseOpenAPI(tmpFile)
	if err == nil {
		t.Error("expected error for duplicate operationId")
	}
}

func TestOperationsOrder(t *testing.T) {
	doc, err := ParseOpenAPI("../../examples/petstore/openapi.yaml")
	if err != nil {
		t.Fatalf("ParseOpenAPI failed: %v", err)
	}

	ops := doc.Operations()
	if len(ops) < 2 {
		t.Skip("need at least 2 operations")
	}

	// Operations should be sorted by path
	for i := 1; i < len(ops); i++ {
		if ops[i].Path < ops[i-1].Path {
			t.Errorf("operations not sorted: %s before %s", ops[i-1].Path, ops[i].Path)
		}
	}
}

func TestHasEnum(t *testing.T) {
	tests := []struct {
		name     string
		schema   *Schema
		expected bool
	}{
		{"nil enum", &Schema{Type: "string"}, false},
		{"empty enum", &Schema{Type: "string", Enum: []interface{}{}}, false},
		{"with enum", &Schema{Type: "string", Enum: []interface{}{"a", "b"}}, true},
		{"single enum", &Schema{Type: "integer", Enum: []interface{}{1}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.schema.HasEnum()
			if got != tt.expected {
				t.Errorf("HasEnum() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseOpenAPIWithEnumSchema(t *testing.T) {
	spec := `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths:
  /items:
    get:
      operationId: getItems
      responses:
        "200":
          description: OK
components:
  schemas:
    Status:
      type: string
      enum:
        - active
        - inactive
        - pending
    Priority:
      type: integer
      enum:
        - 1
        - 2
        - 3
    Size:
      type: string
      enum:
        - small
        - medium
        - large
    EmptyEnum:
      type: string
      enum: []
    SingleEnum:
      type: string
      enum:
        - only
    MixedTypes:
      type: string
      enum:
        - "1"
        - "abc"
        - ""
`
	tmpFile := filepath.Join(t.TempDir(), "enum_schema.yaml")
	if err := os.WriteFile(tmpFile, []byte(spec), 0644); err != nil {
		t.Fatal(err)
	}

	doc, err := ParseOpenAPI(tmpFile)
	if err != nil {
		t.Fatalf("ParseOpenAPI failed: %v", err)
	}

	if len(doc.Components.Schemas) != 6 {
		t.Errorf("expected 6 schemas, got %d", len(doc.Components.Schemas))
	}

	// Test Status enum
	status, ok := doc.Components.Schemas["Status"]
	if !ok {
		t.Fatal("expected Status schema")
	}
	if !status.HasEnum() {
		t.Error("Status should have enum")
	}
	expected := []interface{}{"active", "inactive", "pending"}
	if len(status.Enum) != len(expected) {
		t.Errorf("Status enum length = %d, want %d", len(status.Enum), len(expected))
	}
	for i, v := range expected {
		if status.Enum[i] != v {
			t.Errorf("Status enum[%d] = %v, want %v", i, status.Enum[i], v)
		}
	}

	// Test Priority enum (integer)
	priority, ok := doc.Components.Schemas["Priority"]
	if !ok {
		t.Fatal("expected Priority schema")
	}
	if !priority.HasEnum() {
		t.Error("Priority should have enum")
	}
	if len(priority.Enum) != 3 {
		t.Errorf("Priority enum length = %d, want 3", len(priority.Enum))
	}
	for i, v := range []interface{}{1, 2, 3} {
		if priority.Enum[i] != v {
			t.Errorf("Priority enum[%d] = %v, want %v", i, priority.Enum[i], v)
		}
	}

	// Test EmptyEnum
	empty, ok := doc.Components.Schemas["EmptyEnum"]
	if !ok {
		t.Fatal("expected EmptyEnum schema")
	}
	if empty.HasEnum() {
		t.Error("EmptyEnum should not have enum")
	}

	// Test SingleEnum
	single, ok := doc.Components.Schemas["SingleEnum"]
	if !ok {
		t.Fatal("expected SingleEnum schema")
	}
	if !single.HasEnum() {
		t.Error("SingleEnum should have enum")
	}
	if len(single.Enum) != 1 {
		t.Errorf("SingleEnum enum length = %d, want 1", len(single.Enum))
	}
	if single.Enum[0] != "only" {
		t.Errorf("SingleEnum enum[0] = %v, want 'only'", single.Enum[0])
	}

	// Test MixedTypes enum
	mixed, ok := doc.Components.Schemas["MixedTypes"]
	if !ok {
		t.Fatal("expected MixedTypes schema")
	}
	if !mixed.HasEnum() {
		t.Error("MixedTypes should have enum")
	}
	if len(mixed.Enum) != 3 {
		t.Errorf("MixedTypes enum length = %d, want 3", len(mixed.Enum))
	}
}

func TestParseOpenAPIWithEnumParameter(t *testing.T) {
	spec := `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths:
  /search:
    get:
      operationId: search
      parameters:
        - name: status
          in: query
          required: true
          schema:
            type: string
            enum:
              - active
              - inactive
              - archived
        - name: limit
          in: query
          required: false
          schema:
            type: integer
            enum:
              - 10
              - 25
              - 50
      responses:
        "200":
          description: OK
`
	tmpFile := filepath.Join(t.TempDir(), "enum_param.yaml")
	if err := os.WriteFile(tmpFile, []byte(spec), 0644); err != nil {
		t.Fatal(err)
	}

	doc, err := ParseOpenAPI(tmpFile)
	if err != nil {
		t.Fatalf("ParseOpenAPI failed: %v", err)
	}

	ops := doc.Operations()
	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}

	op := ops[0]
	if len(op.Parameters) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(op.Parameters))
	}

	// Test status parameter
	var statusParam *Parameter
	for i, p := range op.Parameters {
		if p.Name == "status" {
			statusParam = &op.Parameters[i]
			break
		}
	}
	if statusParam == nil {
		t.Fatal("expected status parameter")
	}
	if statusParam.Schema == nil {
		t.Fatal("expected status parameter schema")
	}
	if !statusParam.Schema.HasEnum() {
		t.Error("status parameter should have enum")
	}
	expected := []interface{}{"active", "inactive", "archived"}
	if len(statusParam.Schema.Enum) != len(expected) {
		t.Errorf("status enum length = %d, want %d", len(statusParam.Schema.Enum), len(expected))
	}
	for i, v := range expected {
		if statusParam.Schema.Enum[i] != v {
			t.Errorf("status enum[%d] = %v, want %v", i, statusParam.Schema.Enum[i], v)
		}
	}

	// Test limit parameter
	var limitParam *Parameter
	for i, p := range op.Parameters {
		if p.Name == "limit" {
			limitParam = &op.Parameters[i]
			break
		}
	}
	if limitParam == nil {
		t.Fatal("expected limit parameter")
	}
	if limitParam.Schema == nil {
		t.Fatal("expected limit parameter schema")
	}
	if !limitParam.Schema.HasEnum() {
		t.Error("limit parameter should have enum")
	}
	if len(limitParam.Schema.Enum) != 3 {
		t.Errorf("limit enum length = %d, want 3", len(limitParam.Schema.Enum))
	}
}

func TestEnumValuesPreserved(t *testing.T) {
	spec := `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths:
  /status:
    get:
      operationId: getStatus
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/StatusResponse'
    post:
      operationId: setStatus
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/StatusRequest'
      responses:
        "200":
          description: OK
components:
  schemas:
    StatusResponse:
      type: object
      properties:
        status:
          type: string
          enum:
            - active
            - inactive
            - pending
        priority:
          type: integer
          enum:
            - 0
            - 1
            - 2
            - 3
    StatusRequest:
      type: object
      properties:
        newStatus:
          type: string
          enum:
            - active
            - inactive
            - deleted
`
	tmpFile := filepath.Join(t.TempDir(), "preserved.yaml")
	if err := os.WriteFile(tmpFile, []byte(spec), 0644); err != nil {
		t.Fatal(err)
	}

	doc, err := ParseOpenAPI(tmpFile)
	if err != nil {
		t.Fatalf("ParseOpenAPI failed: %v", err)
	}

	// Test StatusResponse
	statusResp, ok := doc.Components.Schemas["StatusResponse"]
	if !ok {
		t.Fatal("expected StatusResponse schema")
	}
	if statusResp.Properties == nil {
		t.Fatal("expected StatusResponse to have properties")
	}

	// Test status property enum
	statusProp, ok := statusResp.Properties["status"]
	if !ok {
		t.Fatal("expected status property")
	}
	if !statusProp.HasEnum() {
		t.Error("status property should have enum")
	}
	expectedStatus := []interface{}{"active", "inactive", "pending"}
	if len(statusProp.Enum) != len(expectedStatus) {
		t.Errorf("status enum length = %d, want %d", len(statusProp.Enum), len(expectedStatus))
	}
	for i, v := range expectedStatus {
		if statusProp.Enum[i] != v {
			t.Errorf("status enum[%d] = %v, want %v", i, statusProp.Enum[i], v)
		}
	}

	// Test priority property enum
	priorityProp, ok := statusResp.Properties["priority"]
	if !ok {
		t.Fatal("expected priority property")
	}
	if !priorityProp.HasEnum() {
		t.Error("priority property should have enum")
	}
	if len(priorityProp.Enum) != 4 {
		t.Errorf("priority enum length = %d, want 4", len(priorityProp.Enum))
	}

	// Test StatusRequest
	statusReq, ok := doc.Components.Schemas["StatusRequest"]
	if !ok {
		t.Fatal("expected StatusRequest schema")
	}
	if statusReq.Properties == nil {
		t.Fatal("expected StatusRequest to have properties")
	}

	newStatusProp, ok := statusReq.Properties["newStatus"]
	if !ok {
		t.Fatal("expected newStatus property")
	}
	if !newStatusProp.HasEnum() {
		t.Error("newStatus property should have enum")
	}
	expectedNewStatus := []interface{}{"active", "inactive", "deleted"}
	if len(newStatusProp.Enum) != len(expectedNewStatus) {
		t.Errorf("newStatus enum length = %d, want %d", len(newStatusProp.Enum), len(expectedNewStatus))
	}
	for i, v := range expectedNewStatus {
		if newStatusProp.Enum[i] != v {
			t.Errorf("newStatus enum[%d] = %v, want %v", i, newStatusProp.Enum[i], v)
		}
	}
}

func TestEnumEdgeCases(t *testing.T) {
	// Test 1: Schema with no type but with enum
	spec1 := `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths:
  /test:
    get:
      operationId: test1
      responses:
        "200":
          description: OK
components:
  schemas:
    NoTypeEnum:
      enum:
        - value1
        - value2
`
	tmpFile1 := filepath.Join(t.TempDir(), "edge1.yaml")
	if err := os.WriteFile(tmpFile1, []byte(spec1), 0644); err != nil {
		t.Fatal(err)
	}
	doc1, err := ParseOpenAPI(tmpFile1)
	if err != nil {
		t.Fatalf("ParseOpenAPI failed for spec1: %v", err)
	}
	noType, ok := doc1.Components.Schemas["NoTypeEnum"]
	if !ok {
		t.Fatal("expected NoTypeEnum schema")
	}
	if !noType.HasEnum() {
		t.Error("NoTypeEnum should have enum")
	}
	if len(noType.Enum) != 2 {
		t.Errorf("NoTypeEnum enum length = %d, want 2", len(noType.Enum))
	}

	// Test 2: Schema with enum and nullable
	spec2 := `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths:
  /test:
    get:
      operationId: test2
      responses:
        "200":
          description: OK
components:
  schemas:
    NullableEnum:
      type: string
      nullable: true
      enum:
        - a
        - b
`
	tmpFile2 := filepath.Join(t.TempDir(), "edge2.yaml")
	if err := os.WriteFile(tmpFile2, []byte(spec2), 0644); err != nil {
		t.Fatal(err)
	}
	doc2, err := ParseOpenAPI(tmpFile2)
	if err != nil {
		t.Fatalf("ParseOpenAPI failed for spec2: %v", err)
	}
	nullable, ok := doc2.Components.Schemas["NullableEnum"]
	if !ok {
		t.Fatal("expected NullableEnum schema")
	}
	if !nullable.HasEnum() {
		t.Error("NullableEnum should have enum")
	}
	if !nullable.Nullable {
		t.Error("NullableEnum should be nullable")
	}
	if len(nullable.Enum) != 2 {
		t.Errorf("NullableEnum enum length = %d, want 2", len(nullable.Enum))
	}

	// Test 3: Large enum with many values
	largeEnumValues := make([]interface{}, 100)
	largeEnumYaml := ""
	for i := 0; i < 100; i++ {
		largeEnumValues[i] = "value" + string(rune('a'+i%26))
		largeEnumYaml += fmt.Sprintf("        - value%s\n", string(rune('a'+i%26)))
	}
	spec3 := fmt.Sprintf(`openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths:
  /test:
    get:
      operationId: test3
      responses:
        "200":
          description: OK
components:
  schemas:
    LargeEnum:
      type: string
      enum:
%s`, largeEnumYaml)
	tmpFile3 := filepath.Join(t.TempDir(), "edge3.yaml")
	if err := os.WriteFile(tmpFile3, []byte(spec3), 0644); err != nil {
		t.Fatal(err)
	}
	doc3, err := ParseOpenAPI(tmpFile3)
	if err != nil {
		t.Fatalf("ParseOpenAPI failed for spec3: %v", err)
	}
	large, ok := doc3.Components.Schemas["LargeEnum"]
	if !ok {
		t.Fatal("expected LargeEnum schema")
	}
	if !large.HasEnum() {
		t.Error("LargeEnum should have enum")
	}
	if len(large.Enum) != 100 {
		t.Errorf("LargeEnum enum length = %d, want 100", len(large.Enum))
	}
}
