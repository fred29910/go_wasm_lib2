package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ClientConfig holds configuration for the HTTP client
type ClientConfig struct {
	BaseURL     string            `json:"baseUrl"`
	Timeout     int               `json:"timeout,omitempty"` // seconds
	Headers     map[string]string `json:"headers,omitempty"`
	Credentials string            `json:"credentials,omitempty"` // "include", "omit", "same-origin"
}

// DefaultClientConfig returns a default configuration
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		BaseURL:     "",
		Timeout:     30,
		Headers:     make(map[string]string),
		Credentials: "same-origin",
	}
}

// OperationHandler handles a generated OpenAPI operation
type OperationHandler func(context.Context, Request) (*Response, error)

var (
	// DefaultClient is used by generated operation handlers
	DefaultClient = NewHTTPClient(nil)

	operationsMu sync.RWMutex
	operations   = make(map[string]OperationHandler)
)

// RegisterOperation registers a generated OpenAPI operation by operationId
func RegisterOperation(operationID string, handler OperationHandler) {
	if operationID == "" || handler == nil {
		return
	}
	operationsMu.Lock()
	defer operationsMu.Unlock()
	operations[operationID] = handler
}

// GetOperation returns a registered operation handler
func GetOperation(operationID string) (OperationHandler, bool) {
	operationsMu.RLock()
	defer operationsMu.RUnlock()
	handler, ok := operations[operationID]
	return handler, ok
}

// SetDefaultClient replaces the default HTTP client
func SetDefaultClient(config *ClientConfig) {
	DefaultClient = NewHTTPClient(config)
}

// HTTPClient is a generic HTTP client for WASM
type HTTPClient struct {
	config     *ClientConfig
	httpClient *http.Client
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(config *ClientConfig) *HTTPClient {
	if config == nil {
		config = DefaultClientConfig()
	}

	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &HTTPClient{
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Request represents an HTTP request
type Request struct {
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	PathParams  map[string]string `json:"pathParams,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Query       map[string]string `json:"query,omitempty"`
	Body        interface{}       `json:"body,omitempty"`
}

// Response represents an HTTP response
type Response struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
	Error   *WASMError        `json:"error,omitempty"`
}

// Call makes an HTTP request and returns a Response
func (c *HTTPClient) Call(ctx context.Context, req *Request) (*Response, error) {
	// Build URL
	fullURL := c.buildURL(req.Path, req.Query)

	// Serialize body
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, NewError(ErrCodeSerializationFail, "Failed to serialize request body", err.Error())
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, fullURL, bodyReader)
	if err != nil {
		return nil, NewError(ErrCodeRequestFailed, "Failed to create request", err.Error())
	}

	// Set headers
	for k, v := range c.config.Headers {
		httpReq.Header.Set(k, v)
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Set content type for JSON bodies
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		// Check if it's a timeout
		if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
			return nil, NewError(ErrCodeTimeout, "Request timeout", err.Error())
		}
		return nil, NewError(ErrCodeNetworkError, "Network error", err.Error())
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewError(ErrCodeRequestFailed, "Failed to read response", err.Error())
	}

	// Parse response
	var body interface{}
	if len(bodyBytes) > 0 {
		// Try to parse as JSON first
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			// If not JSON, return as string
			body = string(bodyBytes)
		}
	}

	// Build response headers
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return &Response{
		Status:  resp.StatusCode,
		Headers: headers,
		Body:    body,
	}, nil
}

// ResolvePath replaces path parameters and appends query parameters
func ResolvePath(path string, pathParams map[string]string, query map[string]string) string {
	if path == "" {
		path = "/"
	}

	for k, v := range pathParams {
		path = strings.ReplaceAll(path, "{"+k+"}", url.PathEscape(v))
	}

	if len(query) == 0 {
		return path
	}

	params := url.Values{}
	for k, v := range query {
		params.Add(k, v)
	}

	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + params.Encode()
}

// buildURL constructs the full URL with query parameters
func (c *HTTPClient) buildURL(path string, query map[string]string) string {
	base := strings.TrimRight(c.config.BaseURL, "/")
	fullURL := ResolvePath(path, nil, query)
	fullURL = strings.TrimLeft(fullURL, "/")

	if base == "" {
		return fullURL
	}

	return base + "/" + fullURL
}

// SetAuthToken sets an authorization header
func (c *HTTPClient) SetAuthToken(token string, scheme string) {
	if scheme == "" {
		scheme = "Bearer"
	}
	c.config.Headers["Authorization"] = scheme + " " + token
}

// ClearAuthToken removes the authorization header
func (c *HTTPClient) ClearAuthToken() {
	delete(c.config.Headers, "Authorization")
}

// GetConfig returns the current configuration
func (c *HTTPClient) GetConfig() *ClientConfig {
	return c.config
}