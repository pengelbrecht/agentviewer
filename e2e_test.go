//go:build e2e

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

// E2E tests for the agentviewer API.
// These tests start a real server and make actual HTTP requests.
// Run with: go test -tags=e2e -v ./...

// testServer holds a running test server instance.
type testServer struct {
	server   *Server
	listener net.Listener
	baseURL  string
	wsURL    string
	t        *testing.T
}

// newTestServer creates and starts a real server on a random port.
func newTestServer(t *testing.T) *testServer {
	t.Helper()

	srv := NewServer()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	wsURL := fmt.Sprintf("ws://127.0.0.1:%d/ws", port)

	// Start server in background
	go func() {
		srv.Serve(listener)
	}()

	ts := &testServer{
		server:   srv,
		listener: listener,
		baseURL:  baseURL,
		wsURL:    wsURL,
		t:        t,
	}

	// Wait for server to be ready
	ts.waitForReady()

	return ts
}

// waitForReady polls the server until it responds.
func (ts *testServer) waitForReady() {
	ts.t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(ts.baseURL + "/api/status")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	ts.t.Fatal("server did not become ready in time")
}

// close shuts down the test server.
func (ts *testServer) close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ts.server.Shutdown(ctx)
}

// httpRequest makes an HTTP request and returns the response.
func (ts *testServer) httpRequest(method, path string, body interface{}) (*http.Response, []byte) {
	ts.t.Helper()

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			ts.t.Fatalf("failed to marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, ts.baseURL+path, bodyReader)
	if err != nil {
		ts.t.Fatalf("failed to create request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ts.t.Fatalf("request failed: %v", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		ts.t.Fatalf("failed to read response body: %v", err)
	}

	return resp, respBody
}

// === E2E Tests ===

// TestE2E_ServerStatus tests the /api/status endpoint.
func TestE2E_ServerStatus(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	resp, body := ts.httpRequest("GET", "/api/status", nil)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var status StatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if status.Version != Version {
		t.Errorf("expected version %q, got %q", Version, status.Version)
	}
	if status.Tabs != 0 {
		t.Errorf("expected 0 tabs, got %d", status.Tabs)
	}
	if status.Uptime < 0 {
		t.Errorf("expected non-negative uptime, got %d", status.Uptime)
	}
}

// TestE2E_CreateAndGetTab tests creating and retrieving a tab.
func TestE2E_CreateAndGetTab(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Create a markdown tab
	createReq := CreateTabRequest{
		ID:      "e2e-test-1",
		Title:   "E2E Test Tab",
		Type:    "markdown",
		Content: "# Hello E2E\n\nThis is a test.",
	}

	resp, body := ts.httpRequest("POST", "/api/tabs", createReq)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create failed with status %d: %s", resp.StatusCode, string(body))
	}

	var createResp CreateTabResponse
	if err := json.Unmarshal(body, &createResp); err != nil {
		t.Fatalf("failed to parse create response: %v", err)
	}

	if createResp.ID != "e2e-test-1" {
		t.Errorf("expected ID 'e2e-test-1', got %q", createResp.ID)
	}
	if !createResp.Created {
		t.Error("expected Created to be true")
	}

	// Get the tab
	resp, body = ts.httpRequest("GET", "/api/tabs/e2e-test-1", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tab Tab
	if err := json.Unmarshal(body, &tab); err != nil {
		t.Fatalf("failed to parse tab: %v", err)
	}

	if tab.ID != "e2e-test-1" {
		t.Errorf("expected ID 'e2e-test-1', got %q", tab.ID)
	}
	if tab.Title != "E2E Test Tab" {
		t.Errorf("expected title 'E2E Test Tab', got %q", tab.Title)
	}
	if tab.Type != TabTypeMarkdown {
		t.Errorf("expected type 'markdown', got %q", tab.Type)
	}
	if !strings.Contains(tab.Content, "# Hello E2E") {
		t.Errorf("expected content to contain '# Hello E2E', got %q", tab.Content)
	}
}

// TestE2E_ListTabs tests listing all tabs.
func TestE2E_ListTabs(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Create multiple tabs
	tabs := []CreateTabRequest{
		{ID: "list-1", Title: "Tab 1", Type: "markdown", Content: "# One"},
		{ID: "list-2", Title: "Tab 2", Type: "code", Content: "fmt.Println()", Language: "go"},
		{ID: "list-3", Title: "Tab 3", Type: "markdown", Content: "# Three"},
	}

	for _, req := range tabs {
		resp, body := ts.httpRequest("POST", "/api/tabs", req)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("failed to create tab %s: %d - %s", req.ID, resp.StatusCode, string(body))
		}
	}

	// List all tabs
	resp, body := ts.httpRequest("GET", "/api/tabs", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list failed with status %d: %s", resp.StatusCode, string(body))
	}

	var listResp ListTabsResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}

	if len(listResp.Tabs) != 3 {
		t.Errorf("expected 3 tabs, got %d", len(listResp.Tabs))
	}

	// Verify tabs are present
	tabIDs := make(map[string]bool)
	for _, tab := range listResp.Tabs {
		tabIDs[tab.ID] = true
	}

	for _, expected := range []string{"list-1", "list-2", "list-3"} {
		if !tabIDs[expected] {
			t.Errorf("missing tab %q in list", expected)
		}
	}

	// First tab should be active
	for _, tab := range listResp.Tabs {
		if tab.ID == "list-1" && !tab.Active {
			t.Error("expected first tab to be active")
		}
	}
}

