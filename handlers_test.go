package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestServer() *Server {
	srv := NewServer()
	// Start the hub in a goroutine (needed for WebSocket broadcasts)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-ctx.Done()
	}()
	_ = cancel // Will be called when test ends
	go srv.hub.Run()
	return srv
}

func TestCreateTab_InvalidJSON(t *testing.T) {
	srv := setupTestServer()

	req := httptest.NewRequest("POST", "/api/tabs", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTab(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Error != "Invalid JSON" {
		t.Errorf("expected 'Invalid JSON' error, got %q", resp.Error)
	}
}

func TestCreateTab_InvalidType(t *testing.T) {
	srv := setupTestServer()

	body := `{"title": "Test", "type": "invalid", "content": "test"}`
	req := httptest.NewRequest("POST", "/api/tabs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTab(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Error != "Invalid type: must be 'markdown', 'code', or 'diff'" {
		t.Errorf("unexpected error: %q", resp.Error)
	}
}

func TestCreateTab_DiffWithoutData(t *testing.T) {
	srv := setupTestServer()

	body := `{"title": "Test", "type": "diff"}`
	req := httptest.NewRequest("POST", "/api/tabs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTab(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Error != "Diff type requires 'diff' object, 'content', or 'file'" {
		t.Errorf("unexpected error: %q", resp.Error)
	}
}

func TestCreateTab_DiffWithContent(t *testing.T) {
	srv := setupTestServer()

	// Diff with inline content (unified diff string) should succeed
	body := `{"title": "Test", "type": "diff", "content": "--- a/file\n+++ b/file\n@@ -1 +1 @@\n-old\n+new"}`
	req := httptest.NewRequest("POST", "/api/tabs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTab(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateTab_ValidMarkdown(t *testing.T) {
	srv := setupTestServer()

	body := `{"title": "Test", "type": "markdown", "content": "# Hello"}`
	req := httptest.NewRequest("POST", "/api/tabs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTab(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp CreateTabResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Type != "markdown" {
		t.Errorf("expected type 'markdown', got %q", resp.Type)
	}
	if !resp.Created {
		t.Error("expected Created to be true")
	}
}

func TestCreateTab_AutoDetectType(t *testing.T) {
	srv := setupTestServer()

	// Empty type should auto-detect
	body := `{"title": "Test", "content": "# Hello World"}`
	req := httptest.NewRequest("POST", "/api/tabs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTab(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp CreateTabResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	// Auto-detect should find markdown (# heading)
	if resp.Type != "markdown" {
		t.Errorf("expected auto-detected type 'markdown', got %q", resp.Type)
	}
}

func TestCreateTab_FileNotFound(t *testing.T) {
	srv := setupTestServer()

	body := `{"title": "Test", "type": "code", "file": "/nonexistent/file.go"}`
	req := httptest.NewRequest("POST", "/api/tabs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTab(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Error == "" {
		t.Error("expected error message about file not found")
	}
}

func TestGetTab_NotFound(t *testing.T) {
	srv := setupTestServer()

	req := httptest.NewRequest("GET", "/api/tabs/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	srv.handleGetTab(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Error != "Tab not found" {
		t.Errorf("expected 'Tab not found' error, got %q", resp.Error)
	}
}

func TestDeleteTab_NotFound(t *testing.T) {
	srv := setupTestServer()

	req := httptest.NewRequest("DELETE", "/api/tabs/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	srv.handleDeleteTab(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Error != "Tab not found" {
		t.Errorf("expected 'Tab not found' error, got %q", resp.Error)
	}
}

func TestActivateTab_NotFound(t *testing.T) {
	srv := setupTestServer()

	req := httptest.NewRequest("POST", "/api/tabs/nonexistent/activate", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	srv.handleActivateTab(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Error != "Tab not found" {
		t.Errorf("expected 'Tab not found' error, got %q", resp.Error)
	}
}

func TestListTabs_Empty(t *testing.T) {
	srv := setupTestServer()

	req := httptest.NewRequest("GET", "/api/tabs", nil)
	w := httptest.NewRecorder()

	srv.handleListTabs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ListTabsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(resp.Tabs) != 0 {
		t.Errorf("expected 0 tabs, got %d", len(resp.Tabs))
	}
}

func TestClearTabs(t *testing.T) {
	srv := setupTestServer()

	// Create a tab first
	srv.state.CreateTab(&Tab{
		Title:   "Test",
		Type:    TabTypeMarkdown,
		Content: "# Hello",
	})

	if srv.state.TabCount() != 1 {
		t.Fatalf("expected 1 tab, got %d", srv.state.TabCount())
	}

	req := httptest.NewRequest("DELETE", "/api/tabs", nil)
	w := httptest.NewRecorder()

	srv.handleClearTabs(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	if srv.state.TabCount() != 0 {
		t.Errorf("expected 0 tabs after clear, got %d", srv.state.TabCount())
	}
}

func TestStatus(t *testing.T) {
	srv := setupTestServer()

	// Create some tabs
	srv.state.CreateTab(&Tab{Title: "Tab1", Type: TabTypeMarkdown, Content: "test"})
	srv.state.CreateTab(&Tab{Title: "Tab2", Type: TabTypeCode, Content: "test"})

	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	srv.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp StatusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Version != Version {
		t.Errorf("expected version %q, got %q", Version, resp.Version)
	}
	if resp.Tabs != 2 {
		t.Errorf("expected 2 tabs, got %d", resp.Tabs)
	}
	if resp.Uptime < 0 {
		t.Errorf("expected non-negative uptime, got %d", resp.Uptime)
	}
}
