// Package main is the entry point for agentviewer CLI.
package main

import (
	"fmt"
	"os"

	// Import websocket package to ensure dependency is tracked in go.mod.
	// This will be used in websocket.go for WebSocket connections.
	_ "nhooyr.io/websocket"
)

func main() {
	fmt.Println("agentviewer - localhost web server for AI agents")
	fmt.Println("Usage: agentviewer serve [--port PORT] [--open]")
	os.Exit(0)
}
