//go:build e2e

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// startTestServer starts a real server on a random available port and returns the base URL.
func startTestServer(t *testing.T) (string, func()) {
	t.Helper()

	// Find an available port by binding to :0
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get available port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	srv := NewServer()

	// Start server in background
	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			// Ignore closed errors during test cleanup
		}
	}()

	// Wait for server to be ready
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/api/status")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}

	return baseURL, cleanup
}

// createTestTab creates a tab via API and returns the tab ID.
func createTestTab(t *testing.T, baseURL string, title, tabType, content string) string {
	t.Helper()

	body := map[string]string{
		"title":   title,
		"type":    tabType,
		"content": content,
	}
	bodyBytes, _ := json.Marshal(body)

	resp, err := http.Post(baseURL+"/api/tabs", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("failed to create tab: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	id, ok := result["id"].(string)
	if !ok {
		t.Fatal("response missing id")
	}
	return id
}

// === Core Smoke Tests ===

// TestBrowser_PageLoads verifies the page loads correctly.
func TestBrowser_PageLoads(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var title string
	var appVisible bool

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible("#app", chromedp.ByID),
		chromedp.Title(&title),
		chromedp.Evaluate(`document.getElementById('app') !== null`, &appVisible),
	)

	if err != nil {
		t.Fatalf("browser test failed: %v", err)
	}

	if title != "agentviewer" {
		t.Errorf("expected title 'agentviewer', got %q", title)
	}

	if !appVisible {
		t.Error("expected app element to be visible")
	}
}

// TestBrowser_EmptyState verifies the empty state is shown when no tabs exist.
func TestBrowser_EmptyState(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var emptyText string
	var emptyStateVisible bool

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible(".empty-state", chromedp.ByQuery),
		chromedp.Evaluate(`document.querySelector('.empty-state') !== null`, &emptyStateVisible),
		chromedp.Text(".empty-state h2", &emptyText, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("browser test failed: %v", err)
	}

	if !emptyStateVisible {
		t.Error("expected empty state to be visible")
	}

	if emptyText != "No tabs open" {
		t.Errorf("expected 'No tabs open', got %q", emptyText)
	}
}

// TestBrowser_TabsRender verifies that tabs render correctly.
func TestBrowser_TabsRender(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	// Create tabs before opening the browser
	createTestTab(t, baseURL, "First Tab", "markdown", "# Hello World")
	createTestTab(t, baseURL, "Second Tab", "markdown", "# Second Content")

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var tabCount int
	var firstTabTitle string
	var hasActiveTab bool

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible("#tabs-container", chromedp.ByID),
		// Wait for tabs to load
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Evaluate(`document.querySelectorAll('.tab').length`, &tabCount),
		chromedp.Evaluate(`document.querySelector('.tab .tab-title')?.textContent || ''`, &firstTabTitle),
		chromedp.Evaluate(`document.querySelector('.tab.active') !== null`, &hasActiveTab),
	)

	if err != nil {
		t.Fatalf("browser test failed: %v", err)
	}

	if tabCount != 2 {
		t.Errorf("expected 2 tabs, got %d", tabCount)
	}

	if firstTabTitle == "" {
		t.Error("expected tab title to be rendered")
	}

	if !hasActiveTab {
		t.Error("expected one tab to be active")
	}
}

