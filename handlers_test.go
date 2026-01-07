package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

// getTestDataPath returns the absolute path to a file in testdata directory.
func getTestDataPath(t *testing.T, filename string) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	return filepath.Join(cwd, "testdata", filename)
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

// TestCreateTab_WithExplicitID tests creating a tab with a specific ID.
func TestCreateTab_WithExplicitID(t *testing.T) {
	srv := setupTestServer()

	body := `{"id": "my-custom-id", "title": "Custom Tab", "type": "markdown", "content": "# Custom"}`
	req := httptest.NewRequest("POST", "/api/tabs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTab(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CreateTabResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.ID != "my-custom-id" {
		t.Errorf("expected ID 'my-custom-id', got %q", resp.ID)
	}
	if !resp.Created {
		t.Error("expected Created to be true for new tab")
	}

	// Verify tab exists in state
	tab, exists := srv.state.GetTab("my-custom-id")
	if !exists {
		t.Error("tab was not created in state")
	}
	if tab.Title != "Custom Tab" {
		t.Errorf("expected title 'Custom Tab', got %q", tab.Title)
	}
}

// TestCreateTab_UpdateExisting tests updating an existing tab by ID.
func TestCreateTab_UpdateExisting(t *testing.T) {
	srv := setupTestServer()

	// Create initial tab
	body1 := `{"id": "update-test", "title": "Original", "type": "markdown", "content": "# Original"}`
	req1 := httptest.NewRequest("POST", "/api/tabs", bytes.NewBufferString(body1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	srv.handleCreateTab(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("failed to create initial tab: %d", w1.Code)
	}

	// Update the tab with same ID
	body2 := `{"id": "update-test", "title": "Updated", "type": "markdown", "content": "# Updated"}`
	req2 := httptest.NewRequest("POST", "/api/tabs", bytes.NewBufferString(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	srv.handleCreateTab(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200 for update, got %d: %s", w2.Code, w2.Body.String())
	}

	var resp CreateTabResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Created {
		t.Error("expected Created to be false for update")
	}

	// Verify content was updated
	tab, _ := srv.state.GetTab("update-test")
	if tab.Title != "Updated" {
		t.Errorf("expected title 'Updated', got %q", tab.Title)
	}
	if tab.Content != "# Updated" {
		t.Errorf("expected content '# Updated', got %q", tab.Content)
	}
}

// TestCreateTab_CodeWithLanguage tests creating a code tab with explicit language.
func TestCreateTab_CodeWithLanguage(t *testing.T) {
	srv := setupTestServer()

	body := `{"title": "Go Code", "type": "code", "content": "package main\n\nfunc main() {}", "language": "go"}`
	req := httptest.NewRequest("POST", "/api/tabs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTab(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CreateTabResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Type != "code" {
		t.Errorf("expected type 'code', got %q", resp.Type)
	}

	// Verify language was set
	tab, _ := srv.state.GetTab(resp.ID)
	if tab.Language != "go" {
		t.Errorf("expected language 'go', got %q", tab.Language)
	}
}

// TestCreateTab_FromFile tests creating a tab from a file.
func TestCreateTab_FromFile(t *testing.T) {
	srv := setupTestServer()

	filePath := getTestDataPath(t, "sample.go")

	bodyObj := map[string]interface{}{
		"title": "File Tab",
		"type":  "code",
		"file":  filePath,
	}
	bodyBytes, _ := json.Marshal(bodyObj)
	req := httptest.NewRequest("POST", "/api/tabs", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTab(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CreateTabResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Type != "code" {
		t.Errorf("expected type 'code', got %q", resp.Type)
	}

	// Verify content was read from file
	tab, _ := srv.state.GetTab(resp.ID)
	if !strings.Contains(tab.Content, "package main") {
		t.Error("expected content to contain 'package main' from file")
	}
}

// TestCreateTab_FromMarkdownFile tests creating a tab from a markdown file.
func TestCreateTab_FromMarkdownFile(t *testing.T) {
	srv := setupTestServer()

	filePath := getTestDataPath(t, "sample.md")

	bodyObj := map[string]interface{}{
		"title": "Markdown File",
		"file":  filePath,
	}
	bodyBytes, _ := json.Marshal(bodyObj)
	req := httptest.NewRequest("POST", "/api/tabs", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTab(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CreateTabResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Auto-detect should pick markdown for .md file
	if resp.Type != "markdown" {
		t.Errorf("expected auto-detected type 'markdown', got %q", resp.Type)
	}
}

// TestCreateTab_DiffWithUnified tests creating a diff tab with unified diff string.
func TestCreateTab_DiffWithUnified(t *testing.T) {
	srv := setupTestServer()

	unifiedDiff := `--- a/old.txt
+++ b/new.txt
@@ -1,3 +1,3 @@
 line1
-old line
+new line
 line3`

	bodyObj := map[string]interface{}{
		"title": "Unified Diff",
		"type":  "diff",
		"diff": map[string]string{
			"unified": unifiedDiff,
		},
	}
	bodyBytes, _ := json.Marshal(bodyObj)
	req := httptest.NewRequest("POST", "/api/tabs", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTab(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CreateTabResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Type != "diff" {
		t.Errorf("expected type 'diff', got %q", resp.Type)
	}

	// Verify content is the unified diff
	tab, _ := srv.state.GetTab(resp.ID)
	if !strings.Contains(tab.Content, "old line") {
		t.Error("expected content to contain unified diff")
	}
}

// TestCreateTab_DiffFromFiles tests creating a diff by comparing two files.
func TestCreateTab_DiffFromFiles(t *testing.T) {
	srv := setupTestServer()

	leftPath := getTestDataPath(t, "version1.txt")
	rightPath := getTestDataPath(t, "version2.txt")

	bodyObj := map[string]interface{}{
		"title": "File Comparison",
		"type":  "diff",
		"diff": map[string]string{
			"left":       leftPath,
			"right":      rightPath,
			"leftLabel":  "Version 1",
			"rightLabel": "Version 2",
		},
	}
	bodyBytes, _ := json.Marshal(bodyObj)
	req := httptest.NewRequest("POST", "/api/tabs", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTab(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CreateTabResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Verify diff was computed
	tab, _ := srv.state.GetTab(resp.ID)
	if !strings.Contains(tab.Content, "Old content") || !strings.Contains(tab.Content, "New content") {
		t.Errorf("expected diff content, got: %s", tab.Content)
	}
	if tab.DiffMeta == nil {
		t.Error("expected DiffMeta to be set")
	} else {
		if tab.DiffMeta.LeftLabel != "Version 1" {
			t.Errorf("expected LeftLabel 'Version 1', got %q", tab.DiffMeta.LeftLabel)
		}
	}
}

// TestCreateTab_DiffLeftFileNotFound tests error when left file doesn't exist.
func TestCreateTab_DiffLeftFileNotFound(t *testing.T) {
	srv := setupTestServer()

	rightPath := getTestDataPath(t, "version1.txt")

	bodyObj := map[string]interface{}{
		"title": "Broken Diff",
		"type":  "diff",
		"diff": map[string]string{
			"left":  "/nonexistent/left.txt",
			"right": rightPath,
		},
	}
	bodyBytes, _ := json.Marshal(bodyObj)
	req := httptest.NewRequest("POST", "/api/tabs", bytes.NewReader(bodyBytes))
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
	if !strings.Contains(resp.Error, "left file") {
		t.Errorf("expected error about left file, got: %q", resp.Error)
	}
}

// TestCreateTab_DiffRightFileNotFound tests error when right file doesn't exist.
func TestCreateTab_DiffRightFileNotFound(t *testing.T) {
	srv := setupTestServer()

	leftPath := getTestDataPath(t, "version1.txt")

	bodyObj := map[string]interface{}{
		"title": "Broken Diff",
		"type":  "diff",
		"diff": map[string]string{
			"left":  leftPath,
			"right": "/nonexistent/right.txt",
		},
	}
	bodyBytes, _ := json.Marshal(bodyObj)
	req := httptest.NewRequest("POST", "/api/tabs", bytes.NewReader(bodyBytes))
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
	if !strings.Contains(resp.Error, "right file") {
		t.Errorf("expected error about right file, got: %q", resp.Error)
	}
}

// TestCreateTab_AutoDetectCodeByContent tests auto-detecting code type.
func TestCreateTab_AutoDetectCodeByContent(t *testing.T) {
	srv := setupTestServer()

	// Code-like content without explicit type - should detect based on file extension pattern
	body := `{"title": "Auto Code", "content": "function hello() { return 'world'; }", "file": "script.js"}`
	req := httptest.NewRequest("POST", "/api/tabs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTab(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CreateTabResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	// Should detect code type based on .js filename
	if resp.Type != "code" {
		t.Errorf("expected auto-detected type 'code', got %q", resp.Type)
	}
}

// TestGetTab_Success tests successfully getting an existing tab.
func TestGetTab_Success(t *testing.T) {
	srv := setupTestServer()

	// Create a tab first
	tab, _ := srv.state.CreateTab(&Tab{
		ID:      "get-test",
		Title:   "Test Tab",
		Type:    TabTypeMarkdown,
		Content: "# Hello",
	})

	req := httptest.NewRequest("GET", "/api/tabs/get-test", nil)
	req.SetPathValue("id", "get-test")
	w := httptest.NewRecorder()

	srv.handleGetTab(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp Tab
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.ID != tab.ID {
		t.Errorf("expected ID %q, got %q", tab.ID, resp.ID)
	}
	if resp.Title != "Test Tab" {
		t.Errorf("expected title 'Test Tab', got %q", resp.Title)
	}
	if resp.Content != "# Hello" {
		t.Errorf("expected content '# Hello', got %q", resp.Content)
	}
}

// TestDeleteTab_Success tests successfully deleting a tab.
func TestDeleteTab_Success(t *testing.T) {
	srv := setupTestServer()

	// Create a tab first
	srv.state.CreateTab(&Tab{
		ID:    "delete-test",
		Title: "To Delete",
		Type:  TabTypeMarkdown,
	})

	if srv.state.TabCount() != 1 {
		t.Fatalf("expected 1 tab, got %d", srv.state.TabCount())
	}

	req := httptest.NewRequest("DELETE", "/api/tabs/delete-test", nil)
	req.SetPathValue("id", "delete-test")
	w := httptest.NewRecorder()

	srv.handleDeleteTab(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	if srv.state.TabCount() != 0 {
		t.Errorf("expected 0 tabs after delete, got %d", srv.state.TabCount())
	}

	// Verify tab no longer exists
	_, exists := srv.state.GetTab("delete-test")
	if exists {
		t.Error("tab should not exist after deletion")
	}
}

// TestActivateTab_Success tests successfully activating a tab.
func TestActivateTab_Success(t *testing.T) {
	srv := setupTestServer()

	// Create two tabs
	srv.state.CreateTab(&Tab{ID: "tab1", Title: "Tab 1", Type: TabTypeMarkdown})
	srv.state.CreateTab(&Tab{ID: "tab2", Title: "Tab 2", Type: TabTypeMarkdown})

	// First tab should be active by default
	if srv.state.GetActive() != "tab1" {
		t.Errorf("expected tab1 to be active, got %q", srv.state.GetActive())
	}

	// Activate tab2
	req := httptest.NewRequest("POST", "/api/tabs/tab2/activate", nil)
	req.SetPathValue("id", "tab2")
	w := httptest.NewRecorder()

	srv.handleActivateTab(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	// Verify tab2 is now active
	if srv.state.GetActive() != "tab2" {
		t.Errorf("expected tab2 to be active, got %q", srv.state.GetActive())
	}
}

// TestListTabs_WithTabs tests listing tabs with multiple tabs.
func TestListTabs_WithTabs(t *testing.T) {
	srv := setupTestServer()

	// Create some tabs
	srv.state.CreateTab(&Tab{ID: "list1", Title: "First", Type: TabTypeMarkdown})
	srv.state.CreateTab(&Tab{ID: "list2", Title: "Second", Type: TabTypeCode, Language: "go"})
	srv.state.CreateTab(&Tab{ID: "list3", Title: "Third", Type: TabTypeDiff})

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

	if len(resp.Tabs) != 3 {
		t.Errorf("expected 3 tabs, got %d", len(resp.Tabs))
	}

	// Check tab summaries
	tabsByID := make(map[string]*TabSummary)
	for _, tab := range resp.Tabs {
		tabsByID[tab.ID] = tab
	}

	if tab, ok := tabsByID["list1"]; !ok {
		t.Error("missing tab 'list1'")
	} else {
		if tab.Title != "First" {
			t.Errorf("expected title 'First', got %q", tab.Title)
		}
		if tab.Type != "markdown" {
			t.Errorf("expected type 'markdown', got %q", tab.Type)
		}
		if !tab.Active {
			t.Error("expected first tab to be active")
		}
	}

	if tab, ok := tabsByID["list2"]; !ok {
		t.Error("missing tab 'list2'")
	} else {
		if tab.Type != "code" {
			t.Errorf("expected type 'code', got %q", tab.Type)
		}
		if tab.Active {
			t.Error("expected second tab to not be active")
		}
	}

	if tab, ok := tabsByID["list3"]; !ok {
		t.Error("missing tab 'list3'")
	} else {
		if tab.Type != "diff" {
			t.Errorf("expected type 'diff', got %q", tab.Type)
		}
	}
}

// TestStatus_Empty tests status with no tabs.
func TestStatus_Empty(t *testing.T) {
	srv := setupTestServer()

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
	if resp.Tabs != 0 {
		t.Errorf("expected 0 tabs, got %d", resp.Tabs)
	}
}

// TestWriteJSON tests the JSON response helper.
func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()

	data := map[string]string{"key": "value"}
	writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("expected key='value', got %q", result["key"])
	}
}

// TestWriteError tests the error response helper.
func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	writeError(w, http.StatusBadRequest, "something went wrong")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Error != "something went wrong" {
		t.Errorf("expected error 'something went wrong', got %q", resp.Error)
	}
}

// TestValidTabTypes tests the valid tab types constant.
func TestValidTabTypes(t *testing.T) {
	tests := []struct {
		tabType string
		valid   bool
	}{
		{"", true},         // empty = auto-detect
		{"markdown", true},
		{"code", true},
		{"diff", true},
		{"invalid", false},
		{"html", false},
		{"text", false},
	}

	for _, tt := range tests {
		t.Run(tt.tabType, func(t *testing.T) {
			if ValidTabTypes[tt.tabType] != tt.valid {
				t.Errorf("ValidTabTypes[%q] = %v, want %v", tt.tabType, ValidTabTypes[tt.tabType], tt.valid)
			}
		})
	}
}
