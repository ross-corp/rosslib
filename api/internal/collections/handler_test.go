package collections

import (
	"testing"
	"time"
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
			input:    intPtr(5),
			expected: "5",
		},
		{
			name:     "zero",
			input:    intPtr(0),
			expected: "0",
		},
		{
			name:     "negative integer",
			input:    intPtr(-1),
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

func intPtr(i int) *int {
	return &i
}

func TestFormatDate(t *testing.T) {
	validTime := time.Date(2023, 10, 27, 12, 0, 0, 0, time.UTC)
	zeroTime := time.Time{}

	tests := []struct {
		name     string
		input    *time.Time
		expected string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: "",
		},
		{
			name:     "valid time",
			input:    &validTime,
			expected: "2023-10-27",
		},
		{
			name:     "zero time",
			input:    &zeroTime,
			expected: "0001-01-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDate(tt.input)
			if got != tt.expected {
				t.Errorf("formatDate() = %q, want %q", got, tt.expected)
			}
		})
	}
}
