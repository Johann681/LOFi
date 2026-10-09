package main

import "testing"

func TestOriginAllowed(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		allowed []string
		want    bool
	}{
		{
			name:    "configured origin",
			origin:  "https://lo-fi.vercel.app",
			allowed: []string{"https://lo-fi.vercel.app"},
			want:    true,
		},
		{
			name:    "vercel preview subdomain",
			origin:  "https://lo-d13sct9j3-johanns-projects-86b73528.vercel.app",
			allowed: []string{"https://*.vercel.app"},
			want:    true,
		},
		{
			name:    "unrelated suffix",
			origin:  "https://lo-fi.vercel.app.attacker.example",
			allowed: []string{"https://*.vercel.app"},
			want:    false,
		},
		{
			name:    "apex domain",
			origin:  "https://vercel.app",
			allowed: []string{"https://*.vercel.app"},
			want:    false,
		},
		{
			name:    "insecure preview",
			origin:  "http://lo-fi.vercel.app",
			allowed: []string{"https://*.vercel.app"},
			want:    false,
		},
		{
			name:    "nested subdomain",
			origin:  "https://lo.fi.vercel.app",
			allowed: []string{"https://*.vercel.app"},
			want:    false,
		},
		{
			name:    "preview with port",
			origin:  "https://lo-fi.vercel.app:8443",
			allowed: []string{"https://*.vercel.app"},
			want:    false,
		},
		{
			name:    "malformed origin",
			origin:  "not an origin",
			allowed: []string{"https://*.vercel.app"},
			want:    false,
		},
		{
			name:    "exact origin may include port",
			origin:  "http://localhost:3001",
			allowed: []string{"http://localhost:3001"},
			want:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := originAllowed(test.origin, test.allowed); got != test.want {
				t.Errorf("originAllowed(%q, %v) = %t, want %t", test.origin, test.allowed, got, test.want)
			}
		})
	}
}
