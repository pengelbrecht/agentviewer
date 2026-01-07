//go:build e2e

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// CLI E2E tests for agentviewer.
// These tests build and run the actual binary to test CLI behavior.
// Run with: go test -tags=e2e -v -run TestCLI ./...

const testBinaryName = "agentviewer_test_binary"

// buildBinary compiles the test binary and returns a cleanup function.
func buildBinary(t *testing.T) (binaryPath string, cleanup func()) {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	binaryPath = filepath.Join(cwd, testBinaryName)

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v\nstderr: %s", err, stderr.String())
	}

	cleanup = func() {
		os.Remove(binaryPath)
	}

	return binaryPath, cleanup
}

// runCLI executes the CLI binary with the given arguments.
func runCLI(t *testing.T, binaryPath string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return stdout, stderr, exitCode
}

// waitForServer polls the server URL until it responds or timeout.
func waitForServer(t *testing.T, url string, timeout time.Duration) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url + "/api/status")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

// === CLI Help Tests ===

// TestCLI_Help tests the help output.
func TestCLI_Help(t *testing.T) {
	binaryPath, cleanup := buildBinary(t)
	defer cleanup()

	tests := []struct {
		name       string
		args       []string
		wantInOut  []string
		wantExit   int
	}{
		{
			name: "no args shows help",
			args: []string{},
			wantInOut: []string{
				"USAGE:",
				"agentviewer serve",
				"OPTIONS:",
				"--port",
				"--open",
				"API ENDPOINTS:",
				"EXAMPLES:",
			},
			wantExit: 0,
		},
		{
			name:       "help flag",
			args:       []string{"--help"},
			wantInOut:  []string{"USAGE:", "agentviewer serve"},
			wantExit:   0,
		},
		{
			name:       "short help flag",
			args:       []string{"-h"},
			wantInOut:  []string{"USAGE:", "agentviewer serve"},
			wantExit:   0,
		},
		{
			name:       "help command",
			args:       []string{"help"},
			wantInOut:  []string{"USAGE:", "agentviewer serve"},
			wantExit:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _, exitCode := runCLI(t, binaryPath, tt.args...)

			if exitCode != tt.wantExit {
				t.Errorf("expected exit code %d, got %d", tt.wantExit, exitCode)
			}

			for _, want := range tt.wantInOut {
				if !strings.Contains(stdout, want) {
					t.Errorf("expected output to contain %q, got:\n%s", want, stdout)
				}
			}
		})
	}
}

// TestCLI_HelpContainsAPIExamples verifies help includes curl examples for LLMs.
func TestCLI_HelpContainsAPIExamples(t *testing.T) {
	binaryPath, cleanup := buildBinary(t)
	defer cleanup()

	stdout, _, _ := runCLI(t, binaryPath, "--help")

	// Check for API documentation useful for LLMs
	expectedPatterns := []string{
		"curl",
		"POST",
		"/api/tabs",
		"GET",
		"DELETE",
		"markdown",
		"code",
		"diff",
		"Content-Type",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(stdout, pattern) {
			t.Errorf("help should contain %q for LLM-friendly documentation", pattern)
		}
	}
}

// === CLI Error Tests ===

// TestCLI_UnknownCommand tests unknown command error.
func TestCLI_UnknownCommand(t *testing.T) {
	binaryPath, cleanup := buildBinary(t)
	defer cleanup()

	_, stderr, exitCode := runCLI(t, binaryPath, "invalid-command")

	if exitCode == 0 {
		t.Error("expected non-zero exit code for unknown command")
	}

	if !strings.Contains(stderr, "Unknown command") {
		t.Errorf("expected stderr to contain 'Unknown command', got: %s", stderr)
	}

	if !strings.Contains(stderr, "agentviewer --help") {
		t.Errorf("expected stderr to suggest --help, got: %s", stderr)
	}
}

// === CLI Serve Tests ===

