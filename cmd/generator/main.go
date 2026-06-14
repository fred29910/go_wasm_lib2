package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fred29910/gowasm/pkg/generator"
)

func main() {
	specPath := flag.String("spec", "", "Path to OpenAPI YAML specification")
	outDir := flag.String("out", "./generated", "Output directory for generated code")
	moduleName := flag.String("module", "", "Go module name for generated code")
	packageName := flag.String("package", "", "Package name for generated code")
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
}