//go:build js && wasm

package runtime

import (
	"fmt"
	"reflect"
	"syscall/js"

	"github.com/norunners/vert"
)

// Converter handles conversion between JavaScript values and Go structs
type Converter struct{}

// NewConverter creates a new converter
func NewConverter() *Converter {
	return &Converter{}
}

// JSValueToGo converts a JavaScript value to a Go struct pointer
func (c *Converter) JSValueToGo(jsVal js.Value, target interface{}) error {
	v := vert.ValueOf(jsVal)
	return v.AssignTo(target)
}

// GoToJSValue converts a Go value to a JavaScript value
func (c *Converter) GoToJSValue(val interface{}) (js.Value, error) {
	v := vert.ValueOf(val)
	return v.JSValue(), nil
}

// dangerousJSKeys are JavaScript property names that could lead to prototype pollution.
var dangerousJSKeys = map[string]bool{
	"__proto__":            true,
	"constructor":          true,
	"prototype":            true,
	"__defineGetter__":     true,
	"__defineSetter__":     true,
	"__lookupGetter__":     true,
	"__lookupSetter__":     true,
	"hasOwnProperty":       true,
	"isPrototypeOf":        true,
	"propertyIsEnumerable": true,
	"toLocaleString":       true,
	"toString":             true,
	"valueOf":              true,
}

// JSValueToMap converts a JavaScript object to a Go map
func (c *Converter) JSValueToMap(jsVal js.Value) (map[string]interface{}, error) {
	if jsVal.Type() != js.TypeObject {
		return nil, NewContextError(
			ErrCodeDeserializationFail,
			"Expected JavaScript object",
			fmt.Sprintf("Got type %s instead", jsVal.Type().String()),
			"converter.go",
			51,
			"Pass a plain JavaScript object {}, not an array, function, or primitive value",
		)
	}

	result := make(map[string]interface{})
	keys := js.Global().Get("Object").Call("keys", jsVal)
	length := keys.Length()

	for i := 0; i < length; i++ {
		key := keys.Index(i).String()
		// Skip dangerous keys to prevent prototype pollution
		if dangerousJSKeys[key] {
			continue
		}
		val := jsVal.Get(key)
		result[key] = c.jsValueToInterface(val)
	}

	return result, nil
}

// MapToJSValue converts a Go map to a JavaScript object
func (c *Converter) MapToJSValue(m map[string]interface{}) js.Value {
	obj := js.Global().Get("Object").New()
	for k, v := range m {
		obj.Set(k, c.interfaceToJSValue(v))
	}
	return obj
}

// SliceToJSArray converts a Go slice to a JavaScript array
func (c *Converter) SliceToJSArray(slice interface{}) js.Value {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return js.Null()
	}

	length := rv.Len()
	arr := js.Global().Get("Array").New(length)
	for i := 0; i < length; i++ {
		arr.SetIndex(i, c.interfaceToJSValue(rv.Index(i).Interface()))
	}
	return arr
}

// JSArrayToSlice converts a JavaScript array to a Go slice
func (c *Converter) JSArrayToSlice(jsVal js.Value, elementType reflect.Type) (interface{}, error) {
	if jsVal.Type() != js.TypeObject || jsVal.Get("length").Type() == js.TypeUndefined {
		return nil, NewContextError(
			ErrCodeDeserializationFail,
			"Expected JavaScript array",
			fmt.Sprintf("Got type %s instead", jsVal.Type().String()),
			"converter.go",
			98,
			"Pass a JavaScript array [], not an object, function, or primitive value",
		)
	}

	length := jsVal.Get("length").Int()
	slice := reflect.MakeSlice(reflect.SliceOf(elementType), length, length)

	for i := 0; i < length; i++ {
		elem := jsVal.Index(i)
		goElem, err := c.jsValueToReflect(elem, elementType)
		if err != nil {
			return nil, err
		}
		slice.Index(i).Set(goElem)
	}

	return slice.Interface(), nil
}