// TestCLI_ServeStartsServer tests that 'serve' starts a working server.
func TestCLI_ServeStartsServer(t *testing.T) {
	binaryPath, cleanup := buildBinary(t)
	defer cleanup()

	// Use a random high port to avoid conflicts
	port := "19333"
	baseURL := "http://127.0.0.1:" + port

	// Start server in background
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "serve", "--port", port)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// Wait for server to be ready
	if !waitForServer(t, baseURL, 5*time.Second) {
		t.Fatalf("server did not start. stdout: %s, stderr: %s", stdout.String(), stderr.String())
	}

	// Test basic API endpoint
	resp, err := http.Get(baseURL + "/api/status")
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var status StatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		t.Fatalf("failed to parse status response: %v", err)
	}

	if status.Version != Version {
		t.Errorf("expected version %q, got %q", Version, status.Version)
	}
}

// TestCLI_ServeWithFile tests starting server with an initial file.
func TestCLI_ServeWithFile(t *testing.T) {
	binaryPath, cleanup := buildBinary(t)
	defer cleanup()

	// Get path to test file
	cwd, _ := os.Getwd()
	testFile := filepath.Join(cwd, "testdata", "sample.md")

	port := "19334"
	baseURL := "http://127.0.0.1:" + port

	// Start server with file argument
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "serve", "--port", port, testFile)
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// Wait for server
	if !waitForServer(t, baseURL, 5*time.Second) {
		t.Fatal("server did not start")
	}

	// Verify initial tab was created
	resp, err := http.Get(baseURL + "/api/tabs")
	if err != nil {
		t.Fatalf("failed to get tabs: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var listResp ListTabsResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		t.Fatalf("failed to parse tabs response: %v", err)
	}

	if len(listResp.Tabs) != 1 {
		t.Fatalf("expected 1 tab, got %d", len(listResp.Tabs))
	}

	tabSummary := listResp.Tabs[0]
	if tabSummary.ID != "initial" {
		t.Errorf("expected tab ID 'initial', got %q", tabSummary.ID)
	}

	if tabSummary.Type != "markdown" {
		t.Errorf("expected type 'markdown', got %q", tabSummary.Type)
	}

	// Get full tab details to check content
	resp, err = http.Get(baseURL + "/api/tabs/initial")
	if err != nil {
		t.Fatalf("failed to get tab details: %v", err)
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	var fullTab Tab
	json.Unmarshal(body, &fullTab)

	if !strings.Contains(fullTab.Content, "Sample Markdown") {
		t.Errorf("expected content to contain 'Sample Markdown', got %q", fullTab.Content)
	}
}

// TestCLI_ServeWithFileAndType tests starting server with file and explicit type.
func TestCLI_ServeWithFileAndType(t *testing.T) {
	binaryPath, cleanup := buildBinary(t)
	defer cleanup()

	// Get path to test file (Go file, but force it as markdown)
	cwd, _ := os.Getwd()
	testFile := filepath.Join(cwd, "testdata", "sample.go")

	port := "19335"
	baseURL := "http://127.0.0.1:" + port

	// Start server with file and explicit type
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "serve", "--port", port, "--type", "code", testFile)
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// Wait for server
	if !waitForServer(t, baseURL, 5*time.Second) {
		t.Fatal("server did not start")
	}

	// Verify tab type
	resp, err := http.Get(baseURL + "/api/tabs")
	if err != nil {
		t.Fatalf("failed to get tabs: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var listResp ListTabsResponse
	json.Unmarshal(body, &listResp)

	if len(listResp.Tabs) != 1 {
		t.Fatalf("expected 1 tab, got %d", len(listResp.Tabs))
	}

	if listResp.Tabs[0].Type != "code" {
		t.Errorf("expected type 'code', got %q", listResp.Tabs[0].Type)
	}
}

// TestCLI_ServeWithTitle tests starting server with custom title.
func TestCLI_ServeWithTitle(t *testing.T) {
	binaryPath, cleanup := buildBinary(t)
	defer cleanup()

	cwd, _ := os.Getwd()
	testFile := filepath.Join(cwd, "testdata", "sample.md")

	port := "19336"
	baseURL := "http://127.0.0.1:" + port

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "serve", "--port", port, "--title", "Custom Title", testFile)
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	if !waitForServer(t, baseURL, 5*time.Second) {
		t.Fatal("server did not start")
	}

	resp, err := http.Get(baseURL + "/api/tabs")
	if err != nil {
		t.Fatalf("failed to get tabs: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var listResp ListTabsResponse
	json.Unmarshal(body, &listResp)

	if len(listResp.Tabs) != 1 {
		t.Fatalf("expected 1 tab, got %d", len(listResp.Tabs))
	}

	if listResp.Tabs[0].Title != "Custom Title" {
		t.Errorf("expected title 'Custom Title', got %q", listResp.Tabs[0].Title)
	}
}

// TestCLI_ServeShortFlags tests that short flags work (-p, -t).
func TestCLI_ServeShortFlags(t *testing.T) {
	binaryPath, cleanup := buildBinary(t)
	defer cleanup()

	cwd, _ := os.Getwd()
	testFile := filepath.Join(cwd, "testdata", "sample.go")

	port := "19337"
	baseURL := "http://127.0.0.1:" + port

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use short flags -p for port and -t for type
	cmd := exec.CommandContext(ctx, binaryPath, "serve", "-p", port, "-t", "code", testFile)
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	if !waitForServer(t, baseURL, 5*time.Second) {
		t.Fatal("server did not start with short flags")
	}

	// Verify server is working
	resp, err := http.Get(baseURL + "/api/status")
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

// TestCLI_ServeWithMissingFile tests error handling for missing file.
func TestCLI_ServeWithMissingFile(t *testing.T) {
	binaryPath, cleanup := buildBinary(t)
	defer cleanup()

	// Run serve with a non-existent file
	_, stderr, exitCode := runCLI(t, binaryPath, "serve", "--port", "19338", "/nonexistent/file.md")

	if exitCode == 0 {
		t.Error("expected non-zero exit code for missing file")
	}

	if !strings.Contains(stderr, "Error reading file") {
		t.Errorf("expected stderr to contain 'Error reading file', got: %s", stderr)
	}
}

// TestCLI_ServerRespondsToAPI tests that a running server handles API requests correctly.
func TestCLI_ServerRespondsToAPI(t *testing.T) {
	binaryPath, cleanup := buildBinary(t)
	defer cleanup()

	port := "19339"
	baseURL := "http://127.0.0.1:" + port

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "serve", "--port", port)
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	if !waitForServer(t, baseURL, 5*time.Second) {
		t.Fatal("server did not start")
	}

	// Test creating a tab via API
	createBody := `{"id": "cli-test", "title": "CLI Test", "type": "markdown", "content": "# CLI Test"}`
	resp, err := http.Post(
		baseURL+"/api/tabs",
		"application/json",
		strings.NewReader(createBody),
	)
	if err != nil {
		t.Fatalf("failed to create tab: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	// Verify tab was created
	resp, err = http.Get(baseURL + "/api/tabs/cli-test")
	if err != nil {
		t.Fatalf("failed to get tab: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tab Tab
	if err := json.Unmarshal(body, &tab); err != nil {
		t.Fatalf("failed to parse tab: %v", err)
	}

	if tab.ID != "cli-test" {
		t.Errorf("expected tab ID 'cli-test', got %q", tab.ID)
	}
	if tab.Title != "CLI Test" {
		t.Errorf("expected title 'CLI Test', got %q", tab.Title)
	}
}

// TestCLI_ServeAutoDetectsCodeLanguage tests that language is auto-detected for code files.
func TestCLI_ServeAutoDetectsCodeLanguage(t *testing.T) {
	binaryPath, cleanup := buildBinary(t)
	defer cleanup()

	cwd, _ := os.Getwd()
	testFile := filepath.Join(cwd, "testdata", "sample.go")

	port := "19340"
	baseURL := "http://127.0.0.1:" + port

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "serve", "--port", port, testFile)
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	if !waitForServer(t, baseURL, 5*time.Second) {
		t.Fatal("server did not start")
	}

	resp, err := http.Get(baseURL + "/api/tabs")
	if err != nil {
		t.Fatalf("failed to get tabs: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var listResp ListTabsResponse
	json.Unmarshal(body, &listResp)

	if len(listResp.Tabs) != 1 {
		t.Fatalf("expected 1 tab, got %d", len(listResp.Tabs))
	}

	// Since sample.go is a Go file, type should be detected as code and language as go
	tabSummary := listResp.Tabs[0]
	if tabSummary.Type != "code" {
		t.Errorf("expected type 'code', got %q", tabSummary.Type)
	}

	// Get full tab details to check language
	resp, err = http.Get(baseURL + "/api/tabs/initial")
	if err != nil {
		t.Fatalf("failed to get tab details: %v", err)
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	var fullTab Tab
	json.Unmarshal(body, &fullTab)

	if fullTab.Language != "go" {
		t.Errorf("expected language 'go', got %q", fullTab.Language)
	}
}

// TestCLI_ServeOutputMessage tests that serve prints startup message.
func TestCLI_ServeOutputMessage(t *testing.T) {
	binaryPath, cleanup := buildBinary(t)
	defer cleanup()

	port := "19341"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "serve", "--port", port)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// Wait for server to print startup message
	baseURL := fmt.Sprintf("http://127.0.0.1:%s", port)
	if !waitForServer(t, baseURL, 5*time.Second) {
		t.Fatal("server did not start")
	}

	// Give a moment for output buffer to fill
	time.Sleep(100 * time.Millisecond)

	output := stdout.String()
	if !strings.Contains(output, "agentviewer server starting") {
		t.Errorf("expected startup message, got: %s", output)
	}
	if !strings.Contains(output, port) {
		t.Errorf("expected port %s in output, got: %s", port, output)
	}
}

// TestCLI_MultipleServersOnDifferentPorts tests running multiple servers concurrently.
func TestCLI_MultipleServersOnDifferentPorts(t *testing.T) {
	binaryPath, cleanup := buildBinary(t)
	defer cleanup()

	ports := []string{"19350", "19351", "19352"}
	var cmds []*exec.Cmd

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start multiple servers
	for _, port := range ports {
		cmd := exec.CommandContext(ctx, binaryPath, "serve", "--port", port)
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start server on port %s: %v", port, err)
		}
		cmds = append(cmds, cmd)
	}

	// Cleanup
	defer func() {
		for _, cmd := range cmds {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()

	// Wait for all servers to be ready
	for _, port := range ports {
		baseURL := "http://127.0.0.1:" + port
		if !waitForServer(t, baseURL, 5*time.Second) {
			t.Fatalf("server on port %s did not start", port)
		}
	}

	// Verify each server is independent
	for i, port := range ports {
		baseURL := "http://127.0.0.1:" + port

		// Create a unique tab on each server
		createBody := fmt.Sprintf(`{"id": "server-%d", "title": "Server %d", "type": "markdown", "content": "# Server %d"}`, i, i, i)
		resp, err := http.Post(baseURL+"/api/tabs", "application/json", strings.NewReader(createBody))
		if err != nil {
			t.Fatalf("failed to create tab on port %s: %v", port, err)
		}
		resp.Body.Close()

		// Verify tab count is 1 on this server
		resp, err = http.Get(baseURL + "/api/status")
		if err != nil {
			t.Fatalf("failed to get status on port %s: %v", port, err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var status StatusResponse
		json.Unmarshal(body, &status)

		if status.Tabs != 1 {
			t.Errorf("expected 1 tab on port %s, got %d", port, status.Tabs)
		}
	}
}
