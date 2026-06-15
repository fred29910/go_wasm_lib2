package generator

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// OpenAPI represents a minimal OpenAPI 3.x document
type OpenAPI struct {
	OpenAPI     string                 `yaml:"openapi"`
	Info        map[string]interface{} `yaml:"info"`
	Servers     []Server               `yaml:"servers"`
	Paths       map[string]PathItem    `yaml:"paths"`
	Components  Components             `yaml:"components"`
	OperationIDs map[string]bool       `yaml:"-"`
}

type Server struct {
	URL string `yaml:"url"`
}

type PathItem struct {
	Get     *Operation `yaml:"get"`
	Post    *Operation `yaml:"post"`
	Put     *Operation `yaml:"put"`
	Delete  *Operation `yaml:"delete"`
	Patch   *Operation `yaml:"patch"`
	Options *Operation `yaml:"options"`
	Head    *Operation `yaml:"head"`
}

type Operation struct {
	OperationID string                 `yaml:"operationId"`
	Summary     string                 `yaml:"summary"`
	Tags        []string               `yaml:"tags"`
	Parameters  []Parameter            `yaml:"parameters"`
	RequestBody *RequestBody           `yaml:"requestBody"`
	Responses   map[string]Response    `yaml:"responses"`
	Method      string                 `yaml:"-"`
	Path        string                 `yaml:"-"`
}

type Parameter struct {
	Name     string      `yaml:"name"`
	In       string      `yaml:"in"`
	Required bool        `yaml:"required"`
	Schema   *Schema     `yaml:"schema"`
}

type RequestBody struct {
	Required  bool               `yaml:"required"`
	Content   map[string]Media   `yaml:"content"`
}

type Response struct {
	Description string             `yaml:"description"`
	Content     map[string]Media   `yaml:"content"`
}

type Media struct {
	Schema *Schema `yaml:"schema"`
}

type Schema struct {
	Type                 string             `yaml:"type"`
	Format               string             `yaml:"format"`
	Ref                  string             `yaml:"$ref"`
	Items                *Schema            `yaml:"items"`
	Properties           map[string]Schema  `yaml:"properties"`
	Required             []string           `yaml:"required"`
	Enum                 []interface{}      `yaml:"enum"`
	Nullable             bool               `yaml:"nullable"`
	AdditionalProperties *Schema            `yaml:"additionalProperties"`
	OriginalName         string             `yaml:"-"`
}

type Components struct {
	Schemas map[string]Schema `yaml:"schemas"`
}

// ParseOpenAPI parses an OpenAPI YAML file
func ParseOpenAPI(path string) (*OpenAPI, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read openapi file: %w", err)
	}

	var doc OpenAPI
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse openapi yaml: %w", err)
	}

	if doc.OpenAPI == "" {
		return nil, fmt.Errorf("missing openapi field")
	}

	doc.OperationIDs = make(map[string]bool)

	// Normalize operations
	for path, item := range doc.Paths {
		ops := []*Operation{
			item.Get, item.Post, item.Put, item.Delete, item.Patch, item.Options, item.Head,
		}
		methods := []string{"get", "post", "put", "delete", "patch", "options", "head"}

		for i, op := range ops {
			if op == nil {
				continue
			}
			op.Method = strings.ToUpper(methods[i])
			op.Path = path
			if op.OperationID == "" {
				op.OperationID = generateOperationID(op.Method, path)
			}
			if doc.OperationIDs[op.OperationID] {
				return nil, fmt.Errorf("duplicate operationId: %s", op.OperationID)
			}
			doc.OperationIDs[op.OperationID] = true
		}
	}

	// Sort schemas for deterministic output
	if doc.Components.Schemas != nil {
		keys := make([]string, 0, len(doc.Components.Schemas))
		for k := range doc.Components.Schemas {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sorted := make(map[string]Schema)
		for _, k := range keys {
			sorted[k] = doc.Components.Schemas[k]
		}
		doc.Components.Schemas = sorted
	}

	return &doc, nil
}

// Operations returns all operations in deterministic order
func (o *OpenAPI) Operations() []*Operation {
	var ops []*Operation
	paths := make([]string, 0, len(o.Paths))
	for p := range o.Paths {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, p := range paths {
		item := o.Paths[p]
		if item.Get != nil {
			ops = append(ops, item.Get)
		}
		if item.Post != nil {
			ops = append(ops, item.Post)
		}
		if item.Put != nil {
			ops = append(ops, item.Put)
		}
		if item.Delete != nil {
			ops = append(ops, item.Delete)
		}
		if item.Patch != nil {
			ops = append(ops, item.Patch)
		}
		if item.Options != nil {
			ops = append(ops, item.Options)
		}
		if item.Head != nil {
			ops = append(ops, item.Head)
		}
	}
	return ops
}

// HasEnum returns true if the schema has enum values defined
func (s *Schema) HasEnum() bool {
	return len(s.Enum) > 0
}

// DefaultServer returns the first server URL or empty string
func (o *OpenAPI) DefaultServer() string {
	if len(o.Servers) == 0 {
		return ""
	}
	return o.Servers[0].URL
}

