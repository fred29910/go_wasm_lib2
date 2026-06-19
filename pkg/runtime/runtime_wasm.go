//go:build js && wasm

package runtime

import (
	"github.com/fred29910/gowasm/pkg/runtime/build"
	"github.com/fred29910/gowasm/pkg/runtime/client"
	"github.com/fred29910/gowasm/pkg/runtime/errors"
	"github.com/fred29910/gowasm/pkg/runtime/validate"
	"github.com/fred29910/gowasm/pkg/runtime/wasm"
)

// ExportMain is the main entry point for the WASM module.
// It is only available when building with GOOS=js GOARCH=wasm.
func ExportMain() {
	wasm.ExportMain()
}

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
	CompilerAuto   = build.CompilerAuto
	CompilerTinyGo = build.CompilerTinyGo
	CompilerGo     = build.CompilerGo
)

var (
	DetectTinyGo = build.DetectTinyGo
	BuildWASM    = build.BuildWASM
	RunModTidy   = build.RunModTidy
)

// Re-export validate functions.
var (
	IsValidEmail    = validate.IsValidEmail
	IsValidUUID     = validate.IsValidUUID
	IsValidDateTime = validate.IsValidDateTime
	IsValidEnum     = validate.IsValidEnum
	IsValid         = validate.IsValid
)
