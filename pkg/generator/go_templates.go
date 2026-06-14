package generator

import (
	_ "embed"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/sdk.go.tmpl
var goClientTmpl string

//go:embed templates/go.mod.tmpl
var goModTmpl string

type goTemplateData struct {
	GenerationModel
}

func (g *Generator) writeGoClient(outDir string, model GenerationModel) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	data := goTemplateData{GenerationModel: model}
	tmpl := template.Must(template.New("generated.go").Parse(goClientTmpl))
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