// jsValueToInterface converts a JS value to a Go interface{}
func (c *Converter) jsValueToInterface(jsVal js.Value) interface{} {
	switch jsVal.Type() {
	case js.TypeNull, js.TypeUndefined:
		return nil
	case js.TypeBoolean:
		return jsVal.Bool()
	case js.TypeNumber:
		return jsVal.Float()
	case js.TypeString:
		return jsVal.String()
	case js.TypeObject:
		// Check if it's an array
		if jsVal.InstanceOf(js.Global().Get("Array")) {
			length := jsVal.Get("length").Int()
			arr := make([]interface{}, length)
			for i := 0; i < length; i++ {
				arr[i] = c.jsValueToInterface(jsVal.Index(i))
			}
			return arr
		}
		// Regular object
		obj := make(map[string]interface{})
		keys := js.Global().Get("Object").Call("keys", jsVal)
		length := keys.Length()
		for i := 0; i < length; i++ {
			key := keys.Index(i).String()
			obj[key] = c.jsValueToInterface(jsVal.Get(key))
		}
		return obj
	case js.TypeFunction:
		return jsVal
	default:
		return jsVal.String()
	}
}

// interfaceToJSValue converts a Go interface{} to a JS value
func (c *Converter) interfaceToJSValue(val interface{}) js.Value {
	if val == nil {
		return js.Null()
	}

	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Bool:
		return js.ValueOf(rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return js.ValueOf(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return js.ValueOf(rv.Uint())
	case reflect.Float32, reflect.Float64:
		return js.ValueOf(rv.Float())
	case reflect.String:
		return js.ValueOf(rv.String())
	case reflect.Slice, reflect.Array:
		return c.SliceToJSArray(val)
	case reflect.Map:
		if m, ok := val.(map[string]interface{}); ok {
			return c.MapToJSValue(m)
		}
		return js.Null()
	case reflect.Struct:
		return c.structToJSValue(rv)
	case reflect.Ptr:
		if rv.IsNil() {
			return js.Null()
		}
		return c.interfaceToJSValue(rv.Elem().Interface())
	default:
		return js.ValueOf(val)
	}
}

// structToJSValue converts a Go struct to a JavaScript object
func (c *Converter) structToJSValue(rv reflect.Value) js.Value {
	obj := js.Global().Get("Object").New()
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		// Check for json tag
		tag := fieldType.Tag.Get("json")
		if tag == "" || tag == "-" {
			tag = fieldType.Tag.Get("js")
		}
		if tag == "" || tag == "-" {
			tag = fieldType.Name
		}
		// Handle json tag options (omitempty, etc.)
		if idx := len(tag); idx > 0 {
			if commaIdx := -1; true {
				for j, c := range tag {
					if c == ',' {
						commaIdx = j
						break
					}
				}
				if commaIdx > 0 {
					tag = tag[:commaIdx]
				}
			}
		}

		jsVal := c.interfaceToJSValue(field.Interface())
		obj.Set(tag, jsVal)
	}

	return obj
}

// jsValueToReflect converts a JS value to a reflect.Value of the given type
func (c *Converter) jsValueToReflect(jsVal js.Value, targetType reflect.Type) (reflect.Value, error) {
	switch targetType.Kind() {
	case reflect.Bool:
		return reflect.ValueOf(jsVal.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(jsVal.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.ValueOf(uint64(jsVal.Int())), nil
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(jsVal.Float()), nil
	case reflect.String:
		return reflect.ValueOf(jsVal.String()), nil
	case reflect.Slice:
		slice, err := c.JSArrayToSlice(jsVal, targetType.Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(slice), nil
	case reflect.Map:
		m, err := c.JSValueToMap(jsVal)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(m), nil
	case reflect.Struct:
		ptr := reflect.New(targetType)
		err := c.JSValueToGo(jsVal, ptr.Interface())
		if err != nil {
			return reflect.Value{}, err
		}
		return ptr.Elem(), nil
	case reflect.Ptr:
		elem, err := c.jsValueToReflect(jsVal, targetType.Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		ptr := reflect.New(targetType.Elem())
		ptr.Elem().Set(elem)
		return ptr, nil
	default:
		return reflect.Value{}, NewContextError(
			ErrCodeDeserializationFail,
			"Unsupported type for conversion",
			fmt.Sprintf("Cannot convert JavaScript value to Go type %s", targetType.Kind().String()),
			"converter.go",
			271,
			"Supported types: bool, int*, uint*, float*, string, slice, map, struct, pointer. Use a supported Go type as the conversion target",
		)
	}
}
