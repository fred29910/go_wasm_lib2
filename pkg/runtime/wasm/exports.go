//go:build js && wasm

package wasm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"syscall/js"

	runtimeclient "github.com/fred29910/gowasm/pkg/runtime/client"
	runtimeerrors "github.com/fred29910/gowasm/pkg/runtime/errors"
	runtimeconvert "github.com/fred29910/gowasm/pkg/runtime/convert"
)

// WASMExports holds the exported functions for JavaScript
type WASMExports struct {
	client      *runtimeclient.HTTPClient
	promise     *PromiseHelper
	converter   *runtimeconvert.Converter
	initialized bool
	mu          sync.RWMutex
}

// NewWASMExports creates a new exports handler
func NewWASMExports() *WASMExports {
	return &WASMExports{
		promise:   NewPromiseHelper(),
		converter: runtimeconvert.NewConverter(),
	}
}

// ExportFunctions registers all JavaScript-callable functions
func (e *WASMExports) ExportFunctions() {
	js.Global().Set("wasmInitClient", js.FuncOf(e.initClient))
	js.Global().Set("wasmCallAPI", js.FuncOf(e.callAPI))
	js.Global().Set("wasmSetAuthToken", js.FuncOf(e.setAuthToken))
	js.Global().Set("wasmClearAuthToken", js.FuncOf(e.clearAuthToken))
	js.Global().Set("wasmGetConfig", js.FuncOf(e.getConfig))
}

// initClient initializes the HTTP client with configuration
// Args: config object { baseUrl, timeout, headers, credentials }
// Returns: Promise that resolves when initialized
func (e *WASMExports) initClient(this js.Value, args []js.Value) interface{} {
	return e.promise.CreatePromise(func(resolve, reject js.Value) {
		go func() {
			e.mu.Lock()
			defer e.mu.Unlock()

			if len(args) < 1 || args[0].Type() == js.TypeUndefined || args[0].Type() == js.TypeNull {
				e.promise.RejectPromise(reject, runtimeerrors.NewContextError(
					runtimeerrors.ErrCodeInvalidConfig,
					"Expected config object",
					"Received null, undefined, or no arguments",
					"exports.go",
					48,
					"Pass a configuration object: { baseUrl: string, timeout?: number, headers?: object, credentials?: string }",
				))
				return
			}

			configJS := args[0]

			// Parse config
			config := runtimeclient.DefaultClientConfig()

			if baseURL := configJS.Get("baseUrl"); baseURL.Type() == js.TypeString {
				config.BaseURL = baseURL.String()
			}

			if timeout := configJS.Get("timeout"); timeout.Type() == js.TypeNumber {
				config.Timeout = timeout.Int()
			}

			if credentials := configJS.Get("credentials"); credentials.Type() == js.TypeString {
				config.Credentials = credentials.String()
			}

			if headers := configJS.Get("headers"); headers.Type() == js.TypeObject {
				headersMap, err := e.converter.JSValueToMap(headers)
				if err == nil {
					// Convert map[string]interface{} to map[string]string
					config.Headers = make(map[string]string)
					for k, v := range headersMap {
						config.Headers[k] = toString(v)
					}
				}
			}

			// Create client
			e.client = runtimeclient.NewHTTPClient(config)
			e.initialized = true

			// Return success
			result := map[string]interface{}{
				"success": true,
				"message": "Client initialized successfully",
			}
			e.promise.ResolvePromise(resolve, result)
		}()
	})
}