// TestBrowser_TabSwitching verifies clicking tabs switches content.
func TestBrowser_TabSwitching(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	createTestTab(t, baseURL, "Tab One", "markdown", "# Content One")
	createTestTab(t, baseURL, "Tab Two", "markdown", "# Content Two")

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var initialContent string
	var afterClickContent string

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible(".content-markdown", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		// Get initial content (should be tab 1)
		chromedp.Evaluate(`document.querySelector('.content-markdown h1')?.textContent || ''`, &initialContent),
		// Click on the last tab (Tab Two)
		chromedp.Click(`.tab:last-child`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		// Get content after click
		chromedp.Evaluate(`document.querySelector('.content-markdown h1')?.textContent || ''`, &afterClickContent),
	)

	if err != nil {
		t.Fatalf("browser test failed: %v", err)
	}

	if initialContent != "Content One" {
		t.Errorf("expected initial content 'Content One', got %q", initialContent)
	}

	if afterClickContent != "Content Two" {
		t.Errorf("expected content after click 'Content Two', got %q", afterClickContent)
	}
}

// TestBrowser_TabClose verifies closing tabs works.
func TestBrowser_TabClose(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	createTestTab(t, baseURL, "Tab to Close", "markdown", "# Close Me")
	createTestTab(t, baseURL, "Remaining Tab", "markdown", "# Stay")

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var initialTabCount int
	var afterCloseTabCount int

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible("#tabs-container", chromedp.ByID),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Evaluate(`document.querySelectorAll('.tab').length`, &initialTabCount),
		// Click close button on first tab
		chromedp.Click(`.tab:first-child .tab-close`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Evaluate(`document.querySelectorAll('.tab').length`, &afterCloseTabCount),
	)

	if err != nil {
		t.Fatalf("browser test failed: %v", err)
	}

	if initialTabCount != 2 {
		t.Errorf("expected 2 tabs initially, got %d", initialTabCount)
	}

	if afterCloseTabCount != 1 {
		t.Errorf("expected 1 tab after close, got %d", afterCloseTabCount)
	}
}

// TestBrowser_WebSocketUpdates verifies live updates via WebSocket.
func TestBrowser_WebSocketUpdates(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var initialTabCount int

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible("#app", chromedp.ByID),
		chromedp.Sleep(500*time.Millisecond),
		// Check initial tab count (should be 0)
		chromedp.Evaluate(`document.querySelectorAll('.tab').length`, &initialTabCount),
	)

	if err != nil {
		t.Fatalf("browser test failed: %v", err)
	}

	if initialTabCount != 0 {
		t.Errorf("expected 0 tabs initially, got %d", initialTabCount)
	}

	// Create a tab via the API while browser is open
	createTestTab(t, baseURL, "Live Created Tab", "markdown", "# Created Live")

	var afterCreateTabCount int

	// Check if the tab appears via WebSocket update
	err = chromedp.Run(ctx,
		chromedp.Sleep(1*time.Second), // Wait for WebSocket message
		chromedp.Evaluate(`document.querySelectorAll('.tab').length`, &afterCreateTabCount),
	)

	if err != nil {
		t.Fatalf("browser test failed: %v", err)
	}

	if afterCreateTabCount != 1 {
		t.Errorf("expected 1 tab after WebSocket create, got %d", afterCreateTabCount)
	}
}

// TestBrowser_ConnectionStatus verifies the connection status indicator works.
func TestBrowser_ConnectionStatus(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var statusExists bool
	var statusConnected bool

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible("#app", chromedp.ByID),
		// Give time for WebSocket to connect and status to update
		chromedp.Sleep(1*time.Second),
		chromedp.Evaluate(`document.getElementById('connection-status') !== null`, &statusExists),
		chromedp.Evaluate(`document.getElementById('connection-status')?.classList.contains('connected') || false`, &statusConnected),
	)

	if err != nil {
		t.Fatalf("browser test failed: %v", err)
	}

	if !statusExists {
		t.Error("expected connection status element to exist")
	}

	if !statusConnected {
		t.Error("expected connection status to show connected")
	}
}

// TestBrowser_ThemeToggle verifies the theme toggle works.
func TestBrowser_ThemeToggle(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var initialTheme string
	var afterToggleTheme string

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible("#theme-toggle", chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),
		// Get initial theme
		chromedp.Evaluate(`document.documentElement.getAttribute('data-theme') || ''`, &initialTheme),
		// Click theme toggle
		chromedp.Click("#theme-toggle", chromedp.ByID),
		chromedp.Sleep(200*time.Millisecond),
		// Get theme after toggle
		chromedp.Evaluate(`document.documentElement.getAttribute('data-theme') || ''`, &afterToggleTheme),
	)

	if err != nil {
		t.Fatalf("browser test failed: %v", err)
	}

	// After toggle, theme should be set
	if afterToggleTheme == "" {
		t.Error("expected theme to be set after toggle")
	}

	// Valid theme values
	if afterToggleTheme != "light" && afterToggleTheme != "dark" {
		t.Errorf("expected theme to be 'light' or 'dark', got %q", afterToggleTheme)
	}
}

