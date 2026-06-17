package generator

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

//go:embed templates/sdk.go.tmpl
var goClientTmpl string

//go:embed templates/go.mod.tmpl
var goModTmpl string

//go:embed templates/main.go.tmpl
var goMainTmpl string

// Pre-compiled templates (initialized once via sync.Once)
var (
	goClientTemplate     *template.Template
	goModTemplate        *template.Template
	goMainTemplate       *template.Template
	goClientTemplateOnce sync.Once
	goModTemplateOnce    sync.Once
	goMainTemplateOnce   sync.Once
)

// getGoClientTemplate returns the pre-compiled Go client template.
func getGoClientTemplate() *template.Template {
	goClientTemplateOnce.Do(func() {
		funcMap := template.FuncMap{
			"hasPrefix": func(s, prefix string) bool {
				return len(s) >= len(prefix) && s[:len(prefix)] == prefix
			},
			"trimPrefix": func(s, prefix string) string {
				if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
					return s[len(prefix):]
				}
				return s
			},
		}
		goClientTemplate = template.Must(template.New("generated.go").Funcs(funcMap).Parse(goClientTmpl))
	})
	return goClientTemplate
}

// getGoModTemplate returns the pre-compiled go.mod template.
func getGoModTemplate() *template.Template {
	goModTemplateOnce.Do(func() {
		goModTemplate = template.Must(template.New("go.mod").Parse(goModTmpl))
	})
	return goModTemplate
}

// getGoMainTemplate returns the pre-compiled main.go template.
func getGoMainTemplate() *template.Template {
	goMainTemplateOnce.Do(func() {
		goMainTemplate = template.Must(template.New("main.go").Parse(goMainTmpl))
	})
	return goMainTemplate
}

type goTemplateData struct {
	GenerationModel
}

// renderGoTemplate sets up the Go client template with custom template detection,
// shared by writeGoClient and renderGoClient to eliminate duplication.
// Uses pre-compiled template for the default case, or parses custom template if specified.
func (g *Generator) renderGoTemplate(model GenerationModel) (*template.Template, *goTemplateData, error) {
	data := goTemplateData{GenerationModel: model}
	if g.config.GoTemplatePath != "" {
		content, err := os.ReadFile(g.config.GoTemplatePath)
		if err != nil {
			return nil, nil, fmt.Errorf("read custom Go template: %w", err)
		}
		funcMap := template.FuncMap{
			"hasPrefix": func(s, prefix string) bool {
				return len(s) >= len(prefix) && s[:len(prefix)] == prefix
			},
			"trimPrefix": func(s, prefix string) string {
				if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
					return s[len(prefix):]
				}
				return s
			},
		}
		tmpl := template.Must(template.New("generated.go").Funcs(funcMap).Parse(string(content)))
		return tmpl, &data, nil
	}
	return getGoClientTemplate(), &data, nil
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
	tmpl := getGoModTemplate()
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
	tmpl := getGoMainTemplate()
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

// renderGoMod renders the go.mod template to a string without writing to disk.
func (g *Generator) renderGoMod(model GenerationModel) (string, error) {
	tmpl := getGoModTemplate().Option("missingkey=zero")
	var buf strings.Builder
	if err := tmpl.Execute(&buf, model); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// renderGoMain renders the cmd/wasm/main.go template to a string without writing to disk.
func (g *Generator) renderGoMain(model GenerationModel) (string, error) {
	tmpl := getGoMainTemplate().Option("missingkey=zero")
	var buf strings.Builder
	if err := tmpl.Execute(&buf, model); err != nil {
		return "", err
	}
	return buf.String(), nil
}
