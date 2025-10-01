package utils

import (
	"testing"
)

func TestIDFy(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abc123", "abc123"},
		{"ABC123", "abc123"},
		{"A_b.C-1", "a_b.c-1"},
		{"Hello World!", "hello-world-"},
		{"foo@bar.com", "foo-bar.com"},
		{"", ""},
		{"UPPER_lower-123", "upper_lower-123"},
		{"!@#$%^&*()", "----------"},
		{"MiXeD123", "mixed123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SanitizeName(tt.input)
			if got != tt.expected {
				t.Errorf("IDFy(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}
