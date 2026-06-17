package validate

import (
	"testing"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"valid simple", "user@example.com", true},
		{"valid with subdomain", "user@mail.example.com", true},
		{"valid with plus", "user+tag@example.com", true},
		{"valid with dots", "first.last@example.com", true},
		{"too short", "ab", false},
		{"no at sign", "userexample.com", false},
		{"multiple at signs", "user@@example.com", false},
		{"at at start", "@example.com", false},
		{"at at end", "user@", false},
		{"no domain dot", "user@examplecom", false},
		{"dot at domain start", "user@.example.com", false},
		{"dot at domain end", "user@example.", false},
		{"empty string", "", false},
		{"single char", "a", false},
		{"two chars", "ab", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidEmail(tt.email)
			if got != tt.expected {
				t.Errorf("IsValidEmail(%q) = %v, want %v", tt.email, got, tt.expected)
			}
		})
	}
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name     string
		uuid     string
		expected bool
	}{
		{"valid lowercase", "550e8400-e29b-41d4-a716-446655440000", true},
		{"valid uppercase", "550E8400-E29B-41D4-A716-446655440000", true},
		{"valid mixed case", "550e8400-E29B-41d4-A716-446655440000", true},
		{"all zeros", "00000000-0000-0000-0000-000000000000", true},
		{"all f's", "ffffffff-ffff-ffff-ffff-ffffffffffff", true},
		{"wrong length short", "550e8400-e29b-41d4-a716-44665544000", false},
		{"wrong length long", "550e8400-e29b-41d4-a716-4466554400000", false},
		{"missing dashes", "550e8400e29b41d4a716446655440000", false},
		{"wrong dash position", "550e8400e29b-41d4-a716-446655440000", false},
		{"invalid character g", "550e8400-e29b-41d4-a716-44665544000g", false},
		{"empty string", "", false},
		{"spaces", "550e8400 e29b 41d4 a716 446655440000", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidUUID(tt.uuid)
			if got != tt.expected {
				t.Errorf("IsValidUUID(%q) = %v, want %v", tt.uuid, got, tt.expected)
			}
		})
	}
}

func TestIsValidDateTime(t *testing.T) {
	tests := []struct {
		name     string
		dt       string
		expected bool
	}{
		{"valid RFC339", "2026-06-17T16:09:00Z", true},
		{"valid with offset", "2026-06-17T16:09:00+08:00", true},
		{"valid with negative offset", "2026-06-17T16:09:00-05:00", true},
		{"valid with milliseconds", "2026-06-17T16:09:00.123Z", true},
		{"valid date only not RFC3339", "2026-06-17", false},
		{"empty string", "", false},
		{"garbage", "not-a-date", false},
		{"wrong format", "06/17/2026", false},
		{"missing T", "2026-06-17 16:09:00Z", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidDateTime(tt.dt)
			if got != tt.expected {
				t.Errorf("IsValidDateTime(%q) = %v, want %v", tt.dt, got, tt.expected)
			}
		})
	}
}

func TestIsValidEnum(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		allowed  []interface{}
		expected bool
	}{
		{"string found", "available", []interface{}{"available", "pending", "sold"}, true},
		{"string not found", "deleted", []interface{}{"available", "pending", "sold"}, false},
		{"int found", 42, []interface{}{1, 42, 99}, true},
		{"int not found", 7, []interface{}{1, 42, 99}, false},
		{"bool found", true, []interface{}{true, false}, true},
		{"bool not found", false, []interface{}{true}, false},
		{"empty allowed", "x", []interface{}{}, false},
		{"nil value", nil, []interface{}{"a"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidEnum(tt.value, tt.allowed)
			if got != tt.expected {
				t.Errorf("IsValidEnum(%v, %v) = %v, want %v", tt.value, tt.allowed, got, tt.expected)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"non-empty string", "hello", true},
		{"empty string", "", false},
		{"non-zero int", 42, true},
		{"zero int", 0, false},
		{"non-zero int64", int64(1), true},
		{"zero int64", int64(0), false},
		{"non-zero float64", 1.5, true},
		{"zero float64", 0.0, false},
		{"true bool", true, true},
		{"false bool", false, false},
		{"nil", nil, false},
		{"non-nil struct", struct{}{}, true},
		{"non-nil slice", []int{1}, true},
		{"nil slice", ([]int)(nil), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValid(tt.value)
			if got != tt.expected {
				t.Errorf("IsValid(%v) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}
