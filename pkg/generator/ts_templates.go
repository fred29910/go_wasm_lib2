package generator

import (
	_ "embed"
	"fmt"
	ht "html/template"
	"os"
	"path/filepath"
	"strings"
	tt "text/template"
	"unicode"
)

//go:embed templates/sdk.ts.tmpl
var tsClientTmpl string

//go:embed templates/index.html.tmpl
var demoHTMLTmpl string

// renderTSTemplate sets up the TypeScript client template with custom template detection,
// shared by writeTSClient and renderTSClient to eliminate duplication.
func (g *Generator) renderTSTemplate(model GenerationModel) (*tt.Template, GenerationModel, error) {
	var tmplContent string
	if g.config.TSTemplatePath != "" {
		content, err := os.ReadFile(g.config.TSTemplatePath)
		if err != nil {
			return nil, model, fmt.Errorf("read custom TypeScript template: %w", err)
		}
		tmplContent = string(content)
	} else {
		tmplContent = tsClientTmpl
	}

	// text/template is fine for TS code generation (no user-controlled HTML)
	tmpl := tt.Must(tt.New("sdk.ts").Parse(tmplContent))
	return tmpl, model, nil
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

// renderDemoHTMLTemplate sets up the demo HTML template with HTML escaping funcMap,
// shared by writeDemoHTML and renderDemoHTML to eliminate duplication.
func (g *Generator) renderDemoHTMLTemplate(model GenerationModel) (*tt.Template, GenerationModel, error) {
	// Use text/template with manual HTML escaping for user-controlled fields.
	// html/template's JS escaping breaks generated JS code inside <script> blocks.
	tmpl := tt.Must(tt.New("index.html").Funcs(tt.FuncMap{
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
	return tmpl, model, nil
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
