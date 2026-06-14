package runtime

import (
	"fmt"
	"os"
	"os/exec"
)

// DetectTinyGo checks whether the tinygo compiler is available in PATH.
func DetectTinyGo() bool {
	_, err := exec.LookPath("tinygo")
	return err == nil
}

// Compiler represents the WASM compiler to use.
type Compiler string

const (
	// CompilerAuto automatically selects tinygo if available, otherwise falls back to go.
	CompilerAuto Compiler = "auto"
	// CompilerTinyGo forces use of tinygo.
	CompilerTinyGo Compiler = "tinygo"
	// CompilerGo forces use of standard go.
	CompilerGo Compiler = "go"
)

// BuildResult contains information about the WASM build.
type BuildResult struct {
	Compiler string
	Output   string
}

// BuildWASM compiles a Go package to WASM.
// It selects the compiler based on the compiler parameter:
//   - CompilerAuto: prefers tinygo, falls back to go
//   - CompilerTinyGo: requires tinygo (returns error if not found)
//   - CompilerGo: uses standard go compiler
func BuildWASM(outPath, packagePath string, compiler Compiler) (*BuildResult, error) {
	useTinyGo := false

	switch compiler {
	case CompilerTinyGo:
		if !DetectTinyGo() {
			return nil, fmt.Errorf("tinygo not found in PATH")
		}
		useTinyGo = true
	case CompilerAuto:
		useTinyGo = DetectTinyGo()
	case CompilerGo:
		useTinyGo = false
	default:
		return nil, fmt.Errorf("unknown compiler: %s", compiler)
	}

	if useTinyGo {
		return buildWithTinyGo(outPath, packagePath)
	}
	return buildWithGo(outPath, packagePath)
}

func buildWithTinyGo(outPath, packagePath string) (*BuildResult, error) {
	fmt.Printf("Building WASM with tinygo: %s -> %s\n", packagePath, outPath)
	cmd := exec.Command("tinygo", "build", "-o", outPath, "-target=wasm", packagePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tinygo build failed: %w", err)
	}
	return &BuildResult{Compiler: "tinygo", Output: outPath}, nil
}

func buildWithGo(outPath, packagePath string) (*BuildResult, error) {
	fmt.Printf("Building WASM with go: %s -> %s\n", packagePath, outPath)
	cmd := exec.Command("go", "build", "-trimpath", "-o", outPath, packagePath)
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go build failed: %w", err)
	}
	return &BuildResult{Compiler: "go", Output: outPath}, nil
}