// TestBrowser_CodeTabRendering verifies code tabs render with syntax highlighting.
func TestBrowser_CodeTabRendering(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	// Create code tab with explicit language
	body := map[string]string{
		"title":    "Code Tab",
		"type":     "code",
		"content":  "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
		"language": "go",
	}
	bodyBytes, _ := json.Marshal(body)
	resp, _ := http.Post(baseURL+"/api/tabs", "application/json", bytes.NewReader(bodyBytes))
	resp.Body.Close()

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var hasCodeContent bool
	var hasLineNumbers bool
	var languageLabel string

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible(".content-code", chromedp.ByQuery),
		chromedp.Evaluate(`document.querySelector('.content-code') !== null`, &hasCodeContent),
		chromedp.Evaluate(`document.querySelector('.code-line .line-number') !== null`, &hasLineNumbers),
		chromedp.Evaluate(`document.querySelector('.code-language')?.textContent || ''`, &languageLabel),
	)

	if err != nil {
		t.Fatalf("browser test failed: %v", err)
	}

	if !hasCodeContent {
		t.Error("expected code content container to exist")
	}

	if !hasLineNumbers {
		t.Error("expected line numbers to exist")
	}

	if languageLabel != "go" {
		t.Errorf("expected language 'go', got %q", languageLabel)
	}
}

// TestBrowser_DiffTabRendering verifies diff tabs render correctly.
func TestBrowser_DiffTabRendering(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	diff := "--- a/old.txt\n+++ b/new.txt\n@@ -1,3 +1,3 @@\n line1\n-old line\n+new line\n line3"
	createTestTab(t, baseURL, "Diff Tab", "diff", diff)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var hasDiffContent bool
	var hasDiffTable bool
	var diffText string
	var hasObjectObject bool

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible(".content-diff", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond), // Wait for diff rendering
		chromedp.Evaluate(`document.querySelector('.content-diff') !== null`, &hasDiffContent),
		// Check for either diff2html table or fallback table
		chromedp.Evaluate(`document.querySelector('.d2h-diff-table') !== null || document.querySelector('.diff-table') !== null`, &hasDiffTable),
		// Get the text content to verify actual diff is rendered
		chromedp.Evaluate(`document.querySelector('.content-diff').textContent`, &diffText),
		// Check for [object Object] rendering bug
		chromedp.Evaluate(`document.querySelector('.content-diff').textContent.includes('[object Object]')`, &hasObjectObject),
	)

	if err != nil {
		t.Fatalf("browser test failed: %v", err)
	}

	if !hasDiffContent {
		t.Error("expected diff content container to exist")
	}

	if !hasDiffTable {
		t.Error("expected diff table to exist")
	}

	if hasObjectObject {
		t.Errorf("diff rendering bug: found '[object Object]' in content: %s", diffText[:min(200, len(diffText))])
	}

	// Verify actual diff content is visible (at least one of the diff lines)
	if !strings.Contains(diffText, "old line") && !strings.Contains(diffText, "new line") {
		t.Errorf("expected diff content to contain actual diff lines, got: %s", diffText[:min(200, len(diffText))])
	}
}

