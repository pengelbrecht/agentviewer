// Package main provides REST API handlers.
package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// CreateTabRequest is the request body for creating a tab.
type CreateTabRequest struct {
	ID       string   `json:"id,omitempty"`
	Title    string   `json:"title"`
	Type     string   `json:"type"`
	Content  string   `json:"content,omitempty"`
	File     string   `json:"file,omitempty"`
	Language string   `json:"language,omitempty"`
	Diff     *DiffReq `json:"diff,omitempty"`
}

// DiffReq holds diff-specific request parameters.
type DiffReq struct {
	Left       string `json:"left,omitempty"`
	Right      string `json:"right,omitempty"`
	LeftLabel  string `json:"leftLabel,omitempty"`
	RightLabel string `json:"rightLabel,omitempty"`
	Unified    string `json:"unified,omitempty"`
	Language   string `json:"language,omitempty"`
}

// CreateTabResponse is the response for creating a tab.
type CreateTabResponse struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Type    string `json:"type"`
	Created bool   `json:"created"`
}

// ListTabsResponse is the response for listing tabs.
type ListTabsResponse struct {
	Tabs []*TabSummary `json:"tabs"`
}

// TabSummary is a summary of a tab for listing.
type TabSummary struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	Active bool   `json:"active"`
}

// StatusResponse is the response for server status.
type StatusResponse struct {
	Version string `json:"version"`
	Tabs    int    `json:"tabs"`
	Uptime  int64  `json:"uptime"`
}

// ErrorResponse is a standard error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Version is the application version.
var Version = "0.1.0"

// ValidTabTypes is the set of valid tab types.
var ValidTabTypes = map[string]bool{
	"":         true, // empty means auto-detect
	"markdown": true,
	"code":     true,
	"diff":     true,
	"image":    true,
}

// handleCreateTab handles POST /api/tabs.
func (s *Server) handleCreateTab(w http.ResponseWriter, r *http.Request) {
	var req CreateTabRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate tab type
	if !ValidTabTypes[req.Type] {
		writeError(w, http.StatusBadRequest, "Invalid type: must be 'markdown', 'code', 'diff', or 'image'")
		return
	}

	// Validate diff type has diff data
	if req.Type == "diff" && req.Diff == nil && req.Content == "" && req.File == "" {
		writeError(w, http.StatusBadRequest, "Diff type requires 'diff' object, 'content', or 'file'")
		return
	}

	// Determine content
	content := req.Content
	if req.File != "" && content == "" {
		var err error
		// For image files, read as base64 data URL
		if IsImageFile(req.File) {
			content, err = ReadImageAsDataURL(req.File)
		} else {
			content, err = ReadFileContent(req.File)
		}
		if err != nil {
			writeError(w, http.StatusBadRequest, "Cannot read file: "+err.Error())
			return
		}
	}

	// Handle diff type
	var diffMeta *DiffMeta
	if req.Type == "diff" && req.Diff != nil {
		diffMeta = &DiffMeta{
			LeftLabel:  req.Diff.LeftLabel,
			RightLabel: req.Diff.RightLabel,
			Language:   req.Diff.Language,
		}

		if req.Diff.Unified != "" {
			content = req.Diff.Unified
		} else if req.Diff.Left != "" && req.Diff.Right != "" {
			// Read both files and create diff
			leftContent, err := ReadFileContent(req.Diff.Left)
			if err != nil {
				writeError(w, http.StatusBadRequest, "Cannot read left file: "+err.Error())
				return
			}
			rightContent, err := ReadFileContent(req.Diff.Right)
			if err != nil {
				writeError(w, http.StatusBadRequest, "Cannot read right file: "+err.Error())
				return
			}
			content = CreateUnifiedDiff(req.Diff.Left, req.Diff.Right, leftContent, rightContent)
		}
	}

	// Detect content type if not specified
	tabType := TabType(req.Type)
	if tabType == "" {
		tabType = DetectContentType(req.File, content)
	}

	// Detect language for code if not specified
	language := req.Language
	if tabType == TabTypeCode && language == "" {
		language = DetectLanguage(req.File, content)
	}

	// Determine source path for file-based tabs (enables auto-reload)
	sourcePath := ""
	if req.File != "" {
		sourcePath = req.File
	}

	tab := &Tab{
		ID:         req.ID,
		Title:      req.Title,
		Type:       tabType,
		Content:    content,
		Language:   language,
		DiffMeta:   diffMeta,
		SourcePath: sourcePath,
	}

	tab, created := s.state.CreateTab(tab)

	// Register file for watching if it has a source path
	if tab.SourcePath != "" && s.fileWatcher != nil {
		if err := s.fileWatcher.Add(tab.SourcePath, tab.ID); err != nil {
			// Log but don't fail - watching is optional
			// The tab was created successfully
			_ = err // ignore error
		}
	}

	// Broadcast to WebSocket clients
	msgType := "tab_updated"
	if created {
		msgType = "tab_created"
	}
	s.hub.Broadcast(WSMessage{Type: msgType, Tab: tab})

	writeJSON(w, http.StatusOK, CreateTabResponse{
		ID:      tab.ID,
		Title:   tab.Title,
		Type:    string(tab.Type),
		Created: created,
	})
}

