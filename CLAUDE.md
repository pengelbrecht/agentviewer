# CLAUDE.md - agentviewer

## Project Overview

**agentviewer** is a Go CLI tool that runs a localhost web server for AI agents to display rich content (markdown, code, diffs) to users in a browser. See `SPEC.md` for full specification.

## Development Commands

```bash
# Run during development
go run . serve --open

# Build binary
go build -o agentviewer .

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run e2e tests
go test -tags=e2e ./...

# Format code
go fmt ./...

# Lint (if golangci-lint installed)
golangci-lint run
```

## Architecture

```
main.go          → CLI entry point (cobra or stdlib flag)
server.go        → HTTP server setup, routing
websocket.go     → WebSocket hub, broadcast to clients
tabs.go          → Tab state management (in-memory)
handlers.go      → REST API handlers
render.go        → File reading, content type detection
web/             → Embedded frontend assets (go:embed)
```

**Key patterns:**
- All state in-memory, protected by `sync.RWMutex`
- WebSocket hub pattern: one goroutine manages all connections
- Handlers are methods on a `Server` struct for testability
- Frontend assets embedded via `//go:embed web/*`

## Code Style

- Use standard Go formatting (`go fmt`)
- Error handling: return errors, don't panic (except truly unrecoverable)
- Use `context.Context` for cancellation in long-running operations
- Prefer table-driven tests
- Keep handlers thin; business logic in separate functions

## CLI Design (LLM-Friendly)

The `--help` output is designed for both humans and LLMs to parse. Key principles:

1. **Structured sections**: USAGE, DESCRIPTION, OPTIONS, EXAMPLES - easy to parse
2. **Inline API examples**: Show curl commands directly in help so LLMs can copy/adapt
3. **All endpoints listed**: LLM doesn't need to read docs elsewhere
4. **Concrete examples**: Real commands, not abstract descriptions

When modifying CLI help:
- Keep sections clearly delimited with uppercase headers
- Always include working curl examples for API operations
- List all content types and their rendering behavior
- Don't use complex formatting that's hard to parse (avoid tables in help text)

## Testing Strategy

### Unit Tests

Use Go's built-in testing with `httptest` for HTTP handlers.

**What to unit test:**
- Tab CRUD operations (`tabs.go`)
- Content type detection (`render.go`)
- Individual HTTP handlers (`handlers.go`)
- WebSocket message serialization

**Example handler test:**

```go
func TestCreateTab(t *testing.T) {
    srv := NewServer()

    body := `{"id": "test", "title": "Test", "type": "markdown", "content": "# Hello"}`
    req := httptest.NewRequest("POST", "/api/tabs", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    srv.handleCreateTab(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("expected 200, got %d", w.Code)
    }

    // Verify tab was created
    tab, exists := srv.state.GetTab("test")
    if !exists {
        t.Error("tab was not created")
    }
    if tab.Title != "Test" {
        t.Errorf("expected title 'Test', got %q", tab.Title)
    }
}
```

**WebSocket testing:**

```go
func TestWebSocketBroadcast(t *testing.T) {
    srv := NewServer()
    ts := httptest.NewServer(srv)
    defer ts.Close()

    // Connect WebSocket client
    wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
    conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
    if err != nil {
        t.Fatal(err)
    }
    defer conn.Close()

    // Create tab via API, verify WS receives message
    // ...
}
```

### E2E Tests (without Playwright)

We avoid Playwright due to weight/complexity. Instead, use a layered approach:

#### Layer 1: API E2E Tests (Go)

Full integration tests that start a real server and hit it with HTTP/WebSocket clients.

```go
//go:build e2e

func TestFullWorkflow(t *testing.T) {
    // Start server on random port
    srv := NewServer()
    listener, _ := net.Listen("tcp", "127.0.0.1:0")
    port := listener.Addr().(*net.TCPAddr).Port
    go srv.Serve(listener)
    defer srv.Shutdown(context.Background())

    base := fmt.Sprintf("http://127.0.0.1:%d", port)

    // Create markdown tab
    resp, _ := http.Post(base+"/api/tabs", "application/json",
        strings.NewReader(`{"title": "Test", "type": "markdown", "content": "# Hi"}`))
    // assert response...

    // Create diff tab
    // List tabs
    // Delete tab
    // etc.
}
```

#### Layer 2: CLI E2E Tests (shell scripts or Go)

Test the actual binary's CLI interface:

```go
//go:build e2e

func TestCLIServe(t *testing.T) {
    // Build binary first
    cmd := exec.Command("go", "build", "-o", "agentviewer_test", ".")
    if err := cmd.Run(); err != nil {
        t.Fatal(err)
    }
    defer os.Remove("agentviewer_test")

    // Start server
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cmd = exec.CommandContext(ctx, "./agentviewer_test", "serve", "--port", "13333")
    cmd.Start()
    defer cmd.Process.Kill()

    // Wait for server to be ready
    waitForServer(t, "http://127.0.0.1:13333/api/status")

    // Test via curl-like requests
    // ...
}
```