// TestBrowser_MultipleTabTypes verifies different tab types can coexist.
func TestBrowser_MultipleTabTypes(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	// Create tabs of different types
	createTestTab(t, baseURL, "Markdown", "markdown", "# Markdown Content")

	// Create code tab with language
	body := map[string]string{
		"title":    "Code",
		"type":     "code",
		"content":  "fmt.Println(\"test\")",
		"language": "go",
	}
	bodyBytes, _ := json.Marshal(body)
	resp, _ := http.Post(baseURL+"/api/tabs", "application/json", bytes.NewReader(bodyBytes))
	resp.Body.Close()

	createTestTab(t, baseURL, "Diff", "diff", "--- a\n+++ b\n@@ -1 +1 @@\n-old\n+new")

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var tabCount int
	var mdContent bool
	var codeContent bool
	var diffContent bool

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible("#tabs-container", chromedp.ByID),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Evaluate(`document.querySelectorAll('.tab').length`, &tabCount),

		// Click markdown tab and verify content
		chromedp.Click(`.tab:nth-child(1)`, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Evaluate(`document.querySelector('.content-markdown') !== null`, &mdContent),

		// Click code tab and verify content
		chromedp.Click(`.tab:nth-child(2)`, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Evaluate(`document.querySelector('.content-code') !== null`, &codeContent),

		// Click diff tab and verify content
		chromedp.Click(`.tab:nth-child(3)`, chromedp.ByQuery),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Evaluate(`document.querySelector('.content-diff') !== null`, &diffContent),
	)

	if err != nil {
		t.Fatalf("browser test failed: %v", err)
	}

	if tabCount != 3 {
		t.Errorf("expected 3 tabs, got %d", tabCount)
	}

	if !mdContent {
		t.Error("expected markdown content to render")
	}

	if !codeContent {
		t.Error("expected code content to render")
	}

	if !diffContent {
		t.Error("expected diff content to render")
	}
}

// === Mermaid and Feature Tests ===

// TestBrowserMermaidRendering verifies that mermaid diagrams are rendered correctly.
func TestBrowserMermaidRendering(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	// Create a markdown tab with a mermaid diagram
	mermaidContent := `# Mermaid Test

Here is a flowchart:

` + "```mermaid" + `
graph TD
    A[Start] --> B{Is it working?}
    B -->|Yes| C[Great!]
    B -->|No| D[Debug]
    D --> B
` + "```" + `

And some text after.
`

	createTestTab(t, baseURL, "Mermaid Test", "markdown", mermaidContent)

	// Create chromedp context with a timeout
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var hasMermaidDiv bool
	var hasSvg bool
	var svgContent string

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		// Wait for the content to load
		chromedp.WaitVisible(`.content-markdown`, chromedp.ByQuery),
		// Wait a moment for mermaid to render
		chromedp.Sleep(2*time.Second),
		// Check if mermaid div exists
		chromedp.Evaluate(`document.querySelector('.mermaid') !== null`, &hasMermaidDiv),
		// Check if SVG was rendered inside the mermaid div
		chromedp.Evaluate(`document.querySelector('.mermaid svg') !== null`, &hasSvg),
		// Get SVG content if it exists
		chromedp.Evaluate(`(document.querySelector('.mermaid svg')?.outerHTML || '').substring(0, 200)`, &svgContent),
	)

	if err != nil {
		t.Fatalf("chromedp failed: %v", err)
	}

	if !hasMermaidDiv {
		t.Error("mermaid div not found - code block not detected")
	}

	if !hasSvg {
		t.Error("mermaid SVG not rendered - mermaid.render() may have failed")
	}

	if svgContent == "" {
		t.Error("SVG content is empty")
	} else {
		t.Logf("Mermaid SVG rendered successfully (first 200 chars): %s", svgContent)
	}
}

// TestBrowserMermaidError verifies that mermaid errors are handled gracefully.
func TestBrowserMermaidError(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	// Create a markdown tab with invalid mermaid syntax
	invalidMermaidContent := `# Invalid Mermaid Test

` + "```mermaid" + `
this is not valid mermaid syntax
{{{ invalid }}}
` + "```" + `
`

	createTestTab(t, baseURL, "Invalid Mermaid", "markdown", invalidMermaidContent)

	// Create chromedp context
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var hasMermaidDiv bool
	var hasErrorElement bool
	var mermaidInnerHTML string

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible(`.content-markdown`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		chromedp.Evaluate(`document.querySelector('.mermaid') !== null`, &hasMermaidDiv),
		chromedp.Evaluate(`document.querySelector('.mermaid .mermaid-error') !== null || document.querySelector('.mermaid')?.innerHTML.includes('error')`, &hasErrorElement),
		chromedp.Evaluate(`document.querySelector('.mermaid')?.innerHTML || ''`, &mermaidInnerHTML),
	)

	if err != nil {
		t.Fatalf("chromedp failed: %v", err)
	}

	if !hasMermaidDiv {
		t.Error("mermaid div not found")
	}

	// The error should either show an error message or the mermaid library's error UI
	// We just want to make sure it doesn't crash and shows something
	if mermaidInnerHTML == "" {
		t.Error("mermaid container is empty - expected error message or fallback")
	} else {
		t.Logf("Mermaid error handling: inner HTML length = %d", len(mermaidInnerHTML))
	}
}

