// Package runtime provides the core runtime for the WASM HTTP SDK generator.
//
// This package is a facade that re-exports symbols from its sub-packages
// to maintain backward compatibility with existing import paths.
//
// Sub-packages:
//   - client: HTTP client, request/response types, operation registry
//   - errors: WASMError type and error codes
//   - validate: Email, UUID, DateTime validation
//   - convert: JS/Go value conversion (js/wasm only)
//   - wasm: JS exports and Promise helper (js/wasm only)
//   - build: WASM build orchestration
package runtime

import (
	"github.com/fred29910/gowasm/pkg/runtime/build"
	"github.com/fred29910/gowasm/pkg/runtime/client"
	"github.com/fred29910/gowasm/pkg/runtime/errors"
)

// Re-export client types and functions.
type (
	HTTPClient       = client.HTTPClient
	ClientConfig     = client.ClientConfig
	Request          = client.Request
	Response         = client.Response
	OperationHandler = client.OperationHandler
)

var (
	DefaultClientConfig = client.DefaultClientConfig
	NewHTTPClient       = client.NewHTTPClient
	SetDefaultClient    = client.SetDefaultClient
	RegisterOperation   = client.RegisterOperation
	GetOperation        = client.GetOperation
)

// Re-export error types and functions.
type (
	WASMError = errors.WASMError
)

const (
	ErrCodeInvalidConfig       = errors.ErrCodeInvalidConfig
	ErrCodeInitFailed          = errors.ErrCodeInitFailed
	ErrCodeRequestFailed       = errors.ErrCodeRequestFailed
	ErrCodeSerializationFail   = errors.ErrCodeSerializationFail
	ErrCodeDeserializationFail = errors.ErrCodeDeserializationFail
	ErrCodeNotInitialized      = errors.ErrCodeNotInitialized
	ErrCodeInvalidOperation    = errors.ErrCodeInvalidOperation
	ErrCodeNetworkError        = errors.ErrCodeNetworkError
	ErrCodeTimeout             = errors.ErrCodeTimeout
)

var (
	ErrNotInitialized   = errors.ErrNotInitialized
	ErrInvalidOperation = errors.ErrInvalidOperation
	NewError            = errors.NewError
	NewContextError     = errors.NewContextError
	WrapWASMError       = errors.WrapWASMError
	FromError           = errors.FromError
)

// Re-export build types and functions.
type (
	BuildError  = build.BuildError
	Compiler    = build.Compiler
	BuildResult = build.BuildResult
)

const (
	CompilerAuto    = build.CompilerAuto
	CompilerTinyGo  = build.CompilerTinyGo
	CompilerGo      = build.CompilerGo
)

var (
	DetectTinyGo  = build.DetectTinyGo
	BuildWASM     = build.BuildWASM
	RunModTidy    = build.RunModTidy
)
