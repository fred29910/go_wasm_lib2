package generator

import (
	"testing"
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
