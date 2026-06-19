package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repository root (no go.mod found)")
		}
		dir = parent
	}
}

func TestIntegrationPetstoreUsable(t *testing.T) {
	root := findRepoRoot(t)
	outDir := t.TempDir()
	specPath := filepath.Join(root, "examples", "petstore", "openapi.yaml")

	cfg := &Config{
		ModuleName:    "test/petstore",
		OutputDir:     "petstore-generated",
		Package:       "generated",
		RuntimePath:   root,
		RuntimeImport: "github.com/fred29910/gowasm/pkg/runtime",
		Validation:    true,
	}
	g := NewGeneratorFromConfig(cfg)

	result, err := g.Generate(specPath, outDir)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if len(result.Files) != 5 {
		t.Errorf("expected 5 generated files, got %d", len(result.Files))
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = outDir
	if tidyOut, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed in generated module: %v\n%s", err, tidyOut)
	}

	buildCmd := exec.Command("go", "build", "./cmd/wasm")
	buildCmd.Dir = outDir
	buildCmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	if buildOut, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("generated module failed to build: %v\n%s", err, buildOut)
	}

	mainPath := filepath.Join(outDir, "cmd", "wasm", "main.go")
	info, err := os.Stat(mainPath)
	if err != nil {
		t.Fatalf("generated cmd/wasm/main.go is missing: %v", err)
	}
	if info.Size() == 0 {
		t.Error("generated cmd/wasm/main.go is empty")
	}
}
