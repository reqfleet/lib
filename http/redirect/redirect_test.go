// Derived from reqfleet/shibuya; modified for lib.
package redirect

import "testing"

func TestIsValidRedirectURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"empty", "", false},
		{"root", "/", true},
		{"path", "/dashboard", true},
		{"path_with_query", "/dashboard?x=1", true},
		{"relative_path", "dashboard", false},
		{"scheme", "https://example.com/", false},
		{"protocol_relative", "//example.com/", false},
		{"backslash", "/\\evil", false},
		{"control_char", "/a\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidRedirectURL(tt.in); got != tt.want {
				t.Fatalf("IsValidRedirectURL(%q)=%v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
