# agentviewer

A localhost web server that allows AI agents (Claude Code, etc.) to display rich content to users in a browser.

## Overview

**agentviewer** is a single Go binary that:
1. Starts an HTTP server on localhost:3333
2. Serves a web UI with tabbed views
3. Exposes a REST API for agents to push content
4. Uses WebSocket to live-update the browser

```
┌─────────────────┐    WebSocket     ┌─────────────────┐
│  Browser Tab    │◄────────────────►│   agentviewer   │
│  (localhost)    │                  │   HTTP Server   │
└─────────────────┘                  └────────▲────────┘
                                              │ REST API (curl)
                                     ┌────────┴────────┐
                                     │  Claude / Agent │
                                     └─────────────────┘
```

## Content Types (v1)

| Type | Description | Rendering |
|------|-------------|-----------|
| `markdown` | Markdown documents | GFM + Mermaid + LaTeX math |
| `code` | Source code files | Syntax highlighting (language auto-detected or specified) |
| `diff` | File comparisons | Side-by-side diff view with syntax highlighting |

## CLI Interface

### Start Server

```bash
# Start server (foreground, blocks)
agentviewer serve

# Start server and open browser
agentviewer serve --open

# Custom port
agentviewer serve --port 4000

# Start with initial content
agentviewer serve --open README.md
agentviewer serve --open src/main.go --type code
```

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `--port`, `-p` | 3333 | HTTP server port |
| `--open`, `-o` | false | Open browser on start |
| `--type`, `-t` | auto | Content type for initial file (`markdown`, `code`, `diff`) |
| `--title` | filename | Tab title for initial file |

### Help Output (LLM-Friendly)

The `--help` flag produces structured, parseable output designed for both humans and LLMs:

```
agentviewer - Display rich content for AI agents in a browser

USAGE:
  agentviewer serve [OPTIONS] [FILE]
  agentviewer --help

DESCRIPTION:
  Starts a localhost HTTP server that displays markdown, code, and diffs
  in a browser. AI agents update content via REST API (curl).

  Once running, update content with:
    curl -X POST http://localhost:3333/api/tabs \
      -H "Content-Type: application/json" \
      -d '{"title": "Name", "type": "markdown", "content": "# Hello"}'

OPTIONS:
  --port, -p <PORT>     HTTP server port (default: 3333)
  --open, -o            Open browser automatically on start
  --type, -t <TYPE>     Content type: markdown, code, diff (default: auto-detect)
  --title <TITLE>       Tab title (default: filename)
  --help, -h            Show this help message

CONTENT TYPES:
  markdown    Rendered with GFM, Mermaid diagrams, LaTeX math
  code        Syntax highlighted source code
  diff        Side-by-side file comparison

API ENDPOINTS:
  POST   /api/tabs              Create or update a tab
  GET    /api/tabs              List all tabs
  GET    /api/tabs/:id          Get tab content
  DELETE /api/tabs/:id          Delete a tab
  POST   /api/tabs/:id/activate Switch to a tab
  GET    /api/status            Server status

EXAMPLES:
  # Start server and open browser
  agentviewer serve --open

  # Start with a markdown file
  agentviewer serve --open README.md

  # Create markdown tab via API
  curl -X POST localhost:3333/api/tabs \
    -d '{"title": "Notes", "type": "markdown", "content": "# My Notes"}'

  # Create diff tab via API
  curl -X POST localhost:3333/api/tabs \
    -d '{"type": "diff", "diff": {"left": "old.go", "right": "new.go"}}'

  # Show code file via API
  curl -X POST localhost:3333/api/tabs \
    -d '{"title": "main.go", "type": "code", "file": "/path/to/main.go"}'
```

## REST API

Base URL: `http://localhost:3333/api`

All endpoints accept JSON. Content can be provided either inline (`content` field) or by file path (`file` field).

### Create/Update Tab

```
POST /api/tabs
```

**Request Body:**

