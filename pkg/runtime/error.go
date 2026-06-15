package runtime

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
)

// WASMError represents an error that can be serialized to JavaScript
type WASMError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	FilePath   string `json:"filePath,omitempty"`
	LineNumber int    `json:"lineNumber,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
	err       error  `json:"-"`
}

// Unwrap returns the wrapped error, enabling errors.Is/As support.
func (e *WASMError) Unwrap() error {
	return e.err
}

func (e *WASMError) Error() string {
	var msg string
	if e.Details != "" {
		msg = fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	} else {
		msg = fmt.Sprintf("[%s] %s", e.Code, e.Message)
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

// Error codes
const (
	ErrCodeInvalidConfig       = "INVALID_CONFIG"
	ErrCodeInitFailed          = "INIT_FAILED"
	ErrCodeRequestFailed       = "REQUEST_FAILED"
	ErrCodeSerializationFail   = "SERIALIZATION_FAILED"
	ErrCodeDeserializationFail = "DESERIALIZATION_FAILED"
	ErrCodeNotInitialized      = "NOT_INITIALIZED"
	ErrCodeInvalidOperation    = "INVALID_OPERATION"
	ErrCodeNetworkError        = "NETWORK_ERROR"
	ErrCodeTimeout             = "TIMEOUT"
)

// Predefined errors
var (
	ErrNotInitialized = &WASMError{
		Code:    ErrCodeNotInitialized,
		Message: "Client not initialized. Call initClient() first.",
	}
	ErrInvalidOperation = &WASMError{
		Code:    ErrCodeInvalidOperation,
		Message: "Invalid operation ID",
	}
)

// NewError creates a new WASMError
func NewError(code, message, details string) *WASMError {
	return &WASMError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewContextError creates a new WASMError with context information
func NewContextError(code, message, details, filePath string, lineNumber int, suggestion string) *WASMError {
	return &WASMError{
		Code:       code,
		Message:    message,
		Details:    details,
		FilePath:   filePath,
		LineNumber: lineNumber,
		Suggestion: suggestion,
	}
}

// WithContext adds context information to an existing WASMError
func (e *WASMError) WithContext(filePath string, lineNumber int, suggestion string) *WASMError {
	e.FilePath = filePath
	e.LineNumber = lineNumber
	e.Suggestion = suggestion
	return e
}

// WrapWASMError creates a WASMError from an existing error with automatic caller info.
// skip is the number of stack frames to skip (0 = caller of WrapWASMError).
func WrapWASMError(code, message string, err error, suggestion string) *WASMError {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
	}
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	return &WASMError{
		Code:       code,
		Message:    message,
		Details:    errMsg,
		FilePath:   filepath.Base(file),
		LineNumber: line,
		Suggestion: suggestion,
		err:        err,
	}
}

// FromError converts a standard error to WASMError
func FromError(err error) *WASMError {
	if err == nil {
		return nil
	}
	var wasmErr *WASMError
	if errors.As(err, &wasmErr) {
		return wasmErr
	}
	return &WASMError{
		Code:    ErrCodeRequestFailed,
		Message: err.Error(),
	}
}