// TestBrowserMarkdownWithCode verifies code block syntax highlighting.
func TestBrowserMarkdownWithCode(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	mdContent := `# Code Test

` + "```go" + `
package main

func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

` + "```javascript" + `
const greeting = "Hello";
console.log(greeting);
` + "```" + `

` + "```python" + `
def hello():
    print("Hello, World!")
` + "```" + `

` + "```" + `
Plain code without language
` + "```" + `
`

	createTestTab(t, baseURL, "Code Test", "markdown", mdContent)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var codeBlockCount int
	var hasHljsClass bool
	var hasSyntaxTokens bool

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible(`.content-markdown`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		// Count code blocks
		chromedp.Evaluate(`document.querySelectorAll('pre code').length`, &codeBlockCount),
		// Check for hljs class (syntax highlighting applied)
		chromedp.Evaluate(`document.querySelector('pre code.hljs') !== null`, &hasHljsClass),
		// Check for actual syntax tokens
		chromedp.Evaluate(`document.querySelector('.hljs-keyword, .hljs-string, .hljs-function, .hljs-built_in') !== null`, &hasSyntaxTokens),
	)

	if err != nil {
		t.Fatalf("chromedp failed: %v", err)
	}

	if codeBlockCount != 4 {
		t.Errorf("expected 4 code blocks, got %d", codeBlockCount)
	}

	if !hasHljsClass {
		t.Error("highlight.js class not applied to code blocks")
	}

	if !hasSyntaxTokens {
		t.Error("no syntax highlighting tokens found (hljs-keyword, hljs-string, etc.)")
	}
}

// TestBrowserMermaidSequenceDiagram tests a sequence diagram specifically.
func TestBrowserMermaidSequenceDiagram(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	sequenceContent := `# Sequence Diagram Test

` + "```mermaid" + `
sequenceDiagram
    participant A as Alice
    participant B as Bob
    A->>B: Hello Bob!
    B->>A: Hi Alice!
    A->>B: How are you?
    B->>A: Great, thanks!
` + "```" + `
`

	createTestTab(t, baseURL, "Sequence Test", "markdown", sequenceContent)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var hasSvg bool
	var svgHasText bool

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible(`.content-markdown`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		chromedp.Evaluate(`document.querySelector('.mermaid svg') !== null`, &hasSvg),
		// Check if SVG contains text elements (the diagram labels)
		chromedp.Evaluate(`document.querySelector('.mermaid svg text') !== null || document.querySelector('.mermaid svg')?.innerHTML.includes('Alice')`, &svgHasText),
	)

	if err != nil {
		t.Fatalf("chromedp failed: %v", err)
	}

	if !hasSvg {
		t.Error("sequence diagram SVG not rendered")
	}

	if !svgHasText {
		t.Error("sequence diagram does not contain expected text (Alice)")
	}
}

// TestBrowserMermaidTheme verifies mermaid respects theme setting.
func TestBrowserMermaidTheme(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	mermaidContent := `# Theme Test

` + "```mermaid" + `
graph LR
    A --> B
` + "```" + `
`

	createTestTab(t, baseURL, "Theme Test", "markdown", mermaidContent)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var hasSvg bool
	var initialHTML string
	var afterToggleHTML string

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible(`.content-markdown`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		chromedp.Evaluate(`document.querySelector('.mermaid svg') !== null`, &hasSvg),
		chromedp.Evaluate(`document.querySelector('.mermaid')?.innerHTML.substring(0, 500) || ''`, &initialHTML),
	)

	if err != nil {
		t.Fatalf("chromedp failed: %v", err)
	}

	if !hasSvg {
		t.Error("mermaid SVG not rendered")
		return
	}

	// Toggle theme and check again
	err = chromedp.Run(ctx,
		chromedp.Click(`#theme-toggle`, chromedp.ByID),
		chromedp.Sleep(2*time.Second),
		chromedp.Evaluate(`document.querySelector('.mermaid')?.innerHTML.substring(0, 500) || ''`, &afterToggleHTML),
	)

	if err != nil {
		t.Fatalf("chromedp theme toggle failed: %v", err)
	}

	// The diagram should still be rendered (may look different due to theme)
	if afterToggleHTML == "" {
		t.Error("mermaid diagram disappeared after theme toggle")
	}

	// Note: The diagram might not re-render on theme toggle since it's already rendered.
	// The main test is that the diagram survives the theme toggle without errors.
	if strings.Contains(afterToggleHTML, "<svg") || strings.Contains(initialHTML, "<svg") {
		t.Log("Mermaid diagram survived theme toggle")
	}
}

