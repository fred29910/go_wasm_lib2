package validate

import (
	"fmt"
	"reflect"
	"time"
)

// IsValidEmail performs a basic email format check.
func IsValidEmail(email string) bool {
	if len(email) < 3 {
		return false
	}
	atIndex := -1
	for i, c := range email {
		if c == '@' {
			if atIndex != -1 {
				return false
			}
			atIndex = i
		}
	}
	if atIndex < 1 || atIndex >= len(email)-1 {
		return false
	}
	domain := email[atIndex+1:]
	if len(domain) < 3 {
		return false
	}
	// Domain must contain at least one dot, not at start or end
	dotFound := false
	for i, c := range domain {
		if c == '.' {
			if i == 0 || i == len(domain)-1 {
				return false
			}
			dotFound = true
		}
	}
	return dotFound
}

// IsValidUUID checks if a string is a valid UUID (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx).
func IsValidUUID(uuid string) bool {
	if len(uuid) != 36 {
		return false
	}
	hexDigits := "0123456789abcdefABCDEF"
	isValidHex := func(s string) bool {
		if len(s) == 0 {
			return false
		}
		for _, c := range s {
			found := false
			for _, h := range hexDigits {
				if c == h {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}
	return isValidHex(uuid[0:8]) && uuid[8] == '-' &&
		isValidHex(uuid[9:13]) && uuid[13] == '-' &&
		isValidHex(uuid[14:18]) && uuid[18] == '-' &&
		isValidHex(uuid[19:23]) && uuid[23] == '-' &&
		isValidHex(uuid[24:36])
}

// IsValidDateTime checks if a string is a valid ISO 8601 / RFC 3339 datetime.
func IsValidDateTime(dt string) bool {
	_, err := time.Parse(time.RFC3339, dt)
	return err == nil
}

// IsValidEnum checks if a value is in the allowed list.
func IsValidEnum(value interface{}, allowed []interface{}) bool {
	valStr := fmt.Sprintf("%v", value)
	for _, v := range allowed {
		if valStr == fmt.Sprintf("%v", v) {
			return true
		}
	}
	return false
}

// IsValid checks if a value is non-nil and non-zero.
func IsValid(value interface{}) bool {
	if value == nil {
		return false
	}
	// Use reflect to detect nil slices, maps, channels, etc.
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Slice, reflect.Map, reflect.Chan, reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return false
		}
	}
	switch v := value.(type) {
	case string:
		return v != ""
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	case bool:
		return v
	default:
		return true
	}
}
