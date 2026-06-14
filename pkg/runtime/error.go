package runtime

import (
	"errors"
	"fmt"
)

// WASMError represents an error that can be serialized to JavaScript
type WASMError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *WASMError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Error codes
const (
	ErrCodeInvalidConfig     = "INVALID_CONFIG"
	ErrCodeInitFailed        = "INIT_FAILED"
	ErrCodeRequestFailed     = "REQUEST_FAILED"
	ErrCodeSerializationFail = "SERIALIZATION_FAILED"
	ErrCodeDeserializationFail = "DESERIALIZATION_FAILED"
	ErrCodeNotInitialized    = "NOT_INITIALIZED"
	ErrCodeInvalidOperation  = "INVALID_OPERATION"
	ErrCodeNetworkError      = "NETWORK_ERROR"
	ErrCodeTimeout           = "TIMEOUT"
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