// TestBrowserKatexInlineMath verifies inline math ($...$) renders correctly.
func TestBrowserKatexInlineMath(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	mathContent := `# Inline Math Test

The quadratic formula is $x = \frac{-b \pm \sqrt{b^2-4ac}}{2a}$.

Einstein's famous equation: $E = mc^2$.

And here's a Greek letter: $\alpha + \beta = \gamma$.
`

	createTestTab(t, baseURL, "Inline Math", "markdown", mathContent)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var katexSpanCount int
	var hasKatexHTML bool
	var katexSampleHTML string

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible(`.content-markdown`, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		// Count katex-inline spans that have been rendered
		chromedp.Evaluate(`document.querySelectorAll('.katex-inline .katex').length`, &katexSpanCount),
		// Check if KaTeX rendered HTML (contains .katex class with actual content)
		chromedp.Evaluate(`document.querySelector('.katex-inline .katex-html') !== null`, &hasKatexHTML),
		// Get sample rendered HTML
		chromedp.Evaluate(`document.querySelector('.katex-inline')?.innerHTML.substring(0, 300) || ''`, &katexSampleHTML),
	)

	if err != nil {
		t.Fatalf("chromedp failed: %v", err)
	}

	if katexSpanCount < 3 {
		t.Errorf("expected at least 3 rendered inline math expressions, got %d", katexSpanCount)
	}

	if !hasKatexHTML {
		t.Error("KaTeX did not render math to HTML (no .katex-html element found)")
	}

	if katexSampleHTML == "" {
		t.Error("KaTeX rendered content is empty")
	} else {
		t.Logf("KaTeX inline math rendered (first 300 chars): %s", katexSampleHTML)
	}
}

// TestBrowserKatexDisplayMath verifies display math ($$...$$) renders correctly.
func TestBrowserKatexDisplayMath(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	mathContent := `# Display Math Test

Here is Euler's identity:

$$e^{i\pi} + 1 = 0$$

And the integral of a Gaussian:

$$\int_{-\infty}^{\infty} e^{-x^2} dx = \sqrt{\pi}$$

The Schrödinger equation:

$$i\hbar\frac{\partial}{\partial t}\Psi = \hat{H}\Psi$$
`

	createTestTab(t, baseURL, "Display Math", "markdown", mathContent)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var katexDisplayCount int
	var hasKatexHTML bool
	var katexSampleHTML string

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible(`.content-markdown`, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		// Count katex-display spans that have been rendered
		chromedp.Evaluate(`document.querySelectorAll('.katex-display .katex').length`, &katexDisplayCount),
		// Check if KaTeX rendered HTML
		chromedp.Evaluate(`document.querySelector('.katex-display .katex-html') !== null`, &hasKatexHTML),
		// Get sample rendered HTML
		chromedp.Evaluate(`document.querySelector('.katex-display')?.innerHTML.substring(0, 300) || ''`, &katexSampleHTML),
	)

	if err != nil {
		t.Fatalf("chromedp failed: %v", err)
	}

	if katexDisplayCount < 3 {
		t.Errorf("expected at least 3 rendered display math expressions, got %d", katexDisplayCount)
	}

	if !hasKatexHTML {
		t.Error("KaTeX did not render display math to HTML")
	}

	if katexSampleHTML == "" {
		t.Error("KaTeX display rendered content is empty")
	} else {
		t.Logf("KaTeX display math rendered (first 300 chars): %s", katexSampleHTML)
	}
}