// TestE2E_DeleteTab tests deleting a tab.
func TestE2E_DeleteTab(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Create a tab
	createReq := CreateTabRequest{
		ID:      "to-delete",
		Title:   "Delete Me",
		Type:    "markdown",
		Content: "# Will be deleted",
	}
	resp, _ := ts.httpRequest("POST", "/api/tabs", createReq)
	if resp.StatusCode != http.StatusOK {
		t.Fatal("failed to create tab")
	}

	// Verify it exists
	resp, _ = ts.httpRequest("GET", "/api/tabs/to-delete", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatal("tab should exist")
	}

	// Delete the tab
	resp, _ = ts.httpRequest("DELETE", "/api/tabs/to-delete", nil)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", resp.StatusCode)
	}

	// Verify it's gone
	resp, _ = ts.httpRequest("GET", "/api/tabs/to-delete", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404 after delete, got %d", resp.StatusCode)
	}
}

// TestE2E_ActivateTab tests activating a specific tab.
func TestE2E_ActivateTab(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Create two tabs
	ts.httpRequest("POST", "/api/tabs", CreateTabRequest{ID: "active-1", Title: "Tab 1", Type: "markdown", Content: "# One"})
	ts.httpRequest("POST", "/api/tabs", CreateTabRequest{ID: "active-2", Title: "Tab 2", Type: "markdown", Content: "# Two"})

	// First tab should be active initially
	resp, body := ts.httpRequest("GET", "/api/tabs", nil)
	var listResp ListTabsResponse
	json.Unmarshal(body, &listResp)

	for _, tab := range listResp.Tabs {
		if tab.ID == "active-1" && !tab.Active {
			t.Error("expected first tab to be active initially")
		}
	}

	// Activate second tab
	resp, _ = ts.httpRequest("POST", "/api/tabs/active-2/activate", nil)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", resp.StatusCode)
	}

	// Verify second tab is now active
	resp, body = ts.httpRequest("GET", "/api/tabs", nil)
	json.Unmarshal(body, &listResp)

	for _, tab := range listResp.Tabs {
		if tab.ID == "active-2" && !tab.Active {
			t.Error("expected second tab to be active after activation")
		}
		if tab.ID == "active-1" && tab.Active {
			t.Error("expected first tab to be inactive after activating second")
		}
	}
}

// TestE2E_ClearTabs tests clearing all tabs.
func TestE2E_ClearTabs(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Create several tabs
	for i := 1; i <= 5; i++ {
		ts.httpRequest("POST", "/api/tabs", CreateTabRequest{
			ID:      fmt.Sprintf("clear-%d", i),
			Title:   fmt.Sprintf("Tab %d", i),
			Type:    "markdown",
			Content: fmt.Sprintf("# Tab %d", i),
		})
	}

	// Verify tabs exist
	resp, body := ts.httpRequest("GET", "/api/status", nil)
	var status StatusResponse
	json.Unmarshal(body, &status)
	if status.Tabs != 5 {
		t.Fatalf("expected 5 tabs, got %d", status.Tabs)
	}

	// Clear all tabs
	resp, _ = ts.httpRequest("DELETE", "/api/tabs", nil)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", resp.StatusCode)
	}

	// Verify all tabs are gone
	resp, body = ts.httpRequest("GET", "/api/status", nil)
	json.Unmarshal(body, &status)
	if status.Tabs != 0 {
		t.Errorf("expected 0 tabs after clear, got %d", status.Tabs)
	}
}

