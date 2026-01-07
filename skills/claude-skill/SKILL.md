---
name: agentviewer
description: Display rich content (markdown, code, diffs, mermaid diagrams) to users in a browser viewer. Use when you need to show complex content that benefits from rich rendering - reports, documentation, code with syntax highlighting, git diffs, or diagrams.
---

# Agentviewer

Display rich content in a browser-based tabbed viewer. Perfect for showing users:
- Markdown documents with rendered formatting
- Code with syntax highlighting
- Git diffs with side-by-side comparison
- Mermaid diagrams and LaTeX math

## Quick Start

Start the server (runs in background, opens browser):
```bash
agentviewer serve --open &
```

Create a tab with content:
```bash
curl -X POST localhost:3333/api/tabs \
  -d '{"title": "Report", "type": "markdown", "content": "# Hello\n\nContent here..."}'
```

## API Reference

Base URL: `http://localhost:3333`

### Create/Update Tab
```bash
# Markdown
curl -X POST localhost:3333/api/tabs \
  -d '{"title": "Notes", "type": "markdown", "content": "# My Notes\n\n- Item 1\n- Item 2"}'

# Code with syntax highlighting
curl -X POST localhost:3333/api/tabs \
  -d '{"title": "main.go", "type": "code", "content": "package main\n\nfunc main() {}", "language": "go"}'

# From file path (auto-detects type, enables live reload)
curl -X POST localhost:3333/api/tabs \
  -d '{"title": "Config", "file": "/path/to/config.yaml"}'

# Git diff
curl -X POST localhost:3333/api/tabs \
  -d '{"type": "diff", "diff": {"left": "old.go", "right": "new.go"}}'
```

### Other Endpoints
```bash
# List tabs
curl localhost:3333/api/tabs

# Get tab content
curl localhost:3333/api/tabs/{id}

# Delete tab
curl -X DELETE localhost:3333/api/tabs/{id}

# Delete all tabs
curl -X DELETE localhost:3333/api/tabs

# Activate (switch to) tab
curl -X POST localhost:3333/api/tabs/{id}/activate
```

## Content Types

| Type | Description | Features |
|------|-------------|----------|
| markdown | Rendered markdown | GFM, Mermaid diagrams, LaTeX math |
| code | Source code | Syntax highlighting for 180+ languages |
| diff | File comparison | Side-by-side view with line numbers |

## Best Practices

1. **Start server once** at the beginning of a session
2. **Use file paths** when showing existing files (enables auto-reload when file changes)
3. **Use meaningful titles** - they appear in the tab bar
4. **Clean up** - delete tabs when no longer needed
5. **Use markdown** for complex documents - it renders beautifully with diagrams and math

## Example: Showing Analysis Results

```bash
# Start server
agentviewer serve --open &

# Show a markdown report
curl -X POST localhost:3333/api/tabs -d '{
  "title": "Analysis",
  "type": "markdown",
  "content": "# Code Analysis\n\n## Summary\n- 10 files analyzed\n- 2 issues found\n\n## Issues\n\n### Issue 1\n```go\n// problematic code\nfunc foo() {}\n```\n"
}'

# Show a code file
curl -X POST localhost:3333/api/tabs -d '{
  "title": "main.go",
  "file": "./main.go"
}'
```

## Installation

```bash
# macOS/Linux via Homebrew (when available)
brew install agentviewer

# Or via Go
go install github.com/your-org/agentviewer@latest

# Or download binary from releases
```