// callAPI makes an API call
// Args: operationId (string), request object { method, path, headers, query, body }
// Returns: Promise that resolves with response
func (e *WASMExports) callAPI(this js.Value, args []js.Value) interface{} {
	return e.promise.CreatePromise(func(resolve, reject js.Value) {
		go func() {
			e.mu.RLock()
			initialized := e.initialized
			client := e.client
			e.mu.RUnlock()

			if !initialized || client == nil {
				e.promise.RejectPromise(reject, runtimeerrors.ErrNotInitialized)
				return
			}

			if len(args) < 2 {
				e.promise.RejectPromise(reject, runtimeerrors.NewContextError(
					runtimeerrors.ErrCodeInvalidOperation,
					"Expected operationId and request object",
					fmt.Sprintf("Received %d arguments", len(args)),
					"exports.go",
					111,
					"Call with two arguments: wasmCallAPI(operationId, { method: 'GET', path: '/api/resource' })",
				))
				return
			}

			operationID := args[0].String()
			reqJS := args[1]

			// Parse request
			req := &runtimeclient.Request{}

			if method := reqJS.Get("method"); method.Type() == js.TypeString {
				req.Method = method.String()
			} else {
				req.Method = "GET"
			}

			if path := reqJS.Get("path"); path.Type() == js.TypeString {
				req.Path = path.String()
			}

			if headers := reqJS.Get("headers"); headers.Type() == js.TypeObject {
				headersMap, err := e.converter.JSValueToMap(headers)
				if err == nil {
					req.Headers = make(map[string]string)
					for k, v := range headersMap {
						req.Headers[k] = toString(v)
					}
				}
			}

			// Parse pathParams
			if pathParamsJS := reqJS.Get("pathParams"); pathParamsJS.Type() == js.TypeObject {
				pathParamsMap, err := e.converter.JSValueToMap(pathParamsJS)
				if err == nil {
					req.PathParams = make(map[string]string)
					for k, v := range pathParamsMap {
						req.PathParams[k] = toString(v)
					}
				}
			}

			if query := reqJS.Get("query"); query.Type() == js.TypeObject {
				queryMap, err := e.converter.JSValueToMap(query)
				if err == nil {
					req.Query = make(url.Values)
					for k, v := range queryMap {
						req.Query.Add(k, toString(v))
					}
				}
			}

			if body := reqJS.Get("body"); body.Type() != js.TypeUndefined && body.Type() != js.TypeNull {
				// Convert body to Go value
				bodyVal := e.converter.JSValueToInterface(body)
				req.Body = bodyVal
			}

			// Check for registered operation handler first
			ctx := context.Background()
			var resp *runtimeclient.Response
			var err error

			if handler, ok := runtimeclient.GetOperation(operationID); ok {
				resp, err = handler(ctx, *req)
			} else {
				resp, err = client.Call(ctx, req)
			}

			if err != nil {
				e.promise.RejectPromise(reject, err)
				return
			}

			// Convert response to JS
			respJS, err := e.converter.GoToJSValue(resp)
			if err != nil {
				e.promise.RejectPromise(reject, runtimeerrors.NewContextError(
					runtimeerrors.ErrCodeSerializationFail,
					"Failed to serialize response to JavaScript",
					err.Error(),
					"exports.go",
					169,
					"The response contains types that cannot be converted to JavaScript values",
				))
				return
			}

			e.promise.ResolvePromise(resolve, respJS)
		}()
	})
}

// setAuthToken sets the authorization token
func (e *WASMExports) setAuthToken(this js.Value, args []js.Value) interface{} {
	e.mu.RLock()
	initialized := e.initialized
	client := e.client
	e.mu.RUnlock()

	if !initialized || client == nil {
		return map[string]interface{}{
			"success": false,
			"error":   runtimeerrors.ErrNotInitialized.Error(),
		}
	}

	if len(args) < 1 {
		return map[string]interface{}{
			"success": false,
			"error": runtimeerrors.NewContextError(
				runtimeerrors.ErrCodeInvalidOperation,
				"Expected token argument",
				"No token provided",
				"exports.go",
				194,
				"Call with a token string: wasmSetAuthToken('your-token', 'Bearer')",
			).Error(),
		}
	}

	token := args[0].String()
	scheme := "Bearer"
	if len(args) > 1 && args[1].Type() == js.TypeString {
		scheme = args[1].String()
	}

	client.SetAuthToken(token, scheme)

	return map[string]interface{}{
		"success": true,
	}
}

// clearAuthToken clears the authorization token
func (e *WASMExports) clearAuthToken(this js.Value, args []js.Value) interface{} {
	e.mu.RLock()
	initialized := e.initialized
	client := e.client
	e.mu.RUnlock()

	if !initialized || client == nil {
		return map[string]interface{}{
			"success": false,
			"error":   runtimeerrors.ErrNotInitialized.Error(),
		}
	}

	client.ClearAuthToken()

	return map[string]interface{}{
		"success": true,
	}
}

// getConfig returns the current client configuration
func (e *WASMExports) getConfig(this js.Value, args []js.Value) interface{} {
	e.mu.RLock()
	initialized := e.initialized
	client := e.client
	e.mu.RUnlock()

	if !initialized || client == nil {
		return map[string]interface{}{
			"success": false,
			"error":   runtimeerrors.ErrNotInitialized.Error(),
		}
	}

	config := client.GetConfig()
	configJS, err := e.converter.GoToJSValue(config)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
		"config":  configJS,
	}
}

// toString converts an interface{} to string
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		if val {
			return "true"
		}
		return "false"
	}
	if bytes, err := json.Marshal(v); err == nil {
		return string(bytes)
	}
	return fmt.Sprintf("%v", v)
}

// ExportMain is the main entry point for the WASM module
func ExportMain() {
	exports := NewWASMExports()
	exports.ExportFunctions()

	// Signal that WASM is ready
	js.Global().Set("wasmReady", true)

	// Keep the program running
	select {}
}