// TestE2E_UpdateTab tests updating an existing tab.
func TestE2E_UpdateTab(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Create initial tab
	initialReq := CreateTabRequest{
		ID:      "update-test",
		Title:   "Original Title",
		Type:    "markdown",
		Content: "# Original Content",
	}
	resp, body := ts.httpRequest("POST", "/api/tabs", initialReq)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("failed to create initial tab: %d", resp.StatusCode)
	}

	var createResp CreateTabResponse
	json.Unmarshal(body, &createResp)
	if !createResp.Created {
		t.Error("expected Created to be true for new tab")
	}

	// Update the tab with same ID
	updateReq := CreateTabRequest{
		ID:      "update-test",
		Title:   "Updated Title",
		Type:    "markdown",
		Content: "# Updated Content",
	}
	resp, body = ts.httpRequest("POST", "/api/tabs", updateReq)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("failed to update tab: %d", resp.StatusCode)
	}

	json.Unmarshal(body, &createResp)
	if createResp.Created {
		t.Error("expected Created to be false for update")
	}

	// Verify the update
	resp, body = ts.httpRequest("GET", "/api/tabs/update-test", nil)
	var tab Tab
	json.Unmarshal(body, &tab)

	if tab.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %q", tab.Title)
	}
	if tab.Content != "# Updated Content" {
		t.Errorf("expected content '# Updated Content', got %q", tab.Content)
	}
}

// TestE2E_AutoDetectType tests automatic content type detection.
func TestE2E_AutoDetectType(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	tests := []struct {
		name         string
		content      string
		file         string
		expectedType string
	}{
		{
			name:         "markdown by content",
			content:      "# Hello World\n\nThis is **markdown**.",
			expectedType: "markdown",
		},
		{
			name:         "diff by content",
			content:      "--- a/old.txt\n+++ b/new.txt\n@@ -1 +1 @@\n-old\n+new",
			expectedType: "diff",
		},
		{
			name:         "code by file extension",
			content:      "package main\n\nfunc main() {}",
			file:         "main.go",
			expectedType: "code",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateTabRequest{
				ID:      fmt.Sprintf("auto-%d", i),
				Title:   tt.name,
				Content: tt.content,
				File:    tt.file,
				// Type is intentionally omitted for auto-detection
			}

			resp, body := ts.httpRequest("POST", "/api/tabs", req)
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("failed to create tab: %d - %s", resp.StatusCode, string(body))
			}

			var createResp CreateTabResponse
			json.Unmarshal(body, &createResp)

			if createResp.Type != tt.expectedType {
				t.Errorf("expected type %q, got %q", tt.expectedType, createResp.Type)
			}
		})
	}
}

// TestE2E_CreateFromFile tests creating a tab from a file.
func TestE2E_CreateFromFile(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Get path to test file
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	filePath := filepath.Join(cwd, "testdata", "sample.go")

	req := CreateTabRequest{
		ID:    "from-file",
		Title: "From File",
		Type:  "code",
		File:  filePath,
	}

	resp, body := ts.httpRequest("POST", "/api/tabs", req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("failed to create tab: %d - %s", resp.StatusCode, string(body))
	}

	// Verify content was read
	resp, body = ts.httpRequest("GET", "/api/tabs/from-file", nil)
	var tab Tab
	json.Unmarshal(body, &tab)

	if !strings.Contains(tab.Content, "package main") {
		t.Errorf("expected content to contain 'package main', got %q", tab.Content)
	}
}