```json
{
  "id": "main",
  "title": "README",
  "type": "markdown",
  "content": "# Hello World\n\nThis is **bold**."
}
```

Or with file path:

```json
{
  "id": "main",
  "title": "README",
  "type": "markdown",
  "file": "/path/to/README.md"
}
```

**Response:**

```json
{
  "id": "main",
  "title": "README",
  "type": "markdown",
  "created": false
}
```

If `id` exists, content is replaced. If `id` is omitted, a new unique ID is generated.

### Create Diff Tab

```
POST /api/tabs
```

**Request Body (two files):**

```json
{
  "id": "changes",
  "title": "My Changes",
  "type": "diff",
  "diff": {
    "left": "/path/to/original.ts",
    "right": "/path/to/modified.ts",
    "leftLabel": "Original",
    "rightLabel": "Modified"
  }
}
```

**Request Body (unified diff):**

```json
{
  "id": "changes",
  "title": "Git Diff",
  "type": "diff",
  "diff": {
    "unified": "--- a/file.ts\n+++ b/file.ts\n@@ -1,3 +1,4 @@...",
    "language": "typescript"
  }
}
```

### List Tabs

```
GET /api/tabs
```

**Response:**

```json
{
  "tabs": [
    {"id": "main", "title": "README", "type": "markdown", "active": true},
    {"id": "code-1", "title": "main.go", "type": "code", "active": false}
  ]
}
```

### Get Tab Content

```
GET /api/tabs/:id
```

**Response:**

```json
{
  "id": "main",
  "title": "README",
  "type": "markdown",
  "content": "# Hello World..."
}
```

### Delete Tab

```
DELETE /api/tabs/:id
```

### Set Active Tab

```
POST /api/tabs/:id/activate
```

Switches the browser view to this tab.

### Clear All Tabs

```
DELETE /api/tabs
```

### Server Status

```
GET /api/status
```

**Response:**

```json
{
  "version": "0.1.0",
  "tabs": 3,
  "uptime": 3600
}
```

## WebSocket Protocol

Endpoint: `ws://localhost:3333/ws`

The browser connects via WebSocket to receive live updates.

### Server → Client Messages

```json
{"type": "tab_created", "tab": {"id": "main", "title": "README", "type": "markdown"}}
{"type": "tab_updated", "tab": {"id": "main", "title": "README", "type": "markdown"}}
{"type": "tab_deleted", "id": "main"}
{"type": "tab_activated", "id": "main"}
{"type": "content_updated", "id": "main", "content": "..."}
{"type": "tabs_cleared"}
```

### Client → Server Messages

```json
{"type": "activate_tab", "id": "main"}
{"type": "close_tab", "id": "main"}
```

## Web UI

### Layout

```
┌──────────────────────────────────────────────────────────┐
│ [Tab 1] [Tab 2] [Tab 3]                              [×] │
├──────────────────────────────────────────────────────────┤
│                                                          │
│                                                          │
│                    Content Area                          │
│                                                          │
│                                                          │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

### Features

- **Tabs**: Click to switch, middle-click or × to close
- **Dark/Light mode**: Follows system preference, toggle in UI
- **Responsive**: Works on various screen sizes
- **Keyboard shortcuts**:
  - `Ctrl/Cmd + 1-9`: Switch to tab N
  - `Ctrl/Cmd + W`: Close current tab
  - `Ctrl/Cmd + Shift + T`: Reopen last closed tab (if content still in memory)

### Rendering

**Markdown:**
- GitHub Flavored Markdown (tables, task lists, strikethrough)
- Syntax-highlighted code blocks
- Mermaid diagram rendering
- LaTeX math (inline `$...$` and block `$$...$$`)
- Auto-linked URLs

**Code:**
- Syntax highlighting based on file extension or specified language
- Line numbers
- Copy button

**Diff:**
- Side-by-side view
- Syntax highlighting
- Line numbers
- Collapsible unchanged sections
- Navigation between changes

## Example Usage (Claude's Perspective)

### Display a markdown file

```bash
curl -X POST http://localhost:3333/api/tabs \
  -H "Content-Type: application/json" \
  -d '{"id": "readme", "title": "README", "type": "markdown", "file": "/project/README.md"}'
