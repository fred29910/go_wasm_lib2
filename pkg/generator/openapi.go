package generator

import (
	"fmt"
	"os"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
)

// OpenAPI represents a minimal OpenAPI 3.x document used as the internal
// generation model. It is populated from a fully-resolved kin-openapi AST.
type OpenAPI struct {
	OpenAPI      string              `yaml:"openapi"`
	Info         map[string]interface{} `yaml:"info"`
	Servers      []Server            `yaml:"servers"`
	Paths        map[string]PathItem `yaml:"paths"`
	Components   Components          `yaml:"components"`
	OperationIDs map[string]bool     `yaml:"-"`
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
	OperationID string              `yaml:"operationId"`
	Summary     string              `yaml:"summary"`
	Tags        []string            `yaml:"tags"`
	Parameters  []Parameter         `yaml:"parameters"`
	RequestBody *RequestBody        `yaml:"requestBody"`
	Responses   map[string]Response `yaml:"responses"`
	Method      string              `yaml:"-"`
	Path        string              `yaml:"-"`
}

type Parameter struct {
	Name     string  `yaml:"name"`
	In       string  `yaml:"in"`
	Required bool    `yaml:"required"`
	Schema   *Schema `yaml:"schema"`
}

type RequestBody struct {
	Required bool             `yaml:"required"`
	Content  map[string]Media `yaml:"content"`
}

type Response struct {
	Description string            `yaml:"description"`
	Content     map[string]Media  `yaml:"content"`
}

type Media struct {
	Schema *Schema `yaml:"schema"`
}

// Schema is the internal generation model for an OpenAPI schema.
// It is produced from a fully-resolved kin-openapi SchemaRef (all $ref expanded).
type Schema struct {
	Type                 string            `yaml:"type"`
	Format               string            `yaml:"format"`
	Ref                  string            `yaml:"$ref"`
	Items                *Schema           `yaml:"items"`
	Properties           map[string]Schema `yaml:"properties"`
	Required             []string          `yaml:"required"`
	Enum                 []interface{}     `yaml:"enum"`
	Nullable             bool              `yaml:"nullable"`
	AdditionalProperties *Schema           `yaml:"additionalProperties"`
	OriginalName         string            `yaml:"-"`
}

type Components struct {
	Schemas map[string]Schema `yaml:"schemas"`
}

// maxSpecFileSize limits OpenAPI spec file reads to 50 MB.
const maxSpecFileSize = 50 << 20

// ParseOpenAPI parses an OpenAPI YAML or JSON file using kin-openapi for
// loading, validation, and full $ref resolution (including external files
// and nested references). The resolved AST is then converted to the internal
// generation model.
func ParseOpenAPI(path string) (*OpenAPI, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat openapi file: %w", err)
	}
	if fi.Size() > maxSpecFileSize {
		return nil, fmt.Errorf("openapi file too large: %d bytes (max %d MB)", fi.Size(), maxSpecFileSize>>20)
	}

	// Use kin-openapi Loader: it validates the spec and resolves all $ref.
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	spec, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("load openapi spec: %w", err)
	}

	// Validate the spec (structural checks beyond what LoadFromFile does).
	if err := spec.Validate(loader.Context); err != nil {
		return nil, fmt.Errorf("validate openapi spec: %w", err)
	}

	return convertFromKinOpenAPI(spec)
}

