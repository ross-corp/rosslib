package imports

import "testing"

func TestParseDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"whitespace only", "   ", ""},
		{"standard date with slashes", "2023/10/26", "2023-10-26"},
		{"date with spaces and slashes", " 2023/10/26 ", "2023-10-26"},
		{"date with hyphens", "2023-10-26", "2023-10-26"},
		{"mixed separators", "2023/10-26", "2023-10-26"},
		{"random string with slashes", "hello/world", "hello-world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDate(tt.input)
			if got != tt.expected {
				t.Errorf("parseDate(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
