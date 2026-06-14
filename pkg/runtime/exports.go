//go:build js && wasm

package runtime

import (
	"context"
	"encoding/json"
	"strconv"
	"syscall/js"
)

// WASMExports holds the exported functions for JavaScript
type WASMExports struct {
	client   *HTTPClient
	promise  *PromiseHelper
	converter *Converter
	initialized bool
}

// NewWASMExports creates a new exports handler
func NewWASMExports() *WASMExports {
	return &WASMExports{
		promise:   NewPromiseHelper(),
		converter: NewConverter(),
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
			if len(args) < 1 || args[0].Type() == js.TypeUndefined || args[0].Type() == js.TypeNull {
				e.promise.RejectPromise(reject, NewError(ErrCodeInvalidConfig, "Expected config object", ""))
				return
			}

			configJS := args[0]
			
			// Parse config
			config := DefaultClientConfig()
			
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
			e.client = NewHTTPClient(config)
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
			if !e.initialized || e.client == nil {
				e.promise.RejectPromise(reject, ErrNotInitialized)
				return
			}

			if len(args) < 2 {
				e.promise.RejectPromise(reject, NewError(ErrCodeInvalidOperation, "Expected operationId and request object", ""))
				return
			}

			_ = args[0].String() // operationID for future use
			reqJS := args[1]

			// Parse request
			req := &Request{}
			
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
					// Convert map[string]interface{} to map[string]string
					req.Headers = make(map[string]string)
					for k, v := range headersMap {
						req.Headers[k] = toString(v)
					}
				}
			}

			if query := reqJS.Get("query"); query.Type() == js.TypeObject {
				queryMap, err := e.converter.JSValueToMap(query)
				if err == nil {
					// Convert to string map
					req.Query = make(map[string]string)
					for k, v := range queryMap {
						req.Query[k] = toString(v)
					}
				}
			}

			if body := reqJS.Get("body"); body.Type() != js.TypeUndefined && body.Type() != js.TypeNull {
				// Convert body to Go value
				bodyVal := e.converter.jsValueToInterface(body)
				req.Body = bodyVal
			}

			// Make the call
			ctx := context.Background()
			resp, err := e.client.Call(ctx, req)
			if err != nil {
				e.promise.RejectPromise(reject, err)
				return
			}

			// Convert response to JS
			respJS, err := e.converter.GoToJSValue(resp)
			if err != nil {
				e.promise.RejectPromise(reject, NewError(ErrCodeSerializationFail, "Failed to serialize response", err.Error()))
				return
			}

			e.promise.ResolvePromise(resolve, respJS)
		}()
	})
}

// setAuthToken sets the authorization token
func (e *WASMExports) setAuthToken(this js.Value, args []js.Value) interface{} {
	if !e.initialized || e.client == nil {
		return map[string]interface{}{
			"success": false,
			"error":   ErrNotInitialized.Error(),
		}
	}

	if len(args) < 1 {
		return map[string]interface{}{
			"success": false,
			"error":   "Expected token argument",
		}
	}

	token := args[0].String()
	scheme := "Bearer"
	if len(args) > 1 && args[1].Type() == js.TypeString {
		scheme = args[1].String()
	}

	e.client.SetAuthToken(token, scheme)

	return map[string]interface{}{
		"success": true,
	}
}

// clearAuthToken clears the authorization token
func (e *WASMExports) clearAuthToken(this js.Value, args []js.Value) interface{} {
	if !e.initialized || e.client == nil {
		return map[string]interface{}{
			"success": false,
			"error":   ErrNotInitialized.Error(),
		}
	}

	e.client.ClearAuthToken()

	return map[string]interface{}{
		"success": true,
	}
}

// getConfig returns the current client configuration
func (e *WASMExports) getConfig(this js.Value, args []js.Value) interface{} {
	if !e.initialized || e.client == nil {
		return map[string]interface{}{
			"success": false,
			"error":   ErrNotInitialized.Error(),
		}
	}

	config := e.client.GetConfig()
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
	case int:
		return strconv.Itoa(val)
	case int8:
		return strconv.Itoa(int(val))
	case int16:
		return strconv.Itoa(int(val))
	case int32:
		return strconv.Itoa(int(val))
	case int64:
		return strconv.FormatInt(val, 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case uint8:
		return strconv.FormatUint(uint64(val), 10)
	case uint16:
		return strconv.FormatUint(uint64(val), 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		// Try JSON marshal
		if bytes, err := json.Marshal(val); err == nil {
			return string(bytes)
		}
		return ""
	}
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