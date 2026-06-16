package generator

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/sdk.go.tmpl
var goClientTmpl string

//go:embed templates/go.mod.tmpl
var goModTmpl string

//go:embed templates/main.go.tmpl
var goMainTmpl string

type goTemplateData struct {
	GenerationModel
}

// renderGoTemplate sets up the Go client template with custom template detection,
// shared by writeGoClient and renderGoClient to eliminate duplication.
func (g *Generator) renderGoTemplate(model GenerationModel) (*template.Template, *goTemplateData, error) {
	data := goTemplateData{GenerationModel: model}
	funcMap := template.FuncMap{
		"hasPrefix": func(s, prefix string) bool {
			return len(s) >= len(prefix) && s[:len(prefix)] == prefix
		},
	}
	var tmplContent string
	if g.config.GoTemplatePath != "" {
		content, err := os.ReadFile(g.config.GoTemplatePath)
		if err != nil {
			return nil, nil, fmt.Errorf("read custom Go template: %w", err)
		}
		tmplContent = string(content)
	} else {
		tmplContent = goClientTmpl
	}
	tmpl := template.Must(template.New("generated.go").Funcs(funcMap).Parse(tmplContent))
	return tmpl, &data, nil
}

func (g *Generator) writeGoClient(outDir string, model GenerationModel) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	tmpl, data, err := g.renderGoTemplate(model)
	if err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(outDir, "generated.go"))
	if err != nil {
		return err
	}
	defer file.Close()
	return tmpl.Execute(file, data)
}

func (g *Generator) writeGoMod(outDir string, model GenerationModel) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	tmpl := template.Must(template.New("go.mod").Parse(goModTmpl))
	file, err := os.Create(filepath.Join(outDir, "go.mod"))
	if err != nil {
		return err
	}
	defer file.Close()
	return tmpl.Execute(file, model)
}

func (g *Generator) writeGoMain(outDir string, model GenerationModel) error {
	wasmDir := filepath.Join(outDir, "cmd", "wasm")
	if err := os.MkdirAll(wasmDir, 0755); err != nil {
		return err
	}
	tmpl := template.Must(template.New("main.go").Parse(goMainTmpl))
	file, err := os.Create(filepath.Join(wasmDir, "main.go"))
	if err != nil {
		return err
	}
	defer file.Close()
	return tmpl.Execute(file, model)
}

// renderGoClient renders the Go client template to a string without writing to disk.
func (g *Generator) renderGoClient(model GenerationModel) (string, error) {
	tmpl, data, err := g.renderGoTemplate(model)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
