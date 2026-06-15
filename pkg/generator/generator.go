package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Config holds the configuration for code generation.
// Use NewConfig to create a Config with sensible defaults.
type Config struct {
	ModuleName     string // Go module name for generated code
	OutputModule   string // Output module path
	Package        string // Go package name
	RuntimePath    string // Local path to the runtime module (for go.mod replace)
	RuntimeImport  string // Import path of the runtime package
	Validation     bool   // Generate validation methods for request structs
	GoTemplatePath string // Custom Go template file path (optional)
	TSTemplatePath string // Custom TypeScript template file path (optional)
	Verbose        bool   // Show detailed progress during generation
}

// NewConfig creates a Config with default values.
func NewConfig() *Config {
	cwd, _ := os.Getwd()
	return &Config{
		ModuleName:    "github.com/fred29910/gowasm",
		OutputModule:  "generated-sdk",
		Package:       "generated",
		RuntimePath:   cwd,
		RuntimeImport: "github.com/fred29910/gowasm/pkg/runtime",
		Validation:    true,
	}
}

// Generator generates Go and TypeScript SDK from OpenAPI specs.
type Generator struct {
	config  *Config
	verbose bool
}

// NewGenerator creates a Generator with the given parameters.
// Empty strings fall back to defaults.
func NewGenerator(moduleName, outputModule, packageName, runtimePath, runtimeImport string) *Generator {
	cfg := NewConfig()
	if moduleName != "" {
		cfg.ModuleName = moduleName
	}
	if outputModule != "" {
		cfg.OutputModule = outputModule
	}
	if packageName != "" {
		cfg.Package = packageName
	}
	if runtimePath != "" {
		cfg.RuntimePath = runtimePath
	}
	if runtimeImport != "" {
		cfg.RuntimeImport = runtimeImport
	}
	return &Generator{config: cfg}
}

// NewGeneratorFromConfig creates a Generator from a Config.
func NewGeneratorFromConfig(cfg *Config) *Generator {
	if cfg == nil {
		cfg = NewConfig()
	}
	return &Generator{config: cfg, verbose: cfg.Verbose}
}

func (g *Generator) progress(msg string) {
	if g.verbose {
		fmt.Println(msg)
	}
}

func (g *Generator) Generate(specPath, outDir string) (*GenerationResult, error) {
	g.progress("Parsing spec...")
	doc, err := ParseOpenAPI(specPath)
	if err != nil {
		return nil, err
	}

	totalSchemas := len(doc.Components.Schemas)
	totalOps := len(doc.Operations())
	g.progress(fmt.Sprintf("Building model: %d schemas, %d operations...", totalSchemas, totalOps))
	model := g.buildModel(doc)

	g.progress("Rendering Go template...")
	if err := g.writeGoClient(outDir, model); err != nil {
		return nil, err
	}
	g.progress("Writing go.mod...")
	if err := g.writeGoMod(outDir, model); err != nil {
		return nil, err
	}
	g.progress("Writing main.go...")
	if err := g.writeGoMain(outDir, model); err != nil {
		return nil, err
	}
	g.progress("Rendering TypeScript template...")
	if err := g.writeTSClient(outDir, model); err != nil {
		return nil, err
	}
	g.progress("Writing files...")
	if err := g.writeDemoHTML(outDir, model); err != nil {
		return nil, err
	}

	// Collect file info
	var result GenerationResult
	files := []struct {
		path string
		name string
	}{
		{outDir, "generated.go"},
		{outDir, "go.mod"},
		{outDir, "main.go"},
		{outDir, "sdk.ts"},
		{outDir, "index.html"},
	}
	for _, f := range files {
		info, err := os.Stat(filepath.Join(f.path, f.name))
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", f.name, err)
		}
		result.Files = append(result.Files, GeneratedFile{
			Path: f.name,
			Size: int(info.Size()),
		})
		result.TotalSize += int(info.Size())
	}

	g.progress(fmt.Sprintf("Generated %d operations, %d schemas, %d files", totalOps, totalSchemas, len(result.Files)))
	return &result, nil
}

// DryRunFile represents a file that would be generated in dry-run mode.
type DryRunFile struct {
	Path string // Relative path of the file
	Size int    // Size of the rendered content in bytes
}

