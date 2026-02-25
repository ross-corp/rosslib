package storage

import (
	"testing"
)

func TestImageExtension(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantExt     string
		wantOk      bool
	}{
		{"JPEG", "image/jpeg", ".jpg", true},
		{"PNG", "image/png", ".png", true},
		{"GIF", "image/gif", ".gif", true},
		{"WebP", "image/webp", ".webp", true},
		{"PDF", "application/pdf", "", false},
		{"Empty", "", "", false},
		{"Unknown", "application/unknown", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExt, gotOk := imageExtension(tt.contentType)
			if gotExt != tt.wantExt {
				t.Errorf("imageExtension(%q) gotExt = %v, want %v", tt.contentType, gotExt, tt.wantExt)
			}
			if gotOk != tt.wantOk {
				t.Errorf("imageExtension(%q) gotOk = %v, want %v", tt.contentType, gotOk, tt.wantOk)
			}
		})
	}
}
