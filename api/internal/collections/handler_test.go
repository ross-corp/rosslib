package collections

import (
	"testing"
)

func TestFormatRating(t *testing.T) {
	tests := []struct {
		name     string
		input    *int
		expected string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: "",
		},
		{
			name:     "positive integer",
			input:    ptr(5),
			expected: "5",
		},
		{
			name:     "zero",
			input:    ptr(0),
			expected: "0",
		},
		{
			name:     "negative integer",
			input:    ptr(-1),
			expected: "-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRating(tt.input)
			if result != tt.expected {
				t.Errorf("formatRating(%v) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func ptr(i int) *int {
	return &i
}
