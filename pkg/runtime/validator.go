package runtime

import "fmt"

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
	dotIndex := -1
	for i, c := range domain {
		if c == '.' {
			if dotIndex != -1 {
				return false
			}
			dotIndex = i
		}
	}
	if dotIndex < 1 || dotIndex >= len(domain)-1 {
		return false
	}
	return true
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

// IsValidDateTime checks if a string is a valid ISO 8601 datetime (YYYY-MM-DDTHH:MM:SS±ZZ:ZZ).
func IsValidDateTime(dt string) bool {
	if len(dt) < 10 {
		return false
	}
	if dt[4] != '-' || dt[7] != '-' {
		return false
	}
	isDigit := func(c byte) bool {
		return c >= '0' && c <= '9'
	}
	for i := 0; i < 4; i++ {
		if !isDigit(dt[i]) {
			return false
		}
	}
	for i := 5; i < 7; i++ {
		if !isDigit(dt[i]) {
			return false
		}
	}
	for i := 8; i < 10; i++ {
		if !isDigit(dt[i]) {
			return false
		}
	}
	if len(dt) == 10 {
		return true
	}
	if dt[10] != 'T' {
		return false
	}
	if len(dt) < 19 {
		return false
	}
	if dt[13] != ':' || dt[16] != ':' {
		return false
	}
	for i := 11; i < 13; i++ {
		if !isDigit(dt[i]) {
			return false
		}
	}
	for i := 14; i < 16; i++ {
		if !isDigit(dt[i]) {
			return false
		}
	}
	for i := 17; i < 19; i++ {
		if !isDigit(dt[i]) {
			return false
		}
	}
	if len(dt) == 19 {
		return true
	}
	if dt[19] == 'Z' {
		return true
	}
	if (dt[19] == '+' || dt[19] == '-') && len(dt) == 25 && dt[22] == ':' {
		for i := 20; i < 22; i++ {
			if !isDigit(dt[i]) {
				return false
			}
		}
		for i := 23; i < 25; i++ {
			if !isDigit(dt[i]) {
				return false
			}
		}
		return true
	}
	return false
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
