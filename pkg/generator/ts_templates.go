package generator

import (
	_ "embed"
	"fmt"
	ht "html/template"
	"os"
	"path/filepath"
	"strings"
	tt "text/template"
	"sync"
	"unicode"
)

//go:embed templates/sdk.ts.tmpl
var tsClientTmpl string

//go:embed templates/index.html.tmpl
var demoHTMLTmpl string

// Pre-compiled templates (initialized once via sync.Once)
var (
	tsClientTemplate     *tt.Template
	demoHTMLTemplate     *tt.Template
	tsClientTemplateOnce sync.Once
	demoHTMLTemplateOnce sync.Once
)

// getTSClientTemplate returns the pre-compiled TypeScript client template.
func getTSClientTemplate() *tt.Template {
	tsClientTemplateOnce.Do(func() {
		tsClientTemplate = tt.Must(tt.New("sdk.ts").Parse(tsClientTmpl))
	})
	return tsClientTemplate
}

// getDemoHTMLTemplate returns the pre-compiled demo HTML template.
func getDemoHTMLTemplate() *tt.Template {
	demoHTMLTemplateOnce.Do(func() {
		demoHTMLTemplate = tt.Must(tt.New("index.html").Funcs(tt.FuncMap{
			"htmlEscape": func(s string) string { return ht.HTMLEscapeString(s) },
			"toLower": func(s string) string {
				var b strings.Builder
				b.Grow(len(s))
				for _, r := range s {
					b.WriteRune(unicode.ToLower(r))
				}
				return b.String()
			},
		}).Parse(demoHTMLTmpl))
	})
	return demoHTMLTemplate
}

// renderTSTemplate sets up the TypeScript client template with custom template detection,
// shared by writeTSClient and renderTSClient to eliminate duplication.
func (g *Generator) renderTSTemplate(model GenerationModel) (*tt.Template, GenerationModel, error) {
	if g.config.TSTemplatePath != "" {
		content, err := os.ReadFile(g.config.TSTemplatePath)
		if err != nil {
			return nil, model, fmt.Errorf("read custom TypeScript template: %w", err)
		}
		tmpl := tt.Must(tt.New("sdk.ts").Parse(string(content)))
		return tmpl, model, nil
	}
	return getTSClientTemplate(), model, nil
}

func (g *Generator) writeTSClient(outDir string, model GenerationModel) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	tmpl, model, err := g.renderTSTemplate(model)
	if err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(outDir, "sdk.ts"))
	if err != nil {
		return err
	}
	defer file.Close()
	return tmpl.Execute(file, model)
}

// renderDemoHTMLTemplate returns the pre-compiled demo HTML template.
func (g *Generator) renderDemoHTMLTemplate(model GenerationModel) (*tt.Template, GenerationModel, error) {
	return getDemoHTMLTemplate(), model, nil
}

func (g *Generator) writeDemoHTML(outDir string, model GenerationModel) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	tmpl, model, err := g.renderDemoHTMLTemplate(model)
	if err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(outDir, "index.html"))
	if err != nil {
		return err
	}
	defer file.Close()
	return tmpl.Execute(file, model)
}

// renderTSClient renders the TypeScript client template to a string without writing to disk.
func (g *Generator) renderTSClient(model GenerationModel) (string, error) {
	tmpl, model, err := g.renderTSTemplate(model)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, model); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// renderDemoHTML renders the demo HTML template to a string without writing to disk.
func (g *Generator) renderDemoHTML(model GenerationModel) (string, error) {
	tmpl, model, err := g.renderDemoHTMLTemplate(model)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, model); err != nil {
		return "", err
	}
	return buf.String(), nil
}
