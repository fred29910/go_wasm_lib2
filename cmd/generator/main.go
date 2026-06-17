package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

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
				Name:      "generate",
				Usage:     "Generate Go and TypeScript SDK from an OpenAPI spec",
				ArgsUsage: "[--flags]",
				Flags:     generateFlags(),
				Action:    runGenerate,
			},
			{
				Name:   "init",
				Usage:  "Create a sample project structure",
				Action: runInit,
			},
		},
	}

	// Configure exit handler for unified error output
	app.ExitErrHandler = func(ctx *cli.Context, err error) {
		if err == nil {
			return
		}
		if ctx != nil && ctx.String("output") == "json" {
			output := JSONOutput{
				Success: false,
				Error:   err.Error(),
			}
			data, _ := json.MarshalIndent(output, "", "  ")
			fmt.Fprintln(os.Stderr, string(data))
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}

	if err := app.Run(os.Args); err != nil {
		// ExitErrHandler already handled output; just exit with code 1
		os.Exit(1)
	}
}