#### Layer 3: Browser Smoke Tests (chromedp - lightweight CDP)

For critical UI functionality, use `chromedp` (Chrome DevTools Protocol) - much lighter than Playwright:

```go
//go:build e2e

import "github.com/chromedp/chromedp"

func TestBrowserRendering(t *testing.T) {
    // Start server
    srv := startTestServer(t)
    defer srv.Shutdown(context.Background())

    // Create a tab with content
    createTab(t, srv.URL, Tab{
        ID: "test", Title: "Test", Type: "markdown",
        Content: "# Hello\n\nThis is a test.",
    })

    // Use chromedp for browser automation
    ctx, cancel := chromedp.NewContext(context.Background())
    defer cancel()

    var title string
    var contentVisible bool

    err := chromedp.Run(ctx,
        chromedp.Navigate(srv.URL),
        chromedp.WaitVisible(`[data-tab="test"]`),
        chromedp.Text(`h1`, &title),
        chromedp.Evaluate(`document.querySelector('.content h1')?.textContent === 'Hello'`, &contentVisible),
    )

    if err != nil {
        t.Fatal(err)
    }
    if title != "Hello" {
        t.Errorf("expected h1 'Hello', got %q", title)
    }
}
```

**When to use chromedp:**
- Verify markdown renders correctly
- Verify syntax highlighting works
- Verify diff view displays properly
- Verify tab switching works
- Screenshot comparison for visual regressions (optional)

#### Layer 4: Manual Testing with Claude Chrome Extension

For informal UI/UX testing during development:

1. Start agentviewer: `go run . serve --open`
2. Use Claude Chrome Extension to interact with the page
3. Ask Claude to verify UI elements, check rendering, test interactions

**Example prompts for Claude Chrome Extension:**
- "Take a screenshot of the current view"
- "Click on the second tab and verify it switches"
- "Check if the mermaid diagram rendered correctly"
- "Verify the diff view shows line numbers"

This is useful for:
- Exploratory testing
- Visual verification humans would do
- Testing complex interactions
- Accessibility checks

### Test File Organization

```
agentviewer/
├── tabs_test.go           # Unit tests for tab state
├── handlers_test.go       # Unit tests for HTTP handlers
├── websocket_test.go      # Unit tests for WS logic
├── render_test.go         # Unit tests for content detection
├── e2e_test.go            # E2E tests (build tag: e2e)
├── browser_test.go        # chromedp tests (build tag: e2e)
└── testdata/
    ├── sample.md
    ├── sample.go
    └── sample.diff
```

### CI Pipeline

```yaml
# .github/workflows/test.yml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Unit tests
        run: go test -v -race ./...

      - name: E2E tests (API)
        run: go test -v -tags=e2e -run 'TestAPI' ./...

      - name: E2E tests (Browser)
        run: |
          # chromedp uses headless Chrome
          go test -v -tags=e2e -run 'TestBrowser' ./...
```

### Test Coverage Goals

| Area | Target | Notes |
|------|--------|-------|
| Tab state management | 90%+ | Core logic, must be solid |
| REST API handlers | 80%+ | All endpoints covered |
| WebSocket broadcast | 70%+ | Happy path + error cases |
| Content detection | 80%+ | Various file types |
| Browser rendering | Key paths | Smoke tests only |

## Build & Release

```bash
# Local build
go build -o agentviewer .

# Cross-compile
GOOS=darwin GOARCH=arm64 go build -o agentviewer-darwin-arm64 .
GOOS=darwin GOARCH=amd64 go build -o agentviewer-darwin-amd64 .
GOOS=linux GOARCH=amd64 go build -o agentviewer-linux-amd64 .

# With version info
go build -ldflags "-X main.version=0.1.0" -o agentviewer .
```

## Dependencies

**Required:**
- Go 1.22+ (for enhanced ServeMux routing)
- `nhooyr.io/websocket` or `gorilla/websocket`

**Testing:**
- `github.com/chromedp/chromedp` (e2e browser tests)

**Frontend (embedded, no build step):**
- marked.js (markdown)
- highlight.js (syntax highlighting)
- mermaid (diagrams)
- katex (math)
- diff2html (diff rendering)

## Common Tasks

### Adding a new content type

1. Add type constant in `tabs.go`
2. Add detection logic in `render.go`
3. Add rendering in `web/app.js`
4. Add unit tests for detection
5. Add e2e test for rendering

### Debugging WebSocket issues

```go
// Enable verbose logging
srv := NewServer(WithDebugLogging(true))
```

### Testing file reading

Create test files in `testdata/` and reference them in tests:

```go
func TestReadMarkdownFile(t *testing.T) {
    content, err := readFile("testdata/sample.md")
    // ...
}
```
