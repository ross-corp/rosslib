package collections

import (
	"testing"
	"time"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple lowercase", "science", "science"},
		{"simple mixed case", "Science", "science"},
		{"spaces to hyphens", "Science Fiction", "science-fiction"},
		{"special chars removed", "Science & Fiction!", "science-fiction"},
		{"multiple spaces", "Science   Fiction", "science-fiction"},
		{"multiple hyphens", "a--b", "a-b"},
		{"leading hyphens", "-test", "test"},
		{"trailing hyphens", "test-", "test"},
		{"leading spaces", " test", "test"},
		{"trailing spaces", "test ", "test"},
		{"numbers", "Zone 51", "zone-51"},
		{"empty string", "", ""},
		{"only special chars", "!@#$", ""},
		{"mixed special and spaces", "  &  ", ""},
		{"complex case", "My Cool Collection (2024)", "my-cool-collection-2024"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slugify(tt.input)
			if got != tt.expected {
				t.Errorf("slugify(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTagSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple path", "Science/Fiction", "science/fiction"},
		{"mixed case path", "Science/Fiction/Space Opera", "science/fiction/space-opera"},
		{"special chars in segments", "Sci-Fi & Fantasy/Space Opera!", "sci-fi-fantasy/space-opera"},
		{"empty segments removed", "a//b", "a/b"},
		{"leading slash", "/a/b", "a/b"},
		{"trailing slash", "a/b/", "a/b"},
		{"complex path", "My Books/Favorites (2024)/Best-Of", "my-books/favorites-2024/best-of"},
		{"empty string", "", ""},
		{"root slash only", "/", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tagSlugify(tt.input)
			if got != tt.expected {
				t.Errorf("tagSlugify(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDerefStr(t *testing.T) {
	strEmpty := ""
	strHello := "hello"

	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{"nil input", nil, ""},
		{"empty string pointer", &strEmpty, ""},
		{"non-empty string pointer", &strHello, "hello"},
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

func TestFormatRating(t *testing.T) {
	tests := []struct {
		name     string
		input    *int
		expected string
	}{
		{"nil input", nil, ""},
		{"positive integer", intPtr(5), "5"},
		{"zero", intPtr(0), "0"},
		{"negative integer", intPtr(-1), "-1"},
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
		{"nil input", nil, ""},
		{"valid time", &validTime, "2023-10-27"},
		{"zero time", &zeroTime, "0001-01-01"},
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
