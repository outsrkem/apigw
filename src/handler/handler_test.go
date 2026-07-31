package handler

import "testing"

func TestRewritePath(t *testing.T) {
	tests := []struct {
		name       string
		rawPath    string
		reqUri     string
		backendUri string
		mode       string
		expected   string
	}{
		{"EXACT mode", "/api/user", "/api", "http://backend", "EXACT", "http://backend"},
		{"prefix matching", "/api/user/123", "/api", "http://backend", "", "http://backend/user/123"},
		{"remove redundant slashes", "/api//user", "/api", "http://backend", "", "http://backend/user"},
		{"trailing slash reserved", "/api/user/", "/api", "http://backend/", "", "http://backend/user/"},
		{"empty raw path", "", "/api", "http://backend", "", "http://backend"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RewritePath(tt.rawPath, tt.reqUri, tt.backendUri, tt.mode)
			if result != tt.expected {
				t.Errorf("got: %s, expected: %s", result, tt.expected)
			}
		})
	}
}
