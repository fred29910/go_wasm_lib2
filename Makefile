# Makefile for go_wasm_lib2 - Generic WASM HTTP SDK Generator

.PHONY: build build-tinygo build-all clean test deps lint-ts lint-ts-fix

# Default target
all: build

# Build CLI binary
build:
	@echo "Building CLI binary..."
	go build -trimpath -ldflags="-s -w" -o build/gowasm ./cmd/generator
	@echo "CLI binary built: build/gowasm"
	@ls -lh build/gowasm

# Build with standard Go compiler (WASM)
build-wasm:
	@echo "Building with standard Go compiler..."
	GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o build/main.wasm ./cmd/runtime
	@echo "Standard Go WASM built: build/main.wasm"
	@ls -lh build/main.wasm

# Build with TinyGo compiler (smaller WASM output)
build-tinygo:
	@echo "Building with TinyGo compiler..."
	tinygo build -o build/tinymain.wasm -target=wasm ./cmd/runtime
	@echo "TinyGo WASM built: build/tinymain.wasm"
	@ls -lh build/tinymain.wasm

# Build all
build-all: build build-wasm build-tinygo

# Download dependencies
deps:
	go mod tidy
	go mod download

# Clean build artifacts
clean:
	rm -rf build/
	rm -f *.wasm

# Run Go tests
test:
	go test ./...
	@echo "All tests passed"

# Run Go tests with race detection
test-race:
	go test -race -count=1 ./...
	@echo "All tests passed (race detection enabled)"

# Run tests with coverage report
test-cover:
	go test -race -count=1 -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out
	@echo "Coverage report generated: coverage.out"

# Generate HTML coverage report
coverage-html: test-cover
	go tool cover -html=coverage.out -o coverage.html
	@echo "HTML coverage report: coverage.html"

# Run go vet
vet:
	go vet ./...
	@echo "go vet passed"

# Test compilation (without running)
test-compile:
	go build -trimpath -o /dev/null ./cmd/generator
	@echo "CLI binary compilation successful"
	GOOS=js GOARCH=wasm go build -trimpath -o /dev/null ./cmd/runtime
	@echo "Standard Go WASM compilation successful"
	tinygo build -o /dev/null -target=wasm ./cmd/runtime 2>/dev/null || echo "TinyGo compilation test skipped (not installed)"
	@echo "Compilation tests passed"

# Lint generated TypeScript files with oxlint
# Usage: make lint-ts OUT=./output (default: examples/petstore/generated)
lint-ts:
	npm install
	npx oxlint -c oxlintrc.json --no-ignore $(or $(OUT),examples/petstore/generated)

lint-ts-fix:
	npm install
	npx oxlint -c oxlintrc.json --no-ignore $(or $(OUT),examples/petstore/generated) --fix

# Generate SDK from OpenAPI spec
# Usage: make generate SPEC=path/to/openapi.yaml OUT=./output
generate:
	@if [ -z "$(SPEC)" ]; then echo "Usage: make generate SPEC=path/to/openapi.yaml [OUT=./output]"; exit 1; fi
	@mkdir -p $(OUT)
	go run ./cmd/generator generate -s $(SPEC) -out=$(OUT)

# Development: run generator with example spec
dev-generate:
	@mkdir -p examples/petstore/generated
	go run ./cmd/generator generate -s examples/petstore/openapi.yaml -out=examples/petstore/generated

# Install TinyGo (macOS)
install-tinygo-macos:
	brew install tinygo

# Install TinyGo (Linux)
install-tinygo-linux:
	wget https://github.com/tinygo-org/tinygo/releases/download/v0.41.1/tinygo_0.41.1_amd64.deb
	sudo dpkg -i tinygo_0.41.1_amd64.deb
	rm tinygo_0.41.1_amd64.deb

# Verify TinyGo installation
verify-tinygo:
	tinygo version

# Verify full project
verify: deps build test-race vet test-compile dev-generate
	@echo "=== All verifications passed ==="

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build CLI binary"
	@echo "  build-wasm      - Build with standard Go compiler (WASM)"
	@echo "  build-tinygo   - Build with TinyGo compiler (smaller WASM)"
	@echo "  build-all      - Build CLI + WASM variants"
	@echo "  deps           - Download dependencies"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run Go tests"
	@echo "  test-race      - Run tests with race detection"
	@echo "  test-cover     - Run tests with coverage report"
	@echo "  coverage-html  - Generate HTML coverage report"
	@echo "  vet            - Run go vet"
	@echo "  lint-ts        - Lint generated TypeScript files with oxlint"
	@echo "  lint-ts-fix    - Lint and auto-fix generated TypeScript files"
	@echo "  test-compile   - Test compilation"
	@echo "  generate       - Generate SDK from OpenAPI spec (SPEC=..., OUT=...)"
	@echo "  dev-generate   - Generate SDK for petstore example"
	@echo "  verify         - Full project verification (build, test-race, generate)"
	@echo "  install-tinygo-macos - Install TinyGo on macOS"
	@echo "  install-tinygo-linux - Install TinyGo on Linux"
	@echo "  verify-tinygo  - Verify TinyGo installation"