// DryRunResult holds the result of a dry-run generation.
type DryRunResult struct {
	Files     []DryRunFile // List of files that would be generated
	TotalSize int          // Total size of all files in bytes
}

// GeneratedFile represents a file that was generated.
type GeneratedFile struct {
	Path string // Relative path of the file
	Size int    // Size of the file in bytes
}

// GenerationResult holds the result of a generation.
type GenerationResult struct {
	Files     []GeneratedFile // List of files that were generated
	TotalSize int             // Total size of all files in bytes
}

// GenerateDryRun parses the spec, builds the model, renders templates in memory,
// and returns what would be generated without writing any files.
func (g *Generator) GenerateDryRun(specPath, outDir string) (*DryRunResult, error) {
	g.progress("Parsing spec...")
	doc, err := ParseOpenAPI(specPath)
	if err != nil {
		return nil, err
	}

	totalSchemas := len(doc.Components.Schemas)
	totalOps := len(doc.Operations())
	g.progress(fmt.Sprintf("Building model: %d schemas, %d operations...", totalSchemas, totalOps))
	model := g.buildModel(doc)

	var result DryRunResult

	// Render Go client
	g.progress("Rendering Go template...")
	goContent, err := g.renderGoClient(model)
	if err != nil {
		return nil, fmt.Errorf("render Go client: %w", err)
	}
	result.Files = append(result.Files, DryRunFile{
		Path: "generated.go",
		Size: len(goContent),
	})

	// Render TypeScript client
	g.progress("Rendering TypeScript template...")
	tsContent, err := g.renderTSClient(model)
	if err != nil {
		return nil, fmt.Errorf("render TypeScript client: %w", err)
	}
	result.Files = append(result.Files, DryRunFile{
		Path: "sdk.ts",
		Size: len(tsContent),
	})

	// Render demo HTML
	g.progress("Writing files...")
	htmlContent, err := g.renderDemoHTML(model)
	if err != nil {
		return nil, fmt.Errorf("render demo HTML: %w", err)
	}
	result.Files = append(result.Files, DryRunFile{
		Path: "index.html",
		Size: len(htmlContent),
	})

	// Calculate total size
	for _, file := range result.Files {
		result.TotalSize += file.Size
	}

	g.progress(fmt.Sprintf("Generated %d operations, %d schemas, %d files", totalOps, totalSchemas, len(result.Files)))
	return &result, nil
}

type GenerationModel struct {
	Doc           *OpenAPI
	Config        *Config
	InfoTitle     string
	InfoVersion   string
	BaseURL       string
	Schemas       []GeneratedSchema
	Operations    []GeneratedOperation
	OperationIDs  []string
	Validation    bool
}

func (g *Generator) buildModel(doc *OpenAPI) GenerationModel {
	schemas := g.buildSchemas(doc)
	ops, operationIDs := g.buildOperations(doc)
	title, version := g.extractInfo(doc)

	return GenerationModel{
		Doc:           doc,
		Config:        g.config,
		InfoTitle:     title,
		InfoVersion:   version,
		BaseURL:       doc.DefaultServer(),
		Schemas:       schemas,
		Operations:    ops,
		OperationIDs:  operationIDs,
		Validation:    g.config.Validation,
	}
}

func (g *Generator) buildSchemas(doc *OpenAPI) []GeneratedSchema {
	schemas := make([]GeneratedSchema, 0, len(doc.Components.Schemas))
	for _, name := range sortedSchemaMapKeys(doc.Components.Schemas) {
		schema := doc.Components.Schemas[name]
		goName := ToGoName(name)
		s := GeneratedSchema{
			Name:   name,
			GoName: goName,
			TSName: goName,
		}
		for _, prop := range sortedSchemaMapKeys(schema.Properties) {
			propSchema := schema.Properties[prop]
			required := contains(schema.Required, prop)
			s.Properties = append(s.Properties, GeneratedProperty{
				Name:       prop,
				GoName:     ToGoName(prop),
				TSName:     ToTSName(prop),
				GoType:     g.goType(&propSchema, required),
				TSType:     g.tsType(&propSchema, required),
				Required:   required,
				EnumValues: propSchema.Enum,
				Format:     propSchema.Format,
			})
		}
		schemas = append(schemas, s)
	}
	return schemas
}

