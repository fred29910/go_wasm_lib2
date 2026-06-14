//go:build js && wasm

package runtime

import (
	"errors"
	"syscall/js"
)

// PromiseHelper provides utilities for creating and managing JavaScript Promises
type PromiseHelper struct{}

// NewPromiseHelper creates a new promise helper
func NewPromiseHelper() *PromiseHelper {
	return &PromiseHelper{}
}

// CreatePromise creates a new JavaScript Promise
// The executor function receives resolve and reject functions
func (p *PromiseHelper) CreatePromise(executor func(resolve, reject js.Value)) js.Value {
	promiseConstructor := js.Global().Get("Promise")
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 2 {
			return nil
		}
		resolve := args[0]
		reject := args[1]
		executor(resolve, reject)
		return nil
	})
	return promiseConstructor.New(handler)
}

// ResolvePromise resolves a promise with a value
func (p *PromiseHelper) ResolvePromise(resolve js.Value, value interface{}) {
	resolve.Invoke(value)
}

// RejectPromise rejects a promise with an error
func (p *PromiseHelper) RejectPromise(reject js.Value, err error) {
	if err == nil {
		reject.Invoke(js.Global().Get("Error").New("Unknown error"))
		return
	}
	
	var wasmErr *WASMError
	if AsWASMError(err, &wasmErr) {
		// Create a JS Error with our custom properties
		jsErr := js.Global().Get("Error").New(wasmErr.Message)
		jsErr.Set("code", wasmErr.Code)
		if wasmErr.Details != "" {
			jsErr.Set("details", wasmErr.Details)
		}
		reject.Invoke(jsErr)
	} else {
		reject.Invoke(js.Global().Get("Error").New(err.Error()))
	}
}

// AsWASMError attempts to convert an error to WASMError
func AsWASMError(err error, target **WASMError) bool {
	if err == nil {
		return false
	}
	var wasmErr *WASMError
	if errors.As(err, &wasmErr) {
		*target = wasmErr
		return true
	}
	return false
}

// CreateResolvedPromise creates a promise that resolves immediately
func (p *PromiseHelper) CreateResolvedPromise(value interface{}) js.Value {
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.Call("resolve", value)
}

// CreateRejectedPromise creates a promise that rejects immediately
func (p *PromiseHelper) CreateRejectedPromise(err error) js.Value {
	promiseConstructor := js.Global().Get("Promise")
	jsErr := js.Global().Get("Error").New(err.Error())
	return promiseConstructor.Call("reject", jsErr)
}

// PromiseAll waits for all promises to resolve
func (p *PromiseHelper) PromiseAll(promises []js.Value) js.Value {
	promiseConstructor := js.Global().Get("Promise")
	jsPromises := js.Global().Get("Array").New(len(promises))
	for i, prom := range promises {
		jsPromises.SetIndex(i, prom)
	}
	return promiseConstructor.Call("all", jsPromises)
}