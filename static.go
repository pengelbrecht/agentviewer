// Package main provides embedded static file serving.
package main

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed web
var webFS embed.FS

// ServeStaticFile serves embedded static files.
func ServeStaticFile(w http.ResponseWriter, r *http.Request) {
	// Clean the path and default to index.html
	urlPath := strings.TrimPrefix(r.URL.Path, "/")
	if urlPath == "" {
		urlPath = "index.html"
	}

	// Prepend "web/" to access embedded files
	filePath := path.Join("web", urlPath)

	// Try to read the file
	data, err := webFS.ReadFile(filePath)
	if err != nil {
		// If not found and not a specific file request, serve index.html (SPA fallback)
		if !strings.Contains(urlPath, ".") {
			data, err = webFS.ReadFile("web/index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			filePath = "web/index.html"
		} else {
			http.NotFound(w, r)
			return
		}
	}

	// Set content type based on extension
	contentType := getContentType(filePath)
	w.Header().Set("Content-Type", contentType)
	w.Write(data)
}

// getContentType returns the MIME type for a file path.
func getContentType(filePath string) string {
	ext := path.Ext(filePath)
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".ico":
		return "image/x-icon"
	case ".woff2":
		return "font/woff2"
	case ".woff":
		return "font/woff"
	case ".ttf":
		return "font/ttf"
	default:
		return "application/octet-stream"
	}
}

// GetWebFS returns the embedded filesystem for testing.
func GetWebFS() fs.FS {
	return webFS
}