// TestBrowserKatexMixedContent verifies math renders alongside other markdown.
func TestBrowserKatexMixedContent(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	mixedContent := `# Mixed Content Test

This document has **bold text**, *italic text*, and math like $a^2 + b^2 = c^2$.

## Code and Math

Here's some code:

` + "```python" + `
def quadratic(a, b, c):
    return (-b + sqrt(b**2 - 4*a*c)) / (2*a)
` + "```" + `

And the formula it implements: $$x = \frac{-b + \sqrt{b^2 - 4ac}}{2a}$$

## Lists with Math

- Item with math: $\sum_{i=1}^{n} i = \frac{n(n+1)}{2}$
- Another item: $\prod_{i=1}^{n} i = n!$

Regular text to end.
`

	createTestTab(t, baseURL, "Mixed Content", "markdown", mixedContent)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var hasCodeBlock bool
	var hasInlineMath bool
	var hasDisplayMath bool
	var hasBoldText bool

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible(`.content-markdown`, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		// Check code block rendered
		chromedp.Evaluate(`document.querySelector('pre code.hljs') !== null`, &hasCodeBlock),
		// Check inline math rendered
		chromedp.Evaluate(`document.querySelector('.katex-inline .katex') !== null`, &hasInlineMath),
		// Check display math rendered
		chromedp.Evaluate(`document.querySelector('.katex-display .katex') !== null`, &hasDisplayMath),
		// Check bold text rendered
		chromedp.Evaluate(`document.querySelector('strong') !== null`, &hasBoldText),
	)

	if err != nil {
		t.Fatalf("chromedp failed: %v", err)
	}

	if !hasCodeBlock {
		t.Error("code block not rendered in mixed content")
	}

	if !hasInlineMath {
		t.Error("inline math not rendered in mixed content")
	}

	if !hasDisplayMath {
		t.Error("display math not rendered in mixed content")
	}

	if !hasBoldText {
		t.Error("bold text not rendered in mixed content")
	}
}

// TestBrowserMarkdownTaskLists verifies GFM task list styling.
func TestBrowserMarkdownTaskLists(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	taskListContent := `# Task List Test

## Shopping List
- [x] Buy groceries
- [ ] Pick up dry cleaning
- [x] Call the plumber

## Project Tasks
- [ ] Design the architecture
  - [x] Draw diagrams
  - [ ] Write documentation
- [x] Implement core features
- [ ] Write tests
`

	createTestTab(t, baseURL, "Task Lists", "markdown", taskListContent)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var checkboxCount int
	var checkedCount int
	var hasTaskListClass bool

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible(`.content-markdown`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		// Count all checkboxes
		chromedp.Evaluate(`document.querySelectorAll('input[type="checkbox"]').length`, &checkboxCount),
		// Count checked checkboxes
		chromedp.Evaluate(`document.querySelectorAll('input[type="checkbox"]:checked').length`, &checkedCount),
		// Check for task list styling class
		chromedp.Evaluate(`document.querySelector('.contains-task-list') !== null || document.querySelector('.task-list-item') !== null || document.querySelector('li input[type="checkbox"]') !== null`, &hasTaskListClass),
	)

	if err != nil {
		t.Fatalf("chromedp failed: %v", err)
	}

	if checkboxCount < 7 {
		t.Errorf("expected at least 7 checkboxes, got %d", checkboxCount)
	}

	if checkedCount < 4 {
		t.Errorf("expected at least 4 checked checkboxes, got %d", checkedCount)
	}

	if !hasTaskListClass {
		t.Error("task list styling elements not found")
	}

	t.Logf("Task lists: %d checkboxes, %d checked", checkboxCount, checkedCount)
}

// TestBrowserMarkdownTables verifies table styling with striped rows.
func TestBrowserMarkdownTables(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	tableContent := `# Table Test

| Name | Age | City |
|------|-----|------|
| Alice | 28 | New York |
| Bob | 35 | London |
| Carol | 42 | Paris |
| Dave | 31 | Tokyo |
| Eve | 26 | Sydney |

Another table with alignment:

| Left | Center | Right |
|:-----|:------:|------:|
| L1   | C1     | R1    |
| L2   | C2     | R2    |
`

	createTestTab(t, baseURL, "Tables", "markdown", tableContent)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var tableCount int
	var hasTableHeaders bool
	var rowCount int

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible(`.content-markdown`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		// Count tables
		chromedp.Evaluate(`document.querySelectorAll('.content-markdown table').length`, &tableCount),
		// Check for table headers
		chromedp.Evaluate(`document.querySelector('.content-markdown th') !== null`, &hasTableHeaders),
		// Count rows
		chromedp.Evaluate(`document.querySelectorAll('.content-markdown tr').length`, &rowCount),
	)

	if err != nil {
		t.Fatalf("chromedp failed: %v", err)
	}

	if tableCount != 2 {
		t.Errorf("expected 2 tables, got %d", tableCount)
	}

	if !hasTableHeaders {
		t.Error("table headers not found")
	}

	// 2 header rows + 5 + 2 data rows = 9 total rows
	if rowCount < 8 {
		t.Errorf("expected at least 8 table rows, got %d", rowCount)
	}

	t.Logf("Tables: %d tables, %d rows", tableCount, rowCount)
}

