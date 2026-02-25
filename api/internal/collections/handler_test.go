package collections

import (
	"testing"
)

func TestDerefStr(t *testing.T) {
	strEmpty := ""
	strHello := "hello"

	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: "",
		},
		{
			name:     "empty string pointer",
			input:    &strEmpty,
			expected: "",
		},
		{
			name:     "non-empty string pointer",
			input:    &strHello,
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := derefStr(tt.input)
			if result != tt.expected {
				t.Errorf("derefStr(%v) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