// TestE2E_CreateDiff tests creating a diff tab.
func TestE2E_CreateDiff(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Get paths to test files
	cwd, _ := os.Getwd()
	leftPath := filepath.Join(cwd, "testdata", "version1.txt")
	rightPath := filepath.Join(cwd, "testdata", "version2.txt")

	req := CreateTabRequest{
		ID:    "diff-test",
		Title: "Diff Test",
		Type:  "diff",
		Diff: &DiffReq{
			Left:       leftPath,
			Right:      rightPath,
			LeftLabel:  "Version 1",
			RightLabel: "Version 2",
		},
	}

	resp, body := ts.httpRequest("POST", "/api/tabs", req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("failed to create diff tab: %d - %s", resp.StatusCode, string(body))
	}

	// Verify diff was computed
	resp, body = ts.httpRequest("GET", "/api/tabs/diff-test", nil)
	var tab Tab
	json.Unmarshal(body, &tab)

	if tab.Type != TabTypeDiff {
		t.Errorf("expected type 'diff', got %q", tab.Type)
	}
	if !strings.Contains(tab.Content, "-") || !strings.Contains(tab.Content, "+") {
		t.Errorf("expected diff content with -/+ lines, got %q", tab.Content)
	}
	if tab.DiffMeta == nil {
		t.Error("expected DiffMeta to be set")
	} else {
		if tab.DiffMeta.LeftLabel != "Version 1" {
			t.Errorf("expected LeftLabel 'Version 1', got %q", tab.DiffMeta.LeftLabel)
		}
		if tab.DiffMeta.RightLabel != "Version 2" {
			t.Errorf("expected RightLabel 'Version 2', got %q", tab.DiffMeta.RightLabel)
		}
	}
}

// TestE2E_CreateDiffUnified tests creating a diff from unified diff string.
func TestE2E_CreateDiffUnified(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	unifiedDiff := `--- a/old.txt
+++ b/new.txt
@@ -1,3 +1,3 @@
 line1
-old line
+new line
 line3`

	req := CreateTabRequest{
		ID:    "unified-diff",
		Title: "Unified Diff",
		Type:  "diff",
		Diff: &DiffReq{
			Unified: unifiedDiff,
		},
	}

	resp, body := ts.httpRequest("POST", "/api/tabs", req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("failed to create unified diff tab: %d - %s", resp.StatusCode, string(body))
	}

	resp, body = ts.httpRequest("GET", "/api/tabs/unified-diff", nil)
	var tab Tab
	json.Unmarshal(body, &tab)

	if !strings.Contains(tab.Content, "old line") {
		t.Errorf("expected content to contain 'old line', got %q", tab.Content)
	}
}

// TestE2E_ErrorCases tests various error cases.
func TestE2E_ErrorCases(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "get nonexistent tab",
			method:         "GET",
			path:           "/api/tabs/nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Tab not found",
		},
		{
			name:           "delete nonexistent tab",
			method:         "DELETE",
			path:           "/api/tabs/nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Tab not found",
		},
		{
			name:           "activate nonexistent tab",
			method:         "POST",
			path:           "/api/tabs/nonexistent/activate",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Tab not found",
		},
		{
			name:   "invalid tab type",
			method: "POST",
			path:   "/api/tabs",
			body: CreateTabRequest{
				Title:   "Invalid",
				Type:    "invalid",
				Content: "test",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid type",
		},
		{
			name:   "diff without data",
			method: "POST",
			path:   "/api/tabs",
			body: CreateTabRequest{
				Title: "Empty Diff",
				Type:  "diff",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Diff type requires",
		},
		{
			name:   "file not found",
			method: "POST",
			path:   "/api/tabs",
			body: CreateTabRequest{
				Title: "Missing File",
				Type:  "code",
				File:  "/nonexistent/file.go",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Cannot read file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := ts.httpRequest(tt.method, tt.path, tt.body)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedError != "" {
				var errResp ErrorResponse
				json.Unmarshal(body, &errResp)
				if !strings.Contains(errResp.Error, tt.expectedError) {
					t.Errorf("expected error to contain %q, got %q", tt.expectedError, errResp.Error)
				}
			}
		})
	}
}

