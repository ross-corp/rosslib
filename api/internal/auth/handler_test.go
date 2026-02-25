package auth

import (
	"strings"
	"testing"
)

func TestUsernameValidation(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{"valid simple", "user", true},
		{"valid alphanumeric", "user123", true},
		{"valid with hyphen", "user-name", true},
		{"valid min length", "u", true},
		{"valid max length", strings.Repeat("a", 40), true},
		{"invalid empty", "", false},
		{"invalid space", "user name", false},
		{"invalid uppercase", "User", false},
		{"invalid special char", "user@name", false},
		{"invalid underscore", "user_name", false},
		{"invalid too long", strings.Repeat("a", 41), false},
		{"valid starts with hyphen", "-user", true},
		{"valid ends with hyphen", "user-", true},
		{"valid double hyphen", "user--name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := usernameRe.MatchString(tt.username)
			if got != tt.want {
				t.Errorf("usernameRe.MatchString(%q) = %v, want %v", tt.username, got, tt.want)
			}
		})
	}
}