// convertFromKinOpenAPI converts a fully-resolved kin-openapi *T into the
// internal OpenAPI generation model. All $ref references are resolved by the
// kin-openapi Loader, so Schema.Ref will be empty and schemas are fully expanded.
func convertFromKinOpenAPI(spec *openapi3.T) (*OpenAPI, error) {
	doc := &OpenAPI{
		OpenAPI:      spec.OpenAPI,
		OperationIDs: make(map[string]bool),
	}

	// Info
	if spec.Info != nil {
		doc.Info = map[string]interface{}{
			"title":       spec.Info.Title,
			"version":     spec.Info.Version,
			"description": spec.Info.Description,
		}
	}

	// Servers
	for _, s := range spec.Servers {
		doc.Servers = append(doc.Servers, Server{URL: s.URL})
	}

	// Components / Schemas
	doc.Components = Components{
		Schemas: make(map[string]Schema),
	}
	if spec.Components != nil && spec.Components.Schemas != nil {
		for name, schemaRef := range spec.Components.Schemas {
			s := convertSchemaRef(schemaRef)
			s.OriginalName = name
			doc.Components.Schemas[name] = s
		}
	}

	// Paths — convert each operation and populate PathItem.
	doc.Paths = make(map[string]PathItem)
	if spec.Paths != nil {
		for pathStr, pathItem := range spec.Paths.Map() {
			if pathItem == nil {
				continue
			}
			pi := PathItem{}
			methods := []struct {
				op    **Operation
				name  string
				getOp func() *openapi3.Operation
			}{
				{&pi.Get, "GET", func() *openapi3.Operation { return pathItem.Get }},
				{&pi.Post, "POST", func() *openapi3.Operation { return pathItem.Post }},
				{&pi.Put, "PUT", func() *openapi3.Operation { return pathItem.Put }},
				{&pi.Delete, "DELETE", func() *openapi3.Operation { return pathItem.Delete }},
				{&pi.Patch, "PATCH", func() *openapi3.Operation { return pathItem.Patch }},
				{&pi.Options, "OPTIONS", func() *openapi3.Operation { return pathItem.Options }},
				{&pi.Head, "HEAD", func() *openapi3.Operation { return pathItem.Head }},
			}
			for _, m := range methods {
				if kinOp := m.getOp(); kinOp != nil {
					op := convertOperation(kinOp, m.name, pathStr)
					*m.op = &op
					if doc.OperationIDs[op.OperationID] {
						return nil, fmt.Errorf("duplicate operationId: %s", op.OperationID)
					}
					doc.OperationIDs[op.OperationID] = true
				}
			}
			doc.Paths[pathStr] = pi
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

	return doc, nil
}

// convertOperation converts a kin-openapi Operation to the internal Operation model.
// All $ref in schemas are already resolved by the kin-openapi Loader.
func convertOperation(kinOp *openapi3.Operation, method, path string) Operation {
	opID := kinOp.OperationID
	if opID == "" {
		opID = generateOperationID(method, path)
	}

	op := Operation{
		OperationID: opID,
		Summary:     kinOp.Summary,
		Method:      method,
		Path:        path,
		Parameters:  make([]Parameter, 0, len(kinOp.Parameters)),
		Responses:   make(map[string]Response),
	}

	// Parameters
	for _, p := range kinOp.Parameters {
		if p == nil || p.Value == nil {
			continue
		}
		param := Parameter{
			Name:     p.Value.Name,
			In:       p.Value.In,
			Required: p.Value.Required,
		}
		if p.Value.Schema != nil && p.Value.Schema.Value != nil {
			s := convertSchemaRef(p.Value.Schema)
			param.Schema = &s
		}
		// Parameters without a schema (e.g. pure $ref to a parameter object)
		// are kept with a nil Schema; buildOperation must guard against this.
		op.Parameters = append(op.Parameters, param)
	}

	// Request body
	if kinOp.RequestBody != nil && kinOp.RequestBody.Value != nil {
		rb := kinOp.RequestBody.Value
		rbReq := RequestBody{
			Required: rb.Required,
			Content:  make(map[string]Media),
		}
		for mediaType, mediaObj := range rb.Content {
			m := Media{}
			if mediaObj.Schema != nil && mediaObj.Schema.Value != nil {
				s := convertSchemaRef(mediaObj.Schema)
				m.Schema = &s
			}
			rbReq.Content[mediaType] = m
		}
		op.RequestBody = &rbReq
	}

	// Responses
	if kinOp.Responses != nil {
		for code, respRef := range kinOp.Responses.Map() {
			if respRef == nil || respRef.Value == nil {
				continue
			}
			resp := respRef.Value
			desc := ""
			if resp.Description != nil {
				desc = *resp.Description
			}
			r := Response{
				Description: desc,
				Content:     make(map[string]Media),
			}
			for mediaType, mediaObj := range resp.Content {
				m := Media{}
				if mediaObj.Schema != nil && mediaObj.Schema.Value != nil {
					s := convertSchemaRef(mediaObj.Schema)
					m.Schema = &s
				}
				r.Content[mediaType] = m
			}
			op.Responses[code] = r
		}
	}

	return op
}

// convertSchemaRef converts a kin-openapi SchemaRef to the internal Schema.
// All $ref should already be resolved by the Loader, so Ref will typically
// be empty and Value will be populated.
func convertSchemaRef(ref *openapi3.SchemaRef) Schema {
	if ref == nil {
		return Schema{}
	}
	if ref.Value == nil {
		// Unresolved ref — keep the ref string for goType/tsType to handle.
		return Schema{Ref: ref.Ref}
	}
	s := ref.Value
	schema := Schema{
		Format:   s.Format,
		Nullable: s.Nullable,
	}

	// Enum — normalise JSON numbers to Go-native types so that
	// integer enums come through as int, not float64.
	if len(s.Enum) > 0 {
		schema.Enum = make([]interface{}, len(s.Enum))
		for i, v := range s.Enum {
			schema.Enum[i] = normaliseEnumValue(v)
		}
	}

	// Type — kin-openapi uses *Types (slice); take the first.
	if s.Type != nil && len(s.Type.Slice()) > 0 {
		schema.Type = s.Type.Slice()[0]
	}

	// $ref (should be empty after Loader resolution, but keep for safety)
	schema.Ref = ref.Ref

	// Items
	if s.Items != nil && s.Items.Value != nil {
		items := convertSchemaRef(s.Items)
		schema.Items = &items
	}

	// Properties
	if len(s.Properties) > 0 {
		schema.Properties = make(map[string]Schema)
		for name, propRef := range s.Properties {
			schema.Properties[name] = convertSchemaRef(propRef)
		}
	}

	// Required
	schema.Required = s.Required

	// AdditionalProperties
	if s.AdditionalProperties.Schema != nil {
		ap := convertSchemaRef(s.AdditionalProperties.Schema)
		schema.AdditionalProperties = &ap
	}

	return schema
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

// normaliseEnumValue converts JSON-decoded numbers to Go-native types.
// JSON numbers arrive as float64; whole numbers are converted to int.
func normaliseEnumValue(v interface{}) interface{} {
	if f, ok := v.(float64); ok {
		if f == float64(int64(f)) {
			return int(f)
		}
	}
	return v
}
