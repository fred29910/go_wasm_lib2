package generator

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

type Generator struct {
	ModuleName     string
	OutputModule   string
	Package        string
	RuntimePath    string
	RuntimeImport  string
}

func NewGenerator(moduleName, outputModule, packageName, runtimePath, runtimeImport string) *Generator {
	if moduleName == "" {
		moduleName = "github.com/fred29910/gowasm"
	}
	if outputModule == "" {
		outputModule = "generated-sdk"
	}
	if packageName == "" {
		packageName = "generated"
	}
	if runtimePath == "" {
		if cwd, err := os.Getwd(); err == nil {
			runtimePath = cwd
		}
	}
	if runtimeImport == "" {
		runtimeImport = "github.com/fred29910/gowasm/pkg/runtime"
	}
	return &Generator{ModuleName: moduleName, OutputModule: outputModule, Package: packageName, RuntimePath: runtimePath, RuntimeImport: runtimeImport}
}

func (g *Generator) Generate(specPath, outDir string) error {
	doc, err := ParseOpenAPI(specPath)
	if err != nil {
		return err
	}

	model := g.buildModel(doc)

	if err := g.writeGoClient(outDir, model); err != nil {
		return err
	}
	if err := g.writeTSClient(outDir, model); err != nil {
		return err
	}
	if err := g.writeDemoHTML(outDir, model); err != nil {
		return err
	}

	return nil
}

type GenerationModel struct {
	Doc           *OpenAPI
	Package       string
	ModuleName    string
	OutputModule  string
	RuntimePath   string
	RuntimeImport string
	InfoTitle     string
	InfoVersion   string
	BaseURL       string
	Schemas       []GeneratedSchema
	Operations    []GeneratedOperation
	OperationIDs  []string
}

func (g *Generator) buildModel(doc *OpenAPI) GenerationModel {
	// Convert schemas
	schemas := make([]GeneratedSchema, 0, len(doc.Components.Schemas))
	for _, name := range sortedMapKeys(doc.Components.Schemas) {
		schema := doc.Components.Schemas[name]
		goName := ToGoName(name)
		s := GeneratedSchema{
			Name:   name,
			GoName: goName,
			TSName: goName,
		}

		for _, prop := range sortedMapKeys(schema.Properties) {
			propSchema := schema.Properties[prop]
			required := contains(schema.Required, prop)
			s.Properties = append(s.Properties, GeneratedProperty{
				Name:     prop,
				GoName:   ToGoName(prop),
				TSName:   ToTSName(prop),
				GoType:   g.goType(&propSchema, required),
				TSType:   g.tsType(&propSchema, required),
				Required: required,
			})
		}
		schemas = append(schemas, s)
	}

	// Convert operations
	ops := make([]GeneratedOperation, 0, len(doc.Operations()))
	operationIDs := make([]string, 0, len(doc.Operations()))
	for _, op := range doc.Operations() {
		requestName := ToGoName(op.OperationID) + "Request"
		responseName := ToGoName(op.OperationID) + "Response"
		requestInterface := ToGoName(op.OperationID) + "Request"
		responseInterface := ToGoName(op.OperationID) + "Response"

		genOp := GeneratedOperation{
			ID:                  op.OperationID,
			TSName:              ToTSName(op.OperationID),
			Method:              op.Method,
			Path:                op.Path,
			Summary:             op.Summary,
			RequestStructName:   requestName,
			RequestInterface:    requestInterface,
			ResponseStructName:  responseName,
			ResponseInterface:   responseInterface,
			BodyParamName:       "body",
		}

		for _, p := range op.Parameters {
			param := GeneratedParameter{
				Name:     p.Name,
				GoName:   ToGoName(p.Name),
				TSName:   ToTSName(p.Name),
				GoType:   g.goType(p.Schema, p.Required),
				TSType:   g.tsType(p.Schema, p.Required),
				In:       p.In,
				Required: p.Required,
			}
			if p.In == "path" {
				genOp.PathParams = append(genOp.PathParams, param)
			} else if p.In == "query" {
				genOp.QueryParams = append(genOp.QueryParams, param)
			}
		}

		if op.RequestBody != nil {
			genOp.HasBody = true
			genOp.BodyParamName = "body"
			if media := firstJSONMedia(op.RequestBody.Content); media != nil && media.Schema != nil {
				genOp.BodyType = g.goType(media.Schema, op.RequestBody.Required)
			} else {
				genOp.BodyType = "interface{}"
			}
		}

		if resp := op.Responses["200"]; resp.Content != nil {
			if media := firstJSONMedia(resp.Content); media != nil && media.Schema != nil {
				genOp.ResponseType = g.goType(media.Schema, false)
			}
		}
		if genOp.ResponseType == "" {
			genOp.ResponseType = "interface{}"
		}

		ops = append(ops, genOp)
		operationIDs = append(operationIDs, op.OperationID)
	}

	sort.Slice(ops, func(i, j int) bool {
		return ops[i].ID < ops[j].ID
	})
	sort.Strings(operationIDs)

	title := fmt.Sprint(doc.Info["title"])
	version := fmt.Sprint(doc.Info["version"])
	if doc.Info == nil {
		title = "Generated API"
		version = "0.1.0"
	}

	return GenerationModel{
		Doc:           doc,
		Package:       g.Package,
		ModuleName:    g.ModuleName,
		OutputModule:  g.OutputModule,
		RuntimePath:   g.RuntimePath,
		RuntimeImport: g.RuntimeImport,
		InfoTitle:     title,
		InfoVersion:   version,
		BaseURL:       doc.DefaultServer(),
		Schemas:       schemas,
		Operations:    ops,
		OperationIDs:  operationIDs,
	}
}

