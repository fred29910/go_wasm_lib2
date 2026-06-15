package runtime

import (
	"context"
	"testing"
)

func TestResolvePath(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		pathParams map[string]string
		query      map[string]string
		expected   string
	}{
		{
			name:     "empty path",
			path:     "",
			expected: "/",
		},
		{
			name:     "simple path",
			path:     "/pets",
			expected: "/pets",
		},
		{
			name:       "path with params",
			path:       "/pets/{petId}",
			pathParams: map[string]string{"petId": "123"},
			expected:   "/pets/123",
		},
		{
			name:     "path with query",
			path:     "/pets",
			query:    map[string]string{"status": "available"},
			expected: "/pets?status=available",
		},
		{
			name:       "path with params and query",
			path:       "/pets/{petId}",
			pathParams: map[string]string{"petId": "123"},
			query:      map[string]string{"status": "available"},
			expected:   "/pets/123?status=available",
		},
		{
			name:       "path traversal attempt with ..",
			path:       "/pets/{petId}",
			pathParams: map[string]string{"petId": "../etc/passwd"},
			expected:   "/pets/",
		},
		{
			name:       "path traversal attempt with //",
			path:       "/pets/{petId}",
			pathParams: map[string]string{"petId": "foo//bar"},
			expected:   "/pets/",
		},
		{
			name:       "url encoded path traversal attempt",
			path:       "/pets/{petId}",
			pathParams: map[string]string{"petId": "%2e%2e%2fetc%2fpasswd"},
			expected:   "/pets/",
		},
		{
			name:       "url encoded absolute path attempt",
			path:       "/pets/{petId}",
			pathParams: map[string]string{"petId": "%2Fetc%2Fpasswd"},
			expected:   "/pets/",
		},
		{
			name:     "path with existing query",
			path:     "/pets?limit=10",
			query:    map[string]string{"status": "available"},
			expected: "/pets?limit=10&status=available",
		},
		{
			name:       "special chars in path param",
			path:       "/pets/{petId}",
			pathParams: map[string]string{"petId": "hello world"},
			expected:   "/pets/hello%20world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolvePath(tt.path, tt.pathParams, tt.query)
			if got != tt.expected {
				t.Errorf("ResolvePath(%q, %v, %v) = %q, want %q",
					tt.path, tt.pathParams, tt.query, got, tt.expected)
			}
		})
	}
}

func TestDefaultClientConfig(t *testing.T) {
	cfg := DefaultClientConfig()
	if cfg.Timeout != 30 {
		t.Errorf("expected timeout 30, got %d", cfg.Timeout)
	}
	if cfg.Credentials != "same-origin" {
		t.Errorf("expected credentials same-origin, got %s", cfg.Credentials)
	}
	if cfg.Headers == nil {
		t.Error("expected Headers to be initialized")
	}
}

func TestHTTPClient(t *testing.T) {
	client := NewHTTPClient(nil)
	if client == nil {
		t.Fatal("NewHTTPClient returned nil")
	}
	if client.config == nil {
		t.Fatal("client.config is nil")
	}
	if client.httpClient == nil {
		t.Fatal("client.httpClient is nil")
	}
}

func TestHTTPClientCustomConfig(t *testing.T) {
	cfg := &ClientConfig{
		BaseURL:     "https://example.com",
		Timeout:     60,
		Headers:     map[string]string{"X-Custom": "value"},
		Credentials: "include",
	}
	client := NewHTTPClient(cfg)
	if client.config.BaseURL != "https://example.com" {
		t.Errorf("unexpected BaseURL: %s", client.config.BaseURL)
	}
	if client.config.Timeout != 60 {
		t.Errorf("unexpected Timeout: %d", client.config.Timeout)
	}
}

func TestSetAuthToken(t *testing.T) {
	client := NewHTTPClient(nil)
	client.SetAuthToken("my-token", "Bearer")
	if client.config.Headers["Authorization"] != "Bearer my-token" {
		t.Errorf("unexpected auth header: %s", client.config.Headers["Authorization"])
	}

	// Default scheme
	client.SetAuthToken("my-token", "")
	if client.config.Headers["Authorization"] != "Bearer my-token" {
		t.Errorf("unexpected auth header with empty scheme: %s", client.config.Headers["Authorization"])
	}
}

func TestClearAuthToken(t *testing.T) {
	client := NewHTTPClient(nil)
	client.SetAuthToken("my-token", "Bearer")
	client.ClearAuthToken()
	if _, ok := client.config.Headers["Authorization"]; ok {
		t.Error("Authorization header should be removed")
	}
}

func TestGetConfig(t *testing.T) {
	cfg := DefaultClientConfig()
	client := NewHTTPClient(cfg)
	got := client.GetConfig()
	if got.Timeout != cfg.Timeout {
		t.Errorf("GetConfig returned wrong config")
	}
}

func TestRegisterOperation(t *testing.T) {
	RegisterOperation("testOp", func(ctx context.Context, req Request) (*Response, error) {
		return &Response{Status: 200}, nil
	})
	handler, ok := GetOperation("testOp")
	if !ok {
		t.Error("operation not registered")
	}
	if handler == nil {
		t.Error("handler is nil")
	}

	// Empty ID should be no-op
	RegisterOperation("", nil)

	// Non-existent operation
	_, ok = GetOperation("nonexistent")
	if ok {
		t.Error("should not find non-existent operation")
	}
}

func TestWASMError(t *testing.T) {
	err := &WASMError{Code: "TEST", Message: "test message"}
	if err.Error() != "[TEST] test message" {
		t.Errorf("unexpected error string: %s", err.Error())
	}

	errWithDetails := &WASMError{Code: "TEST", Message: "test", Details: "details"}
	if errWithDetails.Error() != "[TEST] test: details" {
		t.Errorf("unexpected error string with details: %s", errWithDetails.Error())
	}
}

func TestNewError(t *testing.T) {
	err := NewError("CODE", "message", "details")
	if err.Code != "CODE" || err.Message != "message" || err.Details != "details" {
		t.Errorf("NewError returned unexpected values: %+v", err)
	}
}

func TestFromError(t *testing.T) {
	// nil error
	if FromError(nil) != nil {
		t.Error("FromError(nil) should return nil")
	}

	// WASMError
	wasmErr := &WASMError{Code: "TEST", Message: "test"}
	result := FromError(wasmErr)
	if result.Code != "TEST" {
		t.Errorf("expected TEST, got %s", result.Code)
	}

	// Standard error
	stdErr := context.DeadlineExceeded
	result = FromError(stdErr)
	if result.Code != ErrCodeRequestFailed {
		t.Errorf("expected %s, got %s", ErrCodeRequestFailed, result.Code)
	}
}