```

### Show code with syntax highlighting

```bash
curl -X POST http://localhost:3333/api/tabs \
  -H "Content-Type: application/json" \
  -d '{"id": "main", "title": "main.go", "type": "code", "file": "/project/main.go"}'
```

### Display inline content

```bash
curl -X POST http://localhost:3333/api/tabs \
  -H "Content-Type: application/json" \
  -d '{"title": "Analysis", "type": "markdown", "content": "# Analysis\n\nHere are my findings..."}'
```

### Show a diff between two files

```bash
curl -X POST http://localhost:3333/api/tabs \
  -H "Content-Type: application/json" \
  -d '{
    "id": "diff",
    "title": "Changes",
    "type": "diff",
    "diff": {
      "left": "/project/file.go.orig",
      "right": "/project/file.go",
      "leftLabel": "Before",
      "rightLabel": "After"
    }
  }'
```

### Switch to a specific tab

```bash
curl -X POST http://localhost:3333/api/tabs/readme/activate
```

## Technical Implementation

### Go Dependencies

- **HTTP/WebSocket**: `net/http` + `gorilla/websocket` or `nhooyr.io/websocket`
- **Routing**: `chi` or standard `http.ServeMux` (Go 1.22+)
- **Embedded assets**: `embed` package for HTML/CSS/JS

### Frontend Dependencies (embedded)

- **Markdown**: [marked](https://marked.js.org/) or [markdown-it](https://github.com/markdown-it/markdown-it)
- **Syntax highlighting**: [highlight.js](https://highlightjs.org/) or [Prism](https://prismjs.com/)
- **Mermaid**: [mermaid](https://mermaid.js.org/)
- **Math**: [KaTeX](https://katex.org/)
- **Diff**: [diff2html](https://diff2html.xyz/) or custom implementation

### Project Structure

```
agentviewer/
├── main.go              # Entry point, CLI parsing
├── server.go            # HTTP server, routing
├── websocket.go         # WebSocket hub and connections
├── tabs.go              # Tab state management
├── handlers.go          # REST API handlers
├── render.go            # Content type detection, file reading
├── web/
│   ├── index.html       # Main HTML template
│   ├── app.js           # Frontend application
│   ├── style.css        # Styles
│   └── vendor/          # Embedded JS libraries
├── go.mod
├── go.sum
└── SPEC.md
```

### State Management

All state is in-memory (ephemeral):

```go
type Tab struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Type      string    `json:"type"`      // "markdown", "code", "diff"
    Content   string    `json:"content"`
    Language  string    `json:"language,omitempty"`  // for code
    DiffMeta  *DiffMeta `json:"diff,omitempty"`      // for diff
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

type DiffMeta struct {
    LeftLabel  string `json:"leftLabel"`
    RightLabel string `json:"rightLabel"`
    Language   string `json:"language"`
}

type State struct {
    mu       sync.RWMutex
    tabs     map[string]*Tab
    order    []string  // tab order
    activeID string
}
```

## Security Considerations

- **Localhost only**: Bind to `127.0.0.1`, never `0.0.0.0`
- **File access**: Only read files, never write; consider optional path restrictions
- **No execution**: Never execute file contents
- **CORS**: Restrict to localhost origins

## Future Enhancements (post-v1)

- [ ] Image display (inline or full tab)
- [ ] HTML email rendering (sanitized)
- [ ] Chart/graph rendering (from data)
- [ ] PDF viewing
- [ ] Terminal output display (ANSI colors)
- [ ] Multiple windows (different ports or browser tabs)
- [ ] File watching mode (auto-reload on change)
- [ ] Export/save rendered content
- [ ] Search within content
- [ ] Configurable themes
