// Package main is the entry point for agentviewer CLI.
package main

import (
	"flag"
	"fmt"
	"os"
)

const helpText = `agentviewer - Display rich content for AI agents in a browser

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
  DELETE /api/tabs              Clear all tabs
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
`

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help" {
		fmt.Print(helpText)
		os.Exit(0)
	}

	switch os.Args[1] {
	case "serve":
		runServe(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		fmt.Fprintf(os.Stderr, "Run 'agentviewer --help' for usage.\n")
		os.Exit(1)
	}
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 3333, "HTTP server port")
	fs.IntVar(port, "p", 3333, "HTTP server port (shorthand)")
	openBrowser := fs.Bool("open", false, "Open browser on start")
	fs.BoolVar(openBrowser, "o", false, "Open browser on start (shorthand)")
	contentType := fs.String("type", "", "Content type (markdown, code, diff)")
	fs.StringVar(contentType, "t", "", "Content type (shorthand)")
	title := fs.String("title", "", "Tab title")

	fs.Parse(args)

	// Get optional file argument
	file := ""
	if fs.NArg() > 0 {
		file = fs.Arg(0)
	}

	// Create server
	srv := NewServer()

	// If a file is provided, create initial tab
	if file != "" {
		content, err := ReadFileContent(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}

		tabType := TabType(*contentType)
		if tabType == "" {
			tabType = DetectContentType(file, content)
		}

		tabTitle := *title
		if tabTitle == "" {
			tabTitle = file
		}

		tab := &Tab{
			ID:       "initial",
			Title:    tabTitle,
			Type:     tabType,
			Content:  content,
			Language: DetectLanguage(file, content),
		}
		srv.state.CreateTab(tab)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	url := fmt.Sprintf("http://%s", addr)

	fmt.Printf("agentviewer server starting on %s\n", url)

	if *openBrowser {
		fmt.Println("Opening browser...")
		if err := OpenBrowser(url); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not open browser: %v\n", err)
		}
	}

	fmt.Println("Press Ctrl+C to stop.")

	if err := srv.ListenAndServe(addr); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