func (g *Generator) goType(schema *Schema, required bool) string {
	if schema == nil {
		return "interface{}"
	}
	if schema.Ref != "" {
		if name, ok := g.docSchemaRef(schema.Ref); ok {
			return ToGoName(name)
		}
		return "interface{}"
	}
	if schema.Type == "array" {
		if schema.Items != nil {
			return "[]" + g.goType(schema.Items, false)
		}
		return "[]interface{}"
	}
	if schema.Type == "object" {
		if schema.AdditionalProperties != nil {
			return "map[string]" + g.goType(schema.AdditionalProperties, false)
		}
		return "map[string]interface{}"
	}

	switch schema.Type {
	case "integer":
		if schema.Format == "int64" {
			return "int64"
		}
		return "int"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	case "string":
		if schema.Format == "date-time" || schema.Format == "date" || schema.Format == "byte" {
			return "string"
		}
		return "string"
	default:
		return "interface{}"
	}
}

func (g *Generator) tsType(schema *Schema, required bool) string {
	if schema == nil {
		return "any"
	}
	if schema.Ref != "" {
		if name, ok := g.docSchemaRef(schema.Ref); ok {
			return ToGoName(name)
		}
		return "any"
	}
	if schema.Type == "array" {
		if schema.Items != nil {
			return "Array<" + g.tsType(schema.Items, false) + ">"
		}
		return "Array<any>"
	}
	if schema.Type == "object" {
		if schema.AdditionalProperties != nil {
			return "Record<string, " + g.tsType(schema.AdditionalProperties, false) + ">"
		}
		return "Record<string, any>"
	}
	switch schema.Type {
	case "integer":
		return "number"
	case "number":
		return "number"
	case "boolean":
		return "boolean"
	case "string":
		return "string"
	default:
		return "any"
	}
}

func (g *Generator) docSchemaRef(ref string) (string, bool) {
	prefix := "#/components/schemas/"
	if !strings.HasPrefix(ref, prefix) {
		return "", false
	}
	return ref[len(prefix):], true
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

func sortedMapKeys(m map[string]Schema) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func firstJSONMedia(content map[string]Media) *Media {
	if len(content) == 0 {
		return nil
	}
	if media, ok := content["application/json"]; ok {
		return &media
	}
	for _, media := range content {
		return &media
	}
	return nil
}

func generateOperationID(method, path string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	parts := re.Split(path, -1)
	var b strings.Builder
	b.WriteString(strings.ToLower(method))
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]))
		if len(p) > 1 {
			b.WriteString(p[1:])
		}
	}
	return b.String()
}