// TestBrowserMarkdownBlockquotes verifies blockquote and nested blockquote styling.
func TestBrowserMarkdownBlockquotes(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	blockquoteContent := `# Blockquote Test

> This is a simple blockquote.
> It can span multiple lines.

> ## Blockquote with header
>
> This blockquote contains a header and multiple paragraphs.
>
> Second paragraph in the blockquote.

> Level 1 quote
>
> > Level 2 nested quote
> >
> > > Level 3 deeply nested quote

Normal text after blockquotes.
`

	createTestTab(t, baseURL, "Blockquotes", "markdown", blockquoteContent)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var blockquoteCount int
	var hasNestedBlockquote bool
	var hasBlockquoteBorder bool

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible(`.content-markdown`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		// Count blockquotes
		chromedp.Evaluate(`document.querySelectorAll('.content-markdown blockquote').length`, &blockquoteCount),
		// Check for nested blockquotes
		chromedp.Evaluate(`document.querySelector('.content-markdown blockquote blockquote') !== null`, &hasNestedBlockquote),
		// Check for border-left styling (via computed style)
		chromedp.Evaluate(`(() => {
			const bq = document.querySelector('.content-markdown blockquote');
			if (!bq) return false;
			const style = window.getComputedStyle(bq);
			return style.borderLeftWidth !== '0px' && style.borderLeftStyle !== 'none';
		})()`, &hasBlockquoteBorder),
	)

	if err != nil {
		t.Fatalf("chromedp failed: %v", err)
	}

	if blockquoteCount < 3 {
		t.Errorf("expected at least 3 blockquotes, got %d", blockquoteCount)
	}

	if !hasNestedBlockquote {
		t.Error("nested blockquote not found")
	}

	if !hasBlockquoteBorder {
		t.Error("blockquote border-left styling not applied")
	}

	t.Logf("Blockquotes: %d found, nested: %v, styled: %v", blockquoteCount, hasNestedBlockquote, hasBlockquoteBorder)
}

// TestBrowserKatexErrorHandling verifies invalid math is handled gracefully.
func TestBrowserKatexErrorHandling(t *testing.T) {
	baseURL, cleanup := startTestServer(t)
	defer cleanup()

	invalidMathContent := `# Error Handling Test

Here's valid math: $x^2$

Here's invalid math: $\invalidcommand{broken}$

And another valid one: $y = mx + b$

Invalid display math:

$$\begin{invalid}
not a real environment
\end{invalid}$$
`

	createTestTab(t, baseURL, "Error Test", "markdown", invalidMathContent)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var hasRenderedMath bool
	var katexSpanCount int
	var pageNotCrashed bool

	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible(`.content-markdown`, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		// Check that valid math still rendered
		chromedp.Evaluate(`document.querySelector('.katex-inline .katex') !== null`, &hasRenderedMath),
		// Count rendered spans (should have some, even with errors)
		chromedp.Evaluate(`document.querySelectorAll('.katex-inline, .katex-display').length`, &katexSpanCount),
		// Check page didn't crash (h1 still exists)
		chromedp.Evaluate(`document.querySelector('h1') !== null`, &pageNotCrashed),
	)

	if err != nil {
		t.Fatalf("chromedp failed: %v", err)
	}

	if !pageNotCrashed {
		t.Error("page appears to have crashed (no h1 found)")
	}

	if !hasRenderedMath {
		t.Error("valid math expressions did not render despite invalid ones present")
	}

	// We expect 4 math spans: 3 inline + 1 display (some may show errors, that's OK)
	if katexSpanCount < 4 {
		t.Errorf("expected at least 4 math spans (including error cases), got %d", katexSpanCount)
	}

	t.Logf("KaTeX error handling: page stable, %d math spans found", katexSpanCount)
}
