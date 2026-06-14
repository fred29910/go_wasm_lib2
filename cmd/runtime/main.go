//go:build js && wasm

package main

import "github.com/fred29910/gowasm/pkg/runtime"

func main() {
	runtime.ExportMain()
}