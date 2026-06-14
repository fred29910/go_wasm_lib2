package generator

import (
	_ "embed"
	ht "html/template"
	"os"
	"path/filepath"
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

	// text/template is fine for TS code generation (no user-controlled HTML)
	tmpl := tt.Must(tt.New("sdk.ts").Parse(tsClientTmpl))

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