// TestE2E_WebSocketBroadcast tests that WebSocket clients receive broadcasts.
func TestE2E_WebSocketBroadcast(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect WebSocket client
	conn, _, err := websocket.Dial(ctx, ts.wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect WebSocket: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Give time for connection to be established
	time.Sleep(100 * time.Millisecond)

	// Create a tab via HTTP (should trigger WebSocket broadcast)
	ts.httpRequest("POST", "/api/tabs", CreateTabRequest{
		ID:      "ws-test",
		Title:   "WS Test",
		Type:    "markdown",
		Content: "# WebSocket Test",
	})

	// Read WebSocket message
	readCtx, readCancel := context.WithTimeout(ctx, 3*time.Second)
	defer readCancel()

	_, data, err := conn.Read(readCtx)
	if err != nil {
		t.Fatalf("failed to read WebSocket message: %v", err)
	}

	var msg WSMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if msg.Type != "tab_created" {
		t.Errorf("expected type 'tab_created', got %q", msg.Type)
	}
	if msg.Tab == nil {
		t.Error("expected Tab to be set")
	} else if msg.Tab.ID != "ws-test" {
		t.Errorf("expected Tab.ID 'ws-test', got %q", msg.Tab.ID)
	}
}

// TestE2E_WebSocketClientMessages tests that WebSocket clients can send messages.
func TestE2E_WebSocketClientMessages(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Create tabs first
	ts.httpRequest("POST", "/api/tabs", CreateTabRequest{ID: "ws-tab-1", Title: "Tab 1", Type: "markdown", Content: "# One"})
	ts.httpRequest("POST", "/api/tabs", CreateTabRequest{ID: "ws-tab-2", Title: "Tab 2", Type: "markdown", Content: "# Two"})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect WebSocket client
	conn, _, err := websocket.Dial(ctx, ts.wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect WebSocket: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	time.Sleep(100 * time.Millisecond)

	// Send activate_tab message via WebSocket
	activateMsg := WSMessage{Type: "activate_tab", ID: "ws-tab-2"}
	msgData, _ := json.Marshal(activateMsg)
	if err := conn.Write(ctx, websocket.MessageText, msgData); err != nil {
		t.Fatalf("failed to write to WebSocket: %v", err)
	}

	// Read broadcast response
	readCtx, readCancel := context.WithTimeout(ctx, 3*time.Second)
	defer readCancel()

	_, data, err := conn.Read(readCtx)
	if err != nil {
		t.Fatalf("failed to read WebSocket response: %v", err)
	}

	var responseMsg WSMessage
	json.Unmarshal(data, &responseMsg)

	if responseMsg.Type != "tab_activated" {
		t.Errorf("expected type 'tab_activated', got %q", responseMsg.Type)
	}
	if responseMsg.ID != "ws-tab-2" {
		t.Errorf("expected ID 'ws-tab-2', got %q", responseMsg.ID)
	}

	// Verify via HTTP that tab was actually activated
	resp, body := ts.httpRequest("GET", "/api/tabs", nil)
	var listResp ListTabsResponse
	json.Unmarshal(body, &listResp)

	for _, tab := range listResp.Tabs {
		if tab.ID == "ws-tab-2" && !tab.Active {
			t.Error("expected ws-tab-2 to be active")
		}
	}
	_ = resp
}

// TestE2E_WebSocketCloseTab tests closing tabs via WebSocket.
func TestE2E_WebSocketCloseTab(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Create a tab
	ts.httpRequest("POST", "/api/tabs", CreateTabRequest{ID: "ws-close", Title: "To Close", Type: "markdown", Content: "# Close me"})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect WebSocket
	conn, _, err := websocket.Dial(ctx, ts.wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect WebSocket: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	time.Sleep(100 * time.Millisecond)

	// Send close_tab message
	closeMsg := WSMessage{Type: "close_tab", ID: "ws-close"}
	msgData, _ := json.Marshal(closeMsg)
	if err := conn.Write(ctx, websocket.MessageText, msgData); err != nil {
		t.Fatalf("failed to write to WebSocket: %v", err)
	}

	// Read broadcast response
	readCtx, readCancel := context.WithTimeout(ctx, 3*time.Second)
	defer readCancel()

	_, data, err := conn.Read(readCtx)
	if err != nil {
		t.Fatalf("failed to read WebSocket response: %v", err)
	}

	var responseMsg WSMessage
	json.Unmarshal(data, &responseMsg)

	if responseMsg.Type != "tab_deleted" {
		t.Errorf("expected type 'tab_deleted', got %q", responseMsg.Type)
	}

	// Verify tab is gone via HTTP
	resp, _ := ts.httpRequest("GET", "/api/tabs/ws-close", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404 after close, got %d", resp.StatusCode)
	}
}

// TestE2E_FullWorkflow tests a complete workflow: create, update, list, activate, delete.
func TestE2E_FullWorkflow(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// 1. Start with empty state
	_, body := ts.httpRequest("GET", "/api/status", nil)
	var status StatusResponse
	json.Unmarshal(body, &status)
	if status.Tabs != 0 {
		t.Fatal("expected 0 tabs initially")
	}

	// 2. Create several tabs of different types
	ts.httpRequest("POST", "/api/tabs", CreateTabRequest{
		ID: "workflow-md", Title: "Markdown Tab", Type: "markdown",
		Content: "# Hello\n\nThis is markdown.",
	})
	ts.httpRequest("POST", "/api/tabs", CreateTabRequest{
		ID: "workflow-code", Title: "Code Tab", Type: "code",
		Content: "func main() {}", Language: "go",
	})
	ts.httpRequest("POST", "/api/tabs", CreateTabRequest{
		ID: "workflow-diff", Title: "Diff Tab", Type: "diff",
		Content: "--- a\n+++ b\n@@ -1 +1 @@\n-old\n+new",
	})

	// 3. Verify tab count
	_, body = ts.httpRequest("GET", "/api/status", nil)
	json.Unmarshal(body, &status)
	if status.Tabs != 3 {
		t.Errorf("expected 3 tabs, got %d", status.Tabs)
	}

	// 4. List and verify all tabs
	_, body = ts.httpRequest("GET", "/api/tabs", nil)
	var listResp ListTabsResponse
	json.Unmarshal(body, &listResp)
	if len(listResp.Tabs) != 3 {
		t.Errorf("expected 3 tabs in list, got %d", len(listResp.Tabs))
	}

	// 5. Update a tab
	ts.httpRequest("POST", "/api/tabs", CreateTabRequest{
		ID: "workflow-md", Title: "Updated Markdown", Type: "markdown",
		Content: "# Updated\n\nThis content was updated.",
	})

	_, body = ts.httpRequest("GET", "/api/tabs/workflow-md", nil)
	var tab Tab
	json.Unmarshal(body, &tab)
	if tab.Title != "Updated Markdown" {
		t.Errorf("expected updated title, got %q", tab.Title)
	}

	// 6. Activate a different tab
	ts.httpRequest("POST", "/api/tabs/workflow-code/activate", nil)
	_, body = ts.httpRequest("GET", "/api/tabs", nil)
	json.Unmarshal(body, &listResp)
	for _, tab := range listResp.Tabs {
		if tab.ID == "workflow-code" && !tab.Active {
			t.Error("expected workflow-code to be active")
		}
	}

	// 7. Delete a tab
	ts.httpRequest("DELETE", "/api/tabs/workflow-diff", nil)
	_, body = ts.httpRequest("GET", "/api/status", nil)
	json.Unmarshal(body, &status)
	if status.Tabs != 2 {
		t.Errorf("expected 2 tabs after delete, got %d", status.Tabs)
	}

	// 8. Clear all tabs
	ts.httpRequest("DELETE", "/api/tabs", nil)
	_, body = ts.httpRequest("GET", "/api/status", nil)
	json.Unmarshal(body, &status)
	if status.Tabs != 0 {
		t.Errorf("expected 0 tabs after clear, got %d", status.Tabs)
	}
}

// TestE2E_ConcurrentRequests tests handling of concurrent requests.
func TestE2E_ConcurrentRequests(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Create many tabs concurrently
	numTabs := 20
	done := make(chan bool, numTabs)

	for i := 0; i < numTabs; i++ {
		go func(n int) {
			ts.httpRequest("POST", "/api/tabs", CreateTabRequest{
				ID:      fmt.Sprintf("concurrent-%d", n),
				Title:   fmt.Sprintf("Tab %d", n),
				Type:    "markdown",
				Content: fmt.Sprintf("# Tab %d", n),
			})
			done <- true
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < numTabs; i++ {
		<-done
	}

	// Verify all tabs were created
	_, body := ts.httpRequest("GET", "/api/status", nil)
	var status StatusResponse
	json.Unmarshal(body, &status)

	if status.Tabs != numTabs {
		t.Errorf("expected %d tabs, got %d", numTabs, status.Tabs)
	}
}

// TestE2E_StaticFiles tests serving static files.
func TestE2E_StaticFiles(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Request the root page (should serve index.html)
	resp, body := ts.httpRequest("GET", "/", nil)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("expected Content-Type to contain 'text/html', got %q", contentType)
	}

	// Verify it contains HTML
	if !strings.Contains(string(body), "<!DOCTYPE html>") && !strings.Contains(string(body), "<html") {
		t.Error("expected HTML content")
	}
}

// TestE2E_ContentTypeHeaders tests that responses have correct content types.
func TestE2E_ContentTypeHeaders(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// JSON endpoints should return application/json
	endpoints := []string{
		"/api/status",
		"/api/tabs",
	}

	for _, endpoint := range endpoints {
		resp, _ := ts.httpRequest("GET", endpoint, nil)
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			t.Errorf("endpoint %s: expected Content-Type 'application/json', got %q", endpoint, contentType)
		}
	}
}
