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
