package project

import (
	"testing"
)

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"My App", "my-app"},
		{"Hello World 123", "hello-world-123"},
		{"UPPER CASE", "upper-case"},
		{"  spaces  everywhere  ", "spaces-everywhere"},
		{"special!@#chars$%^", "special-chars"},
		{"already-kebab", "already-kebab"},
	}

	for _, tt := range tests {
		got := GenerateSlug(tt.input)
		if got != tt.expected {
			t.Errorf("GenerateSlug(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
