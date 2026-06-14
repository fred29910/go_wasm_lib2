package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fred29910/gowasm/pkg/generator"
	"github.com/fred29910/gowasm/pkg/runtime"
	"github.com/fred29910/gowasm/version"
	"github.com/urfave/cli/v2"
)

//go:embed oxlintrc.json
var defaultOxlintrc []byte

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

	g := generator.NewGenerator(moduleName, outDir, packageName, outDir, "")

	if err := g.Generate(specPath, outDir); err != nil {
		return err
	}

	fmt.Printf("Generated SDK in %s\n", outDir)

	if buildWasm {
		if err := runBuildWASM(outDir, wasmOut, compilerStr); err != nil {
			return err
		}
	}

	if !disableLint {
		if err := runOxlint(outDir, oxlintrcPath); err != nil {
			return fmt.Errorf("oxlint failed: %w", err)
		}
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

func runOxlint(targetDir, configPath string) error {
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("oxlint exit: %w", err)
	}
	return nil
}
