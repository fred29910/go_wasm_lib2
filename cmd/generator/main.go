package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fred29910/gowasm/pkg/generator"
)

//go:embed oxlintrc.json
var defaultOxlintrc []byte

func main() {
	specPath := flag.String("spec", "", "Path to OpenAPI YAML specification")
	outDir := flag.String("out", "./generated", "Output directory for generated code")
	moduleName := flag.String("module", "", "Go module name for generated code")
	packageName := flag.String("package", "", "Package name for generated code")
	oxlintrc := flag.String("oxlintrc", "", "Path to custom oxlintrc.json (default: use embedded config)")
	disableLint := flag.Bool("oxlint-disable", false, "Disable oxlint after generation")
	flag.Parse()

	if *specPath == "" {
		fmt.Fprintln(os.Stderr, "Error: -spec flag is required")
		flag.Usage()
		os.Exit(1)
	}

	g := generator.NewGenerator(*moduleName, *outDir, *packageName, *outDir, "")

	if err := g.Generate(*specPath, *outDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated SDK in %s\n", *outDir)

	if !*disableLint {
		lintDir := *outDir
		if err := runOxlint(lintDir, *oxlintrc); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: oxlint failed: %v\n", err)
		}
	}
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