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
	httpServer *http.Server
	state      *State
	hub        *Hub
}

// NewServer creates a new Server instance.
func NewServer() *Server {
	state := NewState()
	hub := NewHub()
	return &Server{
		state: state,
		hub:   hub,
	}
}

// Serve starts the HTTP server on the given listener.
func (s *Server) Serve(listener net.Listener) error {
	mux := http.NewServeMux()
	s.setupRoutes(mux)
	s.httpServer = &http.Server{Handler: mux}

	// Start WebSocket hub
	go s.hub.Run()

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
