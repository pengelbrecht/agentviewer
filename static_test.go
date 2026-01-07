// Package main provides tests for embedded static file serving.
package main

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebFSEmbedding(t *testing.T) {
	// Verify the embedded filesystem contains expected files
	webFS := GetWebFS()

	tests := []struct {
		path     string
		wantFile bool
	}{
		{"web/index.html", true},
		{"web/style.css", true},
		{"web/app.js", true},
		{"web/vendor/README.md", true},
		{"web/nonexistent.html", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			_, err := fs.ReadFile(webFS, tt.path)
			gotFile := err == nil

			if gotFile != tt.wantFile {
				if tt.wantFile {
					t.Errorf("expected file %q to exist, but got error: %v", tt.path, err)
				} else {
					t.Errorf("expected file %q to not exist, but it does", tt.path)
				}
			}
		})
	}
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		path     string
		wantType string
	}{
		{"web/index.html", "text/html; charset=utf-8"},
		{"web/style.css", "text/css; charset=utf-8"},
		{"web/app.js", "application/javascript; charset=utf-8"},
		{"web/data.json", "application/json; charset=utf-8"},
		{"web/image.svg", "image/svg+xml"},
		{"web/icon.png", "image/png"},
		{"web/favicon.ico", "image/x-icon"},
		{"web/unknown.xyz", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := getContentType(tt.path)
			if got != tt.wantType {
				t.Errorf("getContentType(%q) = %q, want %q", tt.path, got, tt.wantType)
			}
		})
	}
}

func TestServeStaticFile(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantStatus   int
		wantType     string
		wantContains string
	}{
		{
			name:         "serve index.html at root",
			path:         "/",
			wantStatus:   http.StatusOK,
			wantType:     "text/html; charset=utf-8",
			wantContains: "<!DOCTYPE html>",
		},
		{
			name:         "serve index.html explicitly",
			path:         "/index.html",
			wantStatus:   http.StatusOK,
			wantType:     "text/html; charset=utf-8",
			wantContains: "<!DOCTYPE html>",
		},
		{
			name:         "serve CSS file",
			path:         "/style.css",
			wantStatus:   http.StatusOK,
			wantType:     "text/css; charset=utf-8",
			wantContains: "",
		},
		{
			name:         "serve JS file",
			path:         "/app.js",
			wantStatus:   http.StatusOK,
			wantType:     "application/javascript; charset=utf-8",
			wantContains: "",
		},
		{
			name:         "serve vendor file",
			path:         "/vendor/README.md",
			wantStatus:   http.StatusOK,
			wantType:     "application/octet-stream", // .md not in our list
			wantContains: "Vendor Libraries",
		},
		{
			name:         "SPA fallback for path without extension",
			path:         "/some/spa/route",
			wantStatus:   http.StatusOK,
			wantType:     "text/html; charset=utf-8",
			wantContains: "<!DOCTYPE html>",
		},
		{
			name:       "404 for nonexistent file with extension",
			path:       "/nonexistent.js",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			ServeStaticFile(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			if tt.wantType != "" {
				gotType := resp.Header.Get("Content-Type")
				if gotType != tt.wantType {
					t.Errorf("Content-Type = %q, want %q", gotType, tt.wantType)
				}
			}

			if tt.wantContains != "" {
				body := w.Body.String()
				if !strings.Contains(body, tt.wantContains) {
					t.Errorf("body does not contain %q, got: %s", tt.wantContains, body[:min(100, len(body))])
				}
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
