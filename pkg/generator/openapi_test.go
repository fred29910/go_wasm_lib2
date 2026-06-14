package generator

import (
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
