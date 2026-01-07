// Package main provides the HTTP server setup and routing for agentviewer.
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

// Server holds the HTTP server and application state.
type Server struct {
	httpServer  *http.Server
	state       *State
	hub         *Hub
	fileWatcher *FileWatcher
}

// NewServer creates a new Server instance.
func NewServer() *Server {
	state := NewState()
	hub := NewHub()
	s := &Server{
		state: state,
		hub:   hub,
	}

	// Initialize file watcher with callbacks
	watcher, err := NewFileWatcherWithCallbacks(FileWatcherCallbacks{
		OnChange: func(path string, tabIDs []string) {
			s.handleFileChange(path, tabIDs)
		},
		OnDelete: func(path string, tabIDs []string) {
			s.handleFileDelete(path, tabIDs)
		},
	})
	if err != nil {
		// Log error but continue without file watching
		fmt.Printf("Warning: file watching disabled: %v\n", err)
	} else {
		s.fileWatcher = watcher
	}

	return s
}

// Serve starts the HTTP server on the given listener.
func (s *Server) Serve(listener net.Listener) error {
	mux := http.NewServeMux()
	s.setupRoutes(mux)
	s.httpServer = &http.Server{Handler: mux}

	// Start WebSocket hub
	go s.hub.Run()

	// Start file watcher if available
	if s.fileWatcher != nil {
		go s.fileWatcher.Run()
	}

	return s.httpServer.Serve(listener)
}

// ListenAndServe starts the HTTP server on the given address.
func (s *Server) ListenAndServe(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return s.Serve(listener)
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	// Stop file watcher first
	if s.fileWatcher != nil {
		s.fileWatcher.Stop()
	}

	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// setupRoutes configures all HTTP routes.
func (s *Server) setupRoutes(mux *http.ServeMux) {
	// API routes
	mux.HandleFunc("POST /api/tabs", s.handleCreateTab)
	mux.HandleFunc("GET /api/tabs", s.handleListTabs)
	mux.HandleFunc("GET /api/tabs/{id}", s.handleGetTab)
	mux.HandleFunc("DELETE /api/tabs/{id}", s.handleDeleteTab)
	mux.HandleFunc("POST /api/tabs/{id}/activate", s.handleActivateTab)
	mux.HandleFunc("DELETE /api/tabs", s.handleClearTabs)
	mux.HandleFunc("GET /api/status", s.handleStatus)

	// WebSocket
	mux.HandleFunc("GET /ws", s.handleWebSocket)

	// Static files (embedded)
	mux.HandleFunc("GET /", s.handleStatic)
}

// OpenBrowser opens the default browser to the given URL.
func OpenBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// StartTime records when the server started.
var StartTime = time.Now()

// handleFileChange is called when a watched file changes.
// It re-reads the file content, updates affected tabs, and broadcasts updates.
func (s *Server) handleFileChange(path string, tabIDs []string) {
	// Re-read the file content
	content, err := ReadFileContent(path)
	if err != nil {
		// File might have been deleted or become unreadable
		// Log but don't remove the watch - file might come back
		fmt.Printf("Warning: cannot read changed file %s: %v\n", path, err)
		return
	}

	// Update each tab that watches this file
	for _, tabID := range tabIDs {
		tab := s.state.UpdateTabContent(tabID, content)
		if tab != nil {
			// Broadcast the update to all connected clients
			s.hub.Broadcast(WSMessage{Type: "tab_updated", Tab: tab})
		}
	}
}

// handleFileDelete is called when a watched file is deleted or renamed.
// It marks affected tabs as stale and broadcasts updates.
func (s *Server) handleFileDelete(path string, tabIDs []string) {
	// Mark each tab that watches this file as stale
	for _, tabID := range tabIDs {
		tab := s.state.MarkTabStale(tabID)
		if tab != nil {
			// Broadcast the update to all connected clients
			s.hub.Broadcast(WSMessage{Type: "tab_stale", Tab: tab})
		}
	}
}