// handleListTabs handles GET /api/tabs.
func (s *Server) handleListTabs(w http.ResponseWriter, r *http.Request) {
	tabs := s.state.ListTabs()
	summaries := make([]*TabSummary, len(tabs))
	for i, tab := range tabs {
		summaries[i] = &TabSummary{
			ID:     tab.ID,
			Title:  tab.Title,
			Type:   string(tab.Type),
			Active: tab.Active,
		}
	}
	writeJSON(w, http.StatusOK, ListTabsResponse{Tabs: summaries})
}

// handleGetTab handles GET /api/tabs/{id}.
func (s *Server) handleGetTab(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tab, exists := s.state.GetTab(id)
	if !exists {
		writeError(w, http.StatusNotFound, "Tab not found")
		return
	}
	writeJSON(w, http.StatusOK, tab)
}

// handleDeleteTab handles DELETE /api/tabs/{id}.
func (s *Server) handleDeleteTab(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Remove from file watcher before deleting
	if s.fileWatcher != nil {
		s.fileWatcher.Remove(id)
	}

	if !s.state.DeleteTab(id) {
		writeError(w, http.StatusNotFound, "Tab not found")
		return
	}

	// Broadcast to WebSocket clients
	s.hub.Broadcast(WSMessage{Type: "tab_deleted", ID: id})

	w.WriteHeader(http.StatusNoContent)
}

// handleActivateTab handles POST /api/tabs/{id}/activate.
func (s *Server) handleActivateTab(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.state.SetActive(id) {
		writeError(w, http.StatusNotFound, "Tab not found")
		return
	}

	// Broadcast to WebSocket clients
	s.hub.Broadcast(WSMessage{Type: "tab_activated", ID: id})

	w.WriteHeader(http.StatusNoContent)
}

// handleClearTabs handles DELETE /api/tabs.
func (s *Server) handleClearTabs(w http.ResponseWriter, r *http.Request) {
	// Clear file watches first
	if s.fileWatcher != nil {
		s.fileWatcher.Clear()
	}

	s.state.Clear()

	// Broadcast to WebSocket clients
	s.hub.Broadcast(WSMessage{Type: "tabs_cleared"})

	w.WriteHeader(http.StatusNoContent)
}

// handleStatus handles GET /api/status.
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	uptime := int64(time.Since(StartTime).Seconds())
	writeJSON(w, http.StatusOK, StatusResponse{
		Version: Version,
		Tabs:    s.state.TabCount(),
		Uptime:  uptime,
	})
}

// handleWebSocket handles WebSocket connections.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	ServeWS(s.hub, w, r, func(data []byte) {
		var msg WSMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return
		}

		// Handle client messages
		switch msg.Type {
		case "activate_tab":
			if msg.ID != "" {
				s.state.SetActive(msg.ID)
				s.hub.Broadcast(WSMessage{Type: "tab_activated", ID: msg.ID})
			}
		case "close_tab":
			if msg.ID != "" && s.state.DeleteTab(msg.ID) {
				s.hub.Broadcast(WSMessage{Type: "tab_deleted", ID: msg.ID})
			}
		}
	})
}

// handleStatic serves embedded static files.
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	ServeStaticFile(w, r)
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}
