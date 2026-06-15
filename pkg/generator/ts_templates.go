package generator

import (
	_ "embed"
	"fmt"
	ht "html/template"
	"os"
	"path/filepath"
	"strings"
	tt "text/template"
)

//go:embed templates/sdk.ts.tmpl
var tsClientTmpl string

//go:embed templates/index.html.tmpl
var demoHTMLTmpl string

func (g *Generator) writeTSClient(outDir string, model GenerationModel) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	// Use custom template if provided, otherwise use embedded default
	var tmplContent string
	if g.config.TSTemplatePath != "" {
		content, err := os.ReadFile(g.config.TSTemplatePath)
		if err != nil {
			return fmt.Errorf("read custom TypeScript template: %w", err)
		}
		tmplContent = string(content)
	} else {
		tmplContent = tsClientTmpl
	}

	// text/template is fine for TS code generation (no user-controlled HTML)
	tmpl := tt.Must(tt.New("sdk.ts").Parse(tmplContent))

	file, err := os.Create(filepath.Join(outDir, "sdk.ts"))
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, model)
}

func (g *Generator) writeDemoHTML(outDir string, model GenerationModel) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	// Use text/template with manual HTML escaping for user-controlled fields.
	// html/template's JS escaping breaks generated JS code inside <script> blocks.
	tmpl := tt.Must(tt.New("index.html").Funcs(tt.FuncMap{
		"htmlEscape": func(s string) string { return ht.HTMLEscapeString(s) },
	}).Parse(demoHTMLTmpl))

	file, err := os.Create(filepath.Join(outDir, "index.html"))
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, model)
}

// renderTSClient renders the TypeScript client template to a string without writing to disk.
func (g *Generator) renderTSClient(model GenerationModel) (string, error) {
	// Use custom template if provided, otherwise use embedded default
	var tmplContent string
	if g.config.TSTemplatePath != "" {
		content, err := os.ReadFile(g.config.TSTemplatePath)
		if err != nil {
			return "", fmt.Errorf("read custom TypeScript template: %w", err)
		}
		tmplContent = string(content)
	} else {
		tmplContent = tsClientTmpl
	}

	// text/template is fine for TS code generation (no user-controlled HTML)
	tmpl := tt.Must(tt.New("sdk.ts").Parse(tmplContent))

	var buf strings.Builder
	if err := tmpl.Execute(&buf, model); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// renderDemoHTML renders the demo HTML template to a string without writing to disk.
func (g *Generator) renderDemoHTML(model GenerationModel) (string, error) {
	// Use text/template with manual HTML escaping for user-controlled fields.
	// html/template's JS escaping breaks generated JS code inside <script> blocks.
	tmpl := tt.Must(tt.New("index.html").Funcs(tt.FuncMap{
		"htmlEscape": func(s string) string { return ht.HTMLEscapeString(s) },
	}).Parse(demoHTMLTmpl))

	var buf strings.Builder
	if err := tmpl.Execute(&buf, model); err != nil {
		return "", err
	}
	return buf.String(), nil
}
