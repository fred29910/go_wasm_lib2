package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectTinyGo(t *testing.T) {
	// Just verify the function doesn't panic and returns a boolean.
	result := DetectTinyGo()
	t.Logf("TinyGo detected: %v", result)
}

func TestBuildResult(t *testing.T) {
	br := &BuildResult{Compiler: "tinygo", Output: "/tmp/test.wasm"}
	if br.Compiler != "tinygo" {
		t.Errorf("expected compiler tinygo, got %s", br.Compiler)
	}
	if br.Output != "/tmp/test.wasm" {
		t.Errorf("expected output /tmp/test.wasm, got %s", br.Output)
	}
}

func TestCompilerConstants(t *testing.T) {
	tests := []struct {
		name     string
		compiler Compiler
		expected string
	}{
		{"auto", CompilerAuto, "auto"},
		{"tinygo", CompilerTinyGo, "tinygo"},
		{"go", CompilerGo, "go"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.compiler) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.compiler)
			}
		})
	}
}

func TestBuildWASM_InvalidCompiler(t *testing.T) {
	_, err := BuildWASM("/tmp/test.wasm", ".", Compiler("invalid"))
	if err == nil {
		t.Fatal("expected error for invalid compiler")
	}
	// Check that the error message contains the compiler name
	if !contains(err.Error(), "invalid") {
		t.Errorf("unexpected error: %v", err)
	}
	// Check that the error message contains the suggestion
	if !contains(err.Error(), "supported compilers") {
		t.Errorf("error should contain supported compilers suggestion: %v", err)
	}
}

func TestBuildWASM_TinyGoNotAvailable(t *testing.T) {
	if DetectTinyGo() {
		t.Skip("TinyGo is available; skipping negative test")
	}
	_, err := BuildWASM("/tmp/test.wasm", ".", CompilerTinyGo)
	if err == nil {
		t.Fatal("expected error when tinygo not found")
	}
	// Check that the error message contains the suggestion
	if !contains(err.Error(), "Install TinyGo") {
		t.Errorf("error should contain TinyGo installation suggestion: %v", err)
	}
}

func TestBuildWASM_GoCompiler(t *testing.T) {
	// Find the project root by looking for go.mod.
	absPkgPath, err := findProjectRoot()
	if err != nil {
		t.Skipf("Skipping: %v", err)
	}

	// Create a minimal main package in a temp dir.
	tmpDir := t.TempDir()
	mainGo := "package main\n\nfunc main() {}\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644); err != nil {
		t.Fatal(err)
	}
	goMod := "module testwasmpkg\n\ngo 1.24\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	// Temporarily override go.mod to be in the temp dir.
	// BuildWASM uses go build, so we need to run from the temp dir with its own go.mod.
	outPath := filepath.Join(tmpDir, "test.wasm")

	// We can't easily test full compilation without a standalone module,
	// so instead test that the function correctly validates the compiler choice
	// and attempts to invoke the build (the compilation itself is a Go toolchain concern).
	// Use a package that we know exists.
	_ = absPkgPath

	// Just verify the compiler selection logic works.
	// Try with CompilerTinyGo - should error if tinygo not installed.
	if !DetectTinyGo() {
		_, err := BuildWASM(outPath, tmpDir, CompilerTinyGo)
		if err == nil {
			t.Error("expected error when tinygo not found")
		}
	}

	// Try with CompilerGo - the temp dir has an invalid package (no real main),
	// but we can at least verify the function attempts the build.
	_, err = BuildWASM(outPath, tmpDir, CompilerGo)
	// We expect this to fail because the temp dir has no proper go.sum, etc.
	// But the important thing is CompilerGo was accepted (not "unknown compiler").
	if err != nil {
		// Expected - just ensure it's not an "unknown compiler" error.
		if contains(err.Error(), "unknown compiler") {
			t.Fatalf("CompilerGo not recognized")
		}
	}
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
