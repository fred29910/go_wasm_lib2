package generator

import (
	"testing"
)

func TestToGoName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "Value"},
		{"name", "Name"},
		{"pet_id", "PetId"},
		{"pet-id", "PetId"},
		{"pet.id", "PetId"},
		{"Pet", "Pet"},
		{"petId", "PetId"},
		{"123abc", "N123abc"},
		{"_private", "Private"},
		{"a", "A"},
		{"AB", "AB"},
		{"a_b_c", "ABC"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToGoName(tt.input)
			if got != tt.expected {
				t.Errorf("ToGoName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToTSName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "value"},
		{"name", "name"},
		{"pet_id", "petId"},
		{"pet-id", "petId"},
		{"Pet", "pet"},
		{"createPet", "createPet"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToTSName(tt.input)
			if got != tt.expected {
				t.Errorf("ToTSName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToPrivateGoName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "value"},
		{"Name", "name"},
		{"PetId", "petId"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToPrivateGoName(tt.input)
			if got != tt.expected {
				t.Errorf("ToPrivateGoName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
