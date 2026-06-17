package errors

import (
	"errors"
	"testing"
)

func TestWASMError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *WASMError
		expected string
	}{
		{
			"simple error",
			&WASMError{Code: "TEST", Message: "test message"},
			"[TEST] test message",
		},
		{
			"with details",
			&WASMError{Code: "TEST", Message: "msg", Details: "detail"},
			"[TEST] msg: detail",
		},
		{
			"with file path",
			&WASMError{Code: "TEST", Message: "msg", FilePath: "client.go"},
			"[TEST] msg (in client.go)",
		},
		{
			"with file path and line",
			&WASMError{Code: "TEST", Message: "msg", FilePath: "client.go", LineNumber: 42},
			"[TEST] msg (in client.go:42)",
		},
		{
			"with suggestion",
			&WASMError{Code: "TEST", Message: "msg", Suggestion: "try this"},
			"[TEST] msg\nSuggestion: try this",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("WASMError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestWASMError_Unwrap(t *testing.T) {
	inner := errors.New("inner error")
	wrapped := &WASMError{
		Code:    "TEST",
		Message: "outer",
		err:     inner,
	}

	if !errors.Is(wrapped, inner) {
		t.Error("WASMError should unwrap to inner error via errors.Is")
	}

	var target *WASMError
	if !errors.As(wrapped, &target) {
		t.Error("WASMError should match via errors.As")
	}
}

func TestNewError(t *testing.T) {
	err := NewError("CODE", "message", "details")
	if err.Code != "CODE" || err.Message != "message" || err.Details != "details" {
		t.Errorf("NewError returned unexpected: %+v", err)
	}
}

func TestNewContextError(t *testing.T) {
	err := NewContextError("CODE", "msg", "details", "file.go", 10, "suggestion")
	if err.FilePath != "file.go" || err.LineNumber != 10 || err.Suggestion != "suggestion" {
		t.Errorf("NewContextError returned unexpected: %+v", err)
	}
}

func TestWithContext(t *testing.T) {
	err := &WASMError{Code: "TEST", Message: "msg"}
	err = err.WithContext("file.go", 42, "fix it")
	if err.FilePath != "file.go" || err.LineNumber != 42 || err.Suggestion != "fix it" {
		t.Errorf("WithContext returned unexpected: %+v", err)
	}
}

func TestWrapWASMError(t *testing.T) {
	inner := errors.New("inner")
	wrapped := WrapWASMError("CODE", "outer", inner, "suggestion")

	if wrapped.Code != "CODE" || wrapped.Message != "outer" {
		t.Errorf("WrapWASMError returned unexpected code/message: %s/%s", wrapped.Code, wrapped.Message)
	}
	if wrapped.Details != "inner" {
		t.Errorf("WrapWASMError details = %q, want %q", wrapped.Details, "inner")
	}
	if wrapped.Suggestion != "suggestion" {
		t.Errorf("WrapWASMError suggestion = %q, want %q", wrapped.Suggestion, "suggestion")
	}
	if wrapped.FilePath == "" || wrapped.FilePath == "unknown" {
		t.Errorf("WrapWASMError FilePath should be auto-detected, got %q", wrapped.FilePath)
	}
	if wrapped.LineNumber == 0 {
		t.Error("WrapWASMError LineNumber should be auto-detected")
	}
	if wrapped.Unwrap() != inner {
		t.Error("WrapWASMError.Unwrap() should return inner error")
	}
}

func TestFromError(t *testing.T) {
	// nil input
	if FromError(nil) != nil {
		t.Error("FromError(nil) should return nil")
	}

	// already WASMError
	wasmErr := &WASMError{Code: "TEST", Message: "msg"}
	result := FromError(wasmErr)
	if result != wasmErr {
		t.Error("FromError should return same WASMError")
	}

	// standard error
	stdErr := errors.New("standard error")
	result = FromError(stdErr)
	if result.Code != ErrCodeRequestFailed || result.Message != "standard error" {
		t.Errorf("FromError(standard) returned unexpected: %+v", result)
	}
}

func TestErrorCodes(t *testing.T) {
	codes := map[string]string{
		"ErrCodeInvalidConfig":       ErrCodeInvalidConfig,
		"ErrCodeInitFailed":          ErrCodeInitFailed,
		"ErrCodeRequestFailed":       ErrCodeRequestFailed,
		"ErrCodeSerializationFail":   ErrCodeSerializationFail,
		"ErrCodeDeserializationFail": ErrCodeDeserializationFail,
		"ErrCodeNotInitialized":      ErrCodeNotInitialized,
		"ErrCodeInvalidOperation":    ErrCodeInvalidOperation,
		"ErrCodeNetworkError":        ErrCodeNetworkError,
		"ErrCodeTimeout":             ErrCodeTimeout,
	}
	for name, code := range codes {
		if code == "" {
			t.Errorf("error code %s is empty", name)
		}
	}
}

func TestPredefinedErrors(t *testing.T) {
	if ErrNotInitialized.Code != ErrCodeNotInitialized {
		t.Error("ErrNotInitialized has wrong code")
	}
	if ErrInvalidOperation.Code != ErrCodeInvalidOperation {
		t.Error("ErrInvalidOperation has wrong code")
	}
}
