package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/fred29910/gowasm/pkg/generator"
	"github.com/fred29910/gowasm/pkg/runtime"
	"github.com/fred29910/gowasm/version"
	"github.com/urfave/cli/v2"
)

//go:embed oxlintrc.json
var defaultOxlintrc []byte

// JSONOutput represents the JSON output structure for CI/CD integration.
type JSONOutput struct {
	Success   bool       `json:"success"`
	Files     []FileInfo `json:"files,omitempty"`
	TotalSize int        `json:"totalSize,omitempty"`
	Duration  int64      `json:"duration"`
	Error     string     `json:"error,omitempty"`
}

// FileInfo represents a generated file.
type FileInfo struct {
	Path string `json:"path"`
	Size int    `json:"size"`
}

func main() {
	app := &cli.App{
		Name:    "gowasm-generator",
		Usage:   "Generate WASM HTTP SDK from OpenAPI specification",
		Version: version.Version,
		Commands: []*cli.Command{
			{
				Name:  "generate",
				Usage: "Generate Go and TypeScript SDK from an OpenAPI spec",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "spec",
						Aliases:  []string{"s"},
						Usage:    "Path to OpenAPI YAML specification",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "out",
						Aliases: []string{"o"},
						Usage:   "Output directory for generated code",
						Value:   "./generated",
					},
					&cli.StringFlag{
						Name:    "module",
						Aliases: []string{"m"},
						Usage:   "Go module name for generated code",
					},
					&cli.StringFlag{
						Name:    "package",
						Aliases: []string{"p"},
						Usage:   "Package name for generated code",
					},
					&cli.StringFlag{
						Name:  "oxlintrc",
						Usage: "Path to custom oxlintrc.json (default: use embedded config)",
					},
					&cli.BoolFlag{
						Name:  "oxlint-disable",
						Usage: "Disable oxlint after generation",
					},
				&cli.BoolFlag{
					Name:  "dry-run",
					Usage: "Show what would be generated without writing files",
				},
				&cli.BoolFlag{
					Name:  "wasm",
					Usage: "Build WASM after generation",
					Value: true,
				},
					&cli.StringFlag{
						Name:  "wasm-out",
						Usage: "Output path for WASM binary (default: <out>/main.wasm)",
					},
				&cli.StringFlag{
					Name:  "compiler",
					Usage: "WASM compiler to use: auto, tinygo, go",
					Value: "auto",
				},
				&cli.BoolFlag{
					Name:    "validation",
					Aliases: []string{"v"},
					Usage:   "Generate validation methods for request structs",
					Value:   true,
				},
				&cli.StringFlag{
					Name:  "go-template",
					Usage: "Path to custom Go template file (optional, uses embedded default if not set)",
				},
				&cli.StringFlag{
					Name:  "ts-template",
					Usage: "Path to custom TypeScript template file (optional, uses embedded default if not set)",
				},
				&cli.BoolFlag{
					Name:    "verbose",
					Aliases: []string{"V"},
					Usage:   "Show detailed progress during generation",
				},
				&cli.StringFlag{
					Name:  "output",
					Usage: "Output format: text (default) or json",
					Value: "text",
				},
			},
			Action: runGenerate,
			},
			{
				Name:   "init",
				Usage:  "Create a sample project structure",
				Action: runInit,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runGenerate(ctx *cli.Context) error {
	specPath := ctx.String("spec")
	outDir := ctx.String("out")
	moduleName := ctx.String("module")
	packageName := ctx.String("package")
	oxlintrcPath := ctx.String("oxlintrc")
	disableLint := ctx.Bool("oxlint-disable")
	buildWasm := ctx.Bool("wasm")
	wasmOut := ctx.String("wasm-out")
	compilerStr := ctx.String("compiler")
	validationEnabled := ctx.Bool("validation")
	goTemplatePath := ctx.String("go-template")
	tsTemplatePath := ctx.String("ts-template")
	dryRun := ctx.Bool("dry-run")
	verbose := ctx.Bool("verbose")
	outputFormat := ctx.String("output")
	if outputFormat != "text" && outputFormat != "json" {
		return fmt.Errorf("invalid output format: %s (must be text or json)", outputFormat)
	}

	start := time.Now()

	cfg := generator.NewConfig()
	if moduleName != "" {
		cfg.ModuleName = moduleName
	}
	if outDir != "" {
		cfg.OutputModule = outDir
	}
	if packageName != "" {
		cfg.Package = packageName
	}
	cfg.Validation = validationEnabled
	cfg.Verbose = verbose
	if goTemplatePath != "" {
		cfg.GoTemplatePath = goTemplatePath
	}
	if tsTemplatePath != "" {
		cfg.TSTemplatePath = tsTemplatePath
	}

	g := generator.NewGeneratorFromConfig(cfg)

	// Helper to output JSON
	outputJSON := func(out JSONOutput) {
		out.Duration = time.Since(start).Milliseconds()
		data, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(data))
	}

	if dryRun {
		result, err := g.GenerateDryRun(specPath, outDir)
		if err != nil {
			if outputFormat == "json" {
				outputJSON(JSONOutput{
					Success: false,
					Error:   err.Error(),
				})
				os.Exit(1)
			}
			return err
		}
		if outputFormat == "json" {
			files := make([]FileInfo, 0, len(result.Files))
			for _, f := range result.Files {
				files = append(files, FileInfo{Path: f.Path, Size: f.Size})
			}
			outputJSON(JSONOutput{
				Success:   true,
				Files:     files,
				TotalSize: result.TotalSize,
			})
			return nil
		}
		// Text output
		fmt.Println("Dry run: files that would be generated:")
		for _, file := range result.Files {
			fmt.Printf("  %s (%d bytes)\n", file.Path, file.Size)
		}
		fmt.Printf("\nTotal: %d file(s), %d bytes\n", len(result.Files), result.TotalSize)
		return nil
	}

	genResult, err := g.Generate(specPath, outDir)
	if err != nil {
		if outputFormat == "json" {
			outputJSON(JSONOutput{
				Success: false,
				Error:   err.Error(),
			})
			os.Exit(1)
		}
		return err
	}

	if !verbose && outputFormat != "json" {
		fmt.Printf("Generated SDK in %s\n", outDir)
	}

	if buildWasm {
		if verbose && outputFormat != "json" {
			fmt.Println("Building WASM...")
		}
		if err := runBuildWASM(outDir, wasmOut, compilerStr); err != nil {
			if outputFormat == "json" {
				outputJSON(JSONOutput{
					Success: false,
					Error:   err.Error(),
				})
				os.Exit(1)
			}
			return err
		}
	}

	if !disableLint {
		if verbose && outputFormat != "json" {
			fmt.Println("Running oxlint...")
		}
		if err := runOxlint(outDir, oxlintrcPath, outputFormat == "json"); err != nil {
			if outputFormat == "json" {
				outputJSON(JSONOutput{
					Success: false,
					Error:   fmt.Sprintf("oxlint failed: %v", err),
				})
				os.Exit(1)
			}
			return fmt.Errorf("oxlint failed: %w", err)
		}
	}

	if outputFormat == "json" {
		files := make([]FileInfo, 0, len(genResult.Files))
		for _, f := range genResult.Files {
			files = append(files, FileInfo{Path: f.Path, Size: f.Size})
		}
		outputJSON(JSONOutput{
			Success:   true,
			Files:     files,
			TotalSize: genResult.TotalSize,
		})
		return nil
	}

	if verbose && outputFormat != "json" {
		fmt.Println("All done!")
	}

	return nil
}

func runInit(ctx *cli.Context) error {
	fmt.Println("Creating sample project structure...")
	dirs := []string{"specs", "generated", "build"}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", d, err)
		}
	}
	fmt.Println("Created directories: specs/, generated/, build/")
	fmt.Println("Next: place your OpenAPI spec in specs/ and run:")
	fmt.Println("  gowasm-generator generate -s specs/openapi.yaml -o generated/")
	return nil
}

func runBuildWASM(outDir, wasmOut, compilerStr string) error {
	compiler := runtime.Compiler(compilerStr)

	if wasmOut == "" {
		wasmOut = filepath.Join(outDir, "main.wasm")
	}

	result, err := runtime.BuildWASM(wasmOut, outDir, compiler)
	if err != nil {
		return fmt.Errorf("wasm build failed: %w", err)
	}

	fmt.Printf("WASM built: %s (compiler: %s)\n", result.Output, result.Compiler)
	return nil
}

func runOxlint(targetDir, configPath string, quiet bool) error {
	var configFile string

	if configPath != "" {
		configFile = configPath
	} else {
		configFile = filepath.Join(targetDir, "oxlintrc.json")
		if err := os.WriteFile(configFile, defaultOxlintrc, 0644); err != nil {
			return fmt.Errorf("write embedded oxlintrc: %w", err)
		}
	}

	cmd := exec.Command("npx", "oxlint", "-c", configFile, "--no-ignore", targetDir)
	if quiet {
		cmd.Stdout = nil
		cmd.Stderr = nil
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("oxlint exit: %w", err)
	}
	return nil
}