func (g *Generator) buildOperations(doc *OpenAPI) ([]GeneratedOperation, []string) {
	ops := make([]GeneratedOperation, 0, len(doc.Operations()))
	operationIDs := make([]string, 0, len(doc.Operations()))
	for _, op := range doc.Operations() {
		genOp := g.buildOperation(op)
		ops = append(ops, genOp)
		operationIDs = append(operationIDs, op.OperationID)
	}
	sort.Slice(ops, func(i, j int) bool {
		return ops[i].ID < ops[j].ID
	})
	sort.Strings(operationIDs)
	return ops, operationIDs
}

func (g *Generator) buildOperation(op *Operation) GeneratedOperation {
	requestInterface := ToGoName(op.OperationID) + "Request"

	genOp := GeneratedOperation{
		ID:                  op.OperationID,
		TSName:              ToTSName(op.OperationID),
		Method:              op.Method,
		Path:                op.Path,
		Summary:             op.Summary,
		RequestStructName:   requestInterface,
		RequestInterface:    requestInterface,
		ResponseStructName:  ToGoName(op.OperationID) + "Response",
		ResponseInterface:   ToGoName(op.OperationID) + "Response",
		BodyParamName:       "body",
	}

	for _, p := range op.Parameters {
		param := GeneratedParameter{
			Name:       p.Name,
			GoName:     ToGoName(p.Name),
			TSName:     ToTSName(p.Name),
			GoType:     g.goType(p.Schema, p.Required),
			TSType:     g.tsType(p.Schema, p.Required),
			In:         p.In,
			Required:   p.Required,
			EnumValues: p.Schema.Enum,
			Format:     p.Schema.Format,
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

	// Capture all responses - first pass: find primary (2xx or first)
	var primaryCode string
	for code := range op.Responses {
		if code >= "200" && code < "300" {
			primaryCode = code
			break
		}
	}
	// Fallback: use first response if no 2xx found
	if primaryCode == "" {
		for code := range op.Responses {
			primaryCode = code
			break
		}
	}

	// Second pass: build response list
	for code, resp := range op.Responses {
		var goType, tsType string
		if resp.Content != nil {
			if media := firstJSONMedia(resp.Content); media != nil && media.Schema != nil {
				goType = g.goType(media.Schema, false)
				tsType = g.tsType(media.Schema, false)
			}
		}
		if goType == "" {
			goType = "interface{}"
		}
		if tsType == "" {
			tsType = "any"
		}

		// Determine struct name: primary uses ResponseStructName, others use {OperationId}{Code}Response
		var structName string
		if code == primaryCode {
			structName = genOp.ResponseStructName
		} else {
			structName = ToGoName(genOp.ID) + code + "Response"
		}

		genResp := GeneratedResponse{
			Code:        code,
			GoType:      goType,
			TSType:      tsType,
			Description: resp.Description,
			Primary:     code == primaryCode,
			StructName:  structName,
		}
		genOp.Responses = append(genOp.Responses, genResp)

		// Set primary response type
		if code == primaryCode {
			genOp.ResponseType = goType
		}
	}

	if genOp.ResponseType == "" {
		genOp.ResponseType = "interface{}"
	}

	return genOp
}

func (g *Generator) extractInfo(doc *OpenAPI) (title, version string) {
	if doc.Info == nil {
		return "Generated API", "0.1.0"
	}
	return fmt.Sprint(doc.Info["title"]), fmt.Sprint(doc.Info["version"])
}

// sortedSchemaMapKeys returns sorted keys from a Schema map.
func sortedSchemaMapKeys(m map[string]Schema) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
	if schema.HasEnum() {
		return g.tsUnionType(schema.Enum)
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

func (g *Generator) tsUnionType(values []interface{}) string {
	var parts []string
	for _, v := range values {
		switch val := v.(type) {
		case string:
			parts = append(parts, fmt.Sprintf("'%s'", val))
		default:
			parts = append(parts, fmt.Sprintf("%v", val))
		}
	}
	return strings.Join(parts, " | ")
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