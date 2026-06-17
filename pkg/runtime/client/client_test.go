package client_test

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/fred29910/gowasm/pkg/runtime/client"
)

func TestResolvePath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		pathParams  map[string]string
		query       url.Values
		expected    string
		expectError bool
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
			query:    url.Values{"status": []string{"available"}},
			expected: "/pets?status=available",
		},
		{
			name:       "path with params and query",
			path:       "/pets/{petId}",
			pathParams: map[string]string{"petId": "123"},
			query:      url.Values{"status": []string{"available"}},
			expected:   "/pets/123?status=available",
		},
		{
			name:        "path traversal attempt with ..",
			path:        "/pets/{petId}",
			pathParams:  map[string]string{"petId": "../etc/passwd"},
			expectError: true,
		},
		{
			name:        "path traversal attempt with //",
			path:        "/pets/{petId}",
			pathParams:  map[string]string{"petId": "foo//bar"},
			expectError: true,
		},
		{
			name:        "url encoded path traversal attempt",
			path:        "/pets/{petId}",
			pathParams:  map[string]string{"petId": "%2e%2e%2fetc%2fpasswd"},
			expectError: true,
		},
		{
			name:        "url encoded absolute path attempt",
			path:        "/pets/{petId}",
			pathParams:  map[string]string{"petId": "%2Fetc%2Fpasswd"},
			expectError: true,
		},
		{
			name:     "path with existing query",
			path:     "/pets?limit=10",
			query:    url.Values{"status": []string{"available"}},
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
			got, err := client.ResolvePath(tt.path, tt.pathParams, tt.query)
			if tt.expectError {
				if err == nil {
					t.Errorf("ResolvePath(%q, %v, %v) expected error, got %q",
						tt.path, tt.pathParams, tt.query, got)
				}
				if !errors.Is(err, client.ErrInvalidPathParam) {
					t.Errorf("ResolvePath(%q, %v, %v) error = %v, want ErrInvalidPathParam",
						tt.path, tt.pathParams, tt.query, err)
				}
				return
			}
			if err != nil {
				t.Errorf("ResolvePath(%q, %v, %v) unexpected error: %v",
					tt.path, tt.pathParams, tt.query, err)
				return
			}
			if got != tt.expected {
				t.Errorf("ResolvePath(%q, %v, %v) = %q, want %q",
					tt.path, tt.pathParams, tt.query, got, tt.expected)
			}
		})
	}
}

func TestDefaultClientConfig(t *testing.T) {
	cfg := client.DefaultClientConfig()
	if cfg.Timeout != 30 {
		t.Errorf("expected timeout 30, got %d", cfg.Timeout)
	}
	if cfg.Credentials != "same-origin" {
		t.Errorf("expected credentials same-origin, got %s", cfg.Credentials)
	}
	if cfg.Headers == nil {
		t.Error("expected Headers to be initialized, got nil")
	}
}

func TestNewHTTPClient(t *testing.T) {
	// Test with nil config
	c := client.NewHTTPClient(nil)
	if c == nil {
		t.Fatal("expected non-nil client")
	}

	// Test with custom config
	cfg := &client.ClientConfig{
		BaseURL: "https://example.com",
		Timeout: 60,
		Headers: map[string]string{"X-Custom": "value"},
	}
	c = client.NewHTTPClient(cfg)
	if c == nil {
		t.Fatal("expected non-nil client")
	}

	// Test with nil Headers — should be initialized automatically
	cfgNoHeaders := &client.ClientConfig{
		BaseURL: "https://example.com",
	}
	c = client.NewHTTPClient(cfgNoHeaders)
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	// Verify Headers was initialized (should not panic)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SetAuthToken panicked on nil Headers: %v", r)
		}
	}()
	c.SetAuthToken("test-token", "Bearer")
}

func TestSetAuthToken(t *testing.T) {
	cfg := client.DefaultClientConfig()
	c := client.NewHTTPClient(cfg)

	c.SetAuthToken("my-token", "Bearer")
	// Verify the token was set via GetConfig
	got := c.GetConfig()
	if got.Headers["Authorization"] != "Bearer my-token" {
		t.Errorf("expected Authorization header 'Bearer my-token', got %q", got.Headers["Authorization"])
	}

	c.ClearAuthToken()
	got = c.GetConfig()
	if _, exists := got.Headers["Authorization"]; exists {
		t.Error("expected Authorization header to be removed")
	}
}

func TestGetConfigReturnsCopy(t *testing.T) {
	cfg := &client.ClientConfig{
		BaseURL: "https://example.com",
		Timeout: 30,
		Headers: map[string]string{"X-Test": "original"},
	}
	c := client.NewHTTPClient(cfg)

	// Get config and modify it
	got1 := c.GetConfig()
	got1.Headers["X-Test"] = "modified"
	got1.BaseURL = "https://modified.com"

	// Get config again — should be unchanged
	got2 := c.GetConfig()
	if got2.Headers["X-Test"] != "original" {
		t.Errorf("GetConfig did not return a copy: Headers modified to %q", got2.Headers["X-Test"])
	}
	if got2.BaseURL != "https://example.com" {
		t.Errorf("GetConfig did not return a copy: BaseURL modified to %q", got2.BaseURL)
	}
}

func TestCall(t *testing.T) {
	// Test with a non-existent server — should return a network error
	cfg := &client.ClientConfig{
		BaseURL: "http://127.0.0.1:1", // port 1 should be unreachable
		Timeout: 1,
	}
	c := client.NewHTTPClient(cfg)

	ctx := context.Background()
	req := &client.Request{
		Method: "GET",
		Path:   "/test",
	}

	resp, err := c.Call(ctx, req)
	if err == nil {
		t.Error("expected error for unreachable server")
		t.Logf("response: %+v", resp)
	}
}
