package runtime

import (
	"fmt"
	"os"
	"os/exec"
)

// BuildError represents an error that occurred during the build process
type BuildError struct {
	Message    string
	FilePath   string
	LineNumber int
	Suggestion string
	Inner      error
}

func (e *BuildError) Error() string {
	msg := e.Message
	if e.Inner != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Inner)
	}
	if e.FilePath != "" {
		if e.LineNumber > 0 {
			msg = fmt.Sprintf("%s (in %s:%d)", msg, e.FilePath, e.LineNumber)
		} else {
			msg = fmt.Sprintf("%s (in %s)", msg, e.FilePath)
		}
	}
	if e.Suggestion != "" {
		msg = fmt.Sprintf("%s\nSuggestion: %s", msg, e.Suggestion)
	}
	return msg
}

func (e *BuildError) Unwrap() error {
	return e.Inner
}

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
			return nil, &BuildError{
				Message:    "TinyGo compiler not found",
				FilePath:   "build.go",
				LineNumber: 44,
				Suggestion: "Install TinyGo: https://tinygo.org/getting-started/install/ or use CompilerAuto to fallback to standard Go",
			}
		}
		useTinyGo = true
	case CompilerAuto:
		useTinyGo = DetectTinyGo()
	case CompilerGo:
		useTinyGo = false
	default:
		return nil, &BuildError{
			Message:    fmt.Sprintf("Unknown compiler: %s", compiler),
			FilePath:   "build.go",
			LineNumber: 52,
			Suggestion: fmt.Sprintf("Use one of the supported compilers: %s, %s, or %s", CompilerAuto, CompilerTinyGo, CompilerGo),
		}
	}

	if useTinyGo {
		return buildWithTinyGo(outPath, packagePath)
	}
	return buildWithGo(outPath, packagePath)
}

func buildWithTinyGo(outPath, packagePath string) (*BuildResult, error) {
	fmt.Printf("Building WASM with tinygo: %s -> %s\n", packagePath, outPath)
	cmd := exec.Command("tinygo", "build", "-o", outPath, "-target=wasm", packagePath)
	if err := runBuildCommand(cmd); err != nil {
		return nil, &BuildError{
			Message:    "TinyGo build failed",
			FilePath:   "build.go",
			LineNumber: 102,
			Suggestion: "Check that the package path is correct and the code compiles with TinyGo. Run 'tinygo build -target=wasm' manually to see detailed errors",
			Inner:      err,
		}
	}
	return &BuildResult{Compiler: "tinygo", Output: outPath}, nil
}

func buildWithGo(outPath, packagePath string) (*BuildResult, error) {
	fmt.Printf("Building WASM with go: %s -> %s\n", packagePath, outPath)
	cmd := exec.Command("go", "build", "-trimpath", "-o", outPath, packagePath)
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	if err := runBuildCommand(cmd); err != nil {
		return nil, &BuildError{
			Message:    "Go build failed",
			FilePath:   "build.go",
			LineNumber: 123,
			Suggestion: "Ensure GOOS=js and GOARCH=wasm are set, and the package compiles for WASM target. Run 'GOOS=js GOARCH=wasm go build' manually to see detailed errors",
			Inner:      err,
		}
	}
	return &BuildResult{Compiler: "go", Output: outPath}, nil
}

func runBuildCommand(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
