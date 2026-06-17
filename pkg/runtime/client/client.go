package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	runtimeerrors "github.com/fred29910/gowasm/pkg/runtime/errors"
)

// maxResponseBodySize limits response body reads to 10 MB (OOM protection).
const maxResponseBodySize = 10 << 20

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
	mu         sync.RWMutex
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(config *ClientConfig) *HTTPClient {
	if config == nil {
		config = DefaultClientConfig()
	}

	if config.Headers == nil {
		config.Headers = make(map[string]string)
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
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	PathParams map[string]string `json:"pathParams,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Query      url.Values        `json:"query,omitempty"`
	Body       interface{}       `json:"body,omitempty"`
}

// Response represents an HTTP response
type Response struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
	Error   *runtimeerrors.WASMError        `json:"error,omitempty"`
}

// Call makes an HTTP request and returns a Response
func (c *HTTPClient) Call(ctx context.Context, req *Request) (*Response, error) {
	// Build URL with path params
	fullURL, err := c.buildURL(req.Path, req.PathParams, req.Query)
	if err != nil {
		return nil, runtimeerrors.WrapWASMError(
			runtimeerrors.ErrCodeRequestFailed,
			"Failed to build URL",
			err,
			"Check that path parameters do not contain invalid values (e.g. '..', '//', or absolute paths)",
		)
	}

	// Serialize body
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, runtimeerrors.WrapWASMError(
				runtimeerrors.ErrCodeSerializationFail,
				"Failed to serialize request body",
				err,
				"Ensure the request body contains only JSON-serializable types (string, number, bool, slice, map, struct)",
			)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, fullURL, bodyReader)
	if err != nil {
		return nil, runtimeerrors.WrapWASMError(
			runtimeerrors.ErrCodeRequestFailed,
			"Failed to create HTTP request",
			err,
			"Check that the method is valid (GET, POST, PUT, DELETE, etc.) and the URL is well-formed",
		)
	}

	// Set headers (protected by read lock)
	c.mu.RLock()
	for k, v := range c.config.Headers {
		httpReq.Header.Set(k, v)
	}
	c.mu.RUnlock()
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
			return nil, runtimeerrors.WrapWASMError(
				runtimeerrors.ErrCodeTimeout,
				"Request timeout",
				err,
				fmt.Sprintf("Increase the timeout value in ClientConfig (current: %v seconds)", c.config.Timeout),
			)
		}
		return nil, runtimeerrors.WrapWASMError(
			runtimeerrors.ErrCodeNetworkError,
			"Network error",
			err,
			"Check network connectivity, verify the BaseURL is correct, and ensure the server is reachable",
		)
	}
	defer resp.Body.Close()

	// Read response body (capped at 10 MB for OOM protection)
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
	if err != nil {
		return nil, runtimeerrors.WrapWASMError(
			runtimeerrors.ErrCodeRequestFailed,
			"Failed to read response body",
			err,
			"The response body may have been corrupted, exceeded 10MB limit, or the connection closed prematurely",
		)
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

var pathParamRe = regexp.MustCompile(`\{([^}]+)\}`)

// ErrInvalidPathParam is returned when a path parameter contains
// potentially dangerous values (path traversal, absolute paths, etc.).
var ErrInvalidPathParam = errors.New("invalid path parameter")

// ResolvePath replaces path parameters and appends query parameters.
// It detects and rejects path traversal attempts (e.g. "..").
// Uses regex for deterministic, single-pass replacement (avoids map-order injection).
// Returns ErrInvalidPathParam when a path parameter contains dangerous values.
func ResolvePath(path string, pathParams map[string]string, query url.Values) (string, error) {
	if path == "" {
		path = "/"
	}

	var resolveErr error
	path = pathParamRe.ReplaceAllStringFunc(path, func(match string) string {
		key := match[1 : len(match)-1]
		if v, exists := pathParams[key]; exists {
			safe, err := safePathParam(v)
			if err != nil {
				resolveErr = fmt.Errorf("%w: %q", ErrInvalidPathParam, v)
				return match // keep original placeholder on error
			}
			return url.PathEscape(safe)
		}
		return match
	})

	if resolveErr != nil {
		return "", resolveErr
	}

	if query == nil || len(query) == 0 {
		return path, nil
	}

	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + query.Encode(), nil
}

// safePathParam validates and returns a safe path parameter value.
// Returns an error if the value contains path traversal attempts.
func safePathParam(v string) (string, error) {
	unescaped, err := url.PathUnescape(v)
	if err != nil {
		return "", fmt.Errorf("decode path parameter: %w", err)
	}
	if strings.Contains(unescaped, "..") || strings.Contains(unescaped, "//") || strings.HasPrefix(unescaped, "/") {
		return "", fmt.Errorf("%w: contains dangerous pattern", ErrInvalidPathParam)
	}
	return v, nil
}

// buildURL constructs the full URL with path and query parameters using url.JoinPath.
func (c *HTTPClient) buildURL(path string, pathParams map[string]string, query url.Values) (string, error) {
	resolvedPath, err := ResolvePath(path, pathParams, query)
	if err != nil {
		return "", err
	}

	if c.config.BaseURL == "" {
		return resolvedPath, nil
	}

	finalURL, err := url.JoinPath(c.config.BaseURL, resolvedPath)
	if err != nil {
		// Fallback: url.JoinPath should only fail on invalid base URLs
		return c.config.BaseURL + resolvedPath, nil
	}
	return finalURL, nil
}

// SetAuthToken sets an authorization header
func (c *HTTPClient) SetAuthToken(token string, scheme string) {
	if scheme == "" {
		scheme = "Bearer"
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.Headers["Authorization"] = scheme + " " + token
}

// ClearAuthToken removes the authorization header
func (c *HTTPClient) ClearAuthToken() {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.config.Headers, "Authorization")
}

// GetConfig returns a copy of the current configuration.
func (c *HTTPClient) GetConfig() *ClientConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	headers := make(map[string]string, len(c.config.Headers))
	for k, v := range c.config.Headers {
		headers[k] = v
	}

	return &ClientConfig{
		BaseURL:     c.config.BaseURL,
		Timeout:     c.config.Timeout,
		Headers:     headers,
		Credentials: c.config.Credentials,
	}
}
