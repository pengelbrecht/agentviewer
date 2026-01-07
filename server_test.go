package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestOpenBrowser(t *testing.T) {
	// OpenBrowser spawns a real browser process via exec.Command.
	// Cannot be meaningfully tested without side effects (opening browser)
	// or mocking exec.Command. The function is simple enough to verify by inspection.
	t.Skip("OpenBrowser requires spawning real browser - not testable in isolation")
}

// TestHandleFileChange verifies that file changes update tabs and broadcast to clients.
// This tests the live file watching feature users depend on when editing files.
func TestHandleFileChange(t *testing.T) {
	// Create a temp file to read from
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	initialContent := "initial content"
	if err := os.WriteFile(tmpFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Create server without file watcher (we'll call handleFileChange directly)
	state := NewState()
	hub := NewHub()
	s := &Server{
		state: state,
		hub:   hub,
	}
	go hub.Run()
	defer hub.Shutdown()

	// Create a tab pointing to the file
	tab, _ := state.CreateTab(&Tab{
		Title:      "Test File",
		Type:       TabTypeCode,
		Content:    initialContent,
		SourcePath: tmpFile,
	})

	// Collector for broadcast messages
	var mu sync.Mutex
	var broadcasts []WSMessage

	// Register a mock client to capture broadcasts
	mockClient := &Client{
		hub:  hub,
		send: make(chan []byte, 10),
	}
	hub.register <- mockClient

	// Give time for registration
	time.Sleep(20 * time.Millisecond)

	// Update the file content
	newContent := "updated content"
	if err := os.WriteFile(tmpFile, []byte(newContent), 0644); err != nil {
		t.Fatalf("failed to update temp file: %v", err)
	}

	// Call handleFileChange as the file watcher would
	s.handleFileChange(tmpFile, []string{tab.ID})

	// Collect broadcast messages
	timeout := time.After(100 * time.Millisecond)
loop:
	for {
		select {
		case data := <-mockClient.send:
			var msg WSMessage
			if err := json.Unmarshal(data, &msg); err == nil {
				mu.Lock()
				broadcasts = append(broadcasts, msg)
				mu.Unlock()
			}
		case <-timeout:
			break loop
		}
	}

	// Verify the tab was updated
	updatedTab, exists := state.GetTab(tab.ID)
	if !exists {
		t.Fatal("tab not found after update")
	}
	if updatedTab.Content != newContent {
		t.Errorf("tab content = %q, want %q", updatedTab.Content, newContent)
	}

	// Verify broadcast was sent
	mu.Lock()
	defer mu.Unlock()
	if len(broadcasts) == 0 {
		t.Error("expected broadcast message, got none")
	} else {
		msg := broadcasts[0]
		if msg.Type != "tab_updated" {
			t.Errorf("broadcast type = %q, want %q", msg.Type, "tab_updated")
		}
		if msg.Tab == nil || msg.Tab.Content != newContent {
			t.Error("broadcast should contain updated tab with new content")
		}
	}
}

// TestHandleFileDelete verifies that deleted files mark tabs as stale.
// This tests the UX when a user deletes a file that's being viewed.
func TestHandleFileDelete(t *testing.T) {
	// Create server without file watcher (we'll call handleFileDelete directly)
	state := NewState()
	hub := NewHub()
	s := &Server{
		state: state,
		hub:   hub,
	}
	go hub.Run()
	defer hub.Shutdown()

	// Create a tab with some content
	tab, _ := state.CreateTab(&Tab{
		Title:      "Test File",
		Type:       TabTypeCode,
		Content:    "some content",
		SourcePath: "/tmp/deleted.txt",
	})

	// Verify tab is not stale initially
	if tab.Stale {
		t.Fatal("tab should not be stale initially")
	}

	// Register a mock client to capture broadcasts
	mockClient := &Client{
		hub:  hub,
		send: make(chan []byte, 10),
	}
	hub.register <- mockClient
	time.Sleep(20 * time.Millisecond)

	// Call handleFileDelete as the file watcher would
	s.handleFileDelete("/tmp/deleted.txt", []string{tab.ID})

	// Collect broadcasts
	var broadcasts []WSMessage
	timeout := time.After(100 * time.Millisecond)
loop:
	for {
		select {
		case data := <-mockClient.send:
			var msg WSMessage
			if err := json.Unmarshal(data, &msg); err == nil {
				broadcasts = append(broadcasts, msg)
			}
		case <-timeout:
			break loop
		}
	}

	// Verify the tab is now stale
	staleTab, exists := state.GetTab(tab.ID)
	if !exists {
		t.Fatal("tab not found")
	}
	if !staleTab.Stale {
		t.Error("tab should be marked stale after file deletion")
	}

	// Verify broadcast was sent
	if len(broadcasts) == 0 {
		t.Error("expected broadcast message, got none")
	} else {
		msg := broadcasts[0]
		if msg.Type != "tab_stale" {
			t.Errorf("broadcast type = %q, want %q", msg.Type, "tab_stale")
		}
		if msg.Tab == nil || !msg.Tab.Stale {
			t.Error("broadcast should contain stale tab")
		}
	}
}

// TestHandleFileChangeWithUnreadableFile verifies graceful handling when a changed file cannot be read.
func TestHandleFileChangeWithUnreadableFile(t *testing.T) {
	state := NewState()
	hub := NewHub()
	s := &Server{
		state: state,
		hub:   hub,
	}
	go hub.Run()
	defer hub.Shutdown()

	// Create a tab pointing to a non-existent file
	originalContent := "original content"
	tab, _ := state.CreateTab(&Tab{
		Title:      "Missing File",
		Type:       TabTypeCode,
		Content:    originalContent,
		SourcePath: "/nonexistent/path/file.txt",
	})

	// Call handleFileChange for a file that doesn't exist
	// This should not panic and should preserve the original content
	s.handleFileChange("/nonexistent/path/file.txt", []string{tab.ID})

	// Verify the tab content is unchanged (file couldn't be read)
	unchangedTab, exists := state.GetTab(tab.ID)
	if !exists {
		t.Fatal("tab not found")
	}
	if unchangedTab.Content != originalContent {
		t.Errorf("tab content should be unchanged when file cannot be read, got %q", unchangedTab.Content)
	}
}

// TestHandleFileChangeMultipleTabs verifies that a single file change updates all tabs watching it.
func TestHandleFileChangeMultipleTabs(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "shared.txt")
	initialContent := "shared content"
	if err := os.WriteFile(tmpFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	state := NewState()
	hub := NewHub()
	s := &Server{
		state: state,
		hub:   hub,
	}
	go hub.Run()
	defer hub.Shutdown()

	// Create multiple tabs pointing to the same file
	tab1, _ := state.CreateTab(&Tab{
		Title:      "Tab 1",
		Type:       TabTypeCode,
		Content:    initialContent,
		SourcePath: tmpFile,
	})
	tab2, _ := state.CreateTab(&Tab{
		Title:      "Tab 2",
		Type:       TabTypeCode,
		Content:    initialContent,
		SourcePath: tmpFile,
	})

	// Update the file
	newContent := "updated shared content"
	if err := os.WriteFile(tmpFile, []byte(newContent), 0644); err != nil {
		t.Fatalf("failed to update temp file: %v", err)
	}

	// Call handleFileChange with both tab IDs
	s.handleFileChange(tmpFile, []string{tab1.ID, tab2.ID})

	// Verify both tabs were updated
	for _, tabID := range []string{tab1.ID, tab2.ID} {
		updated, exists := state.GetTab(tabID)
		if !exists {
			t.Fatalf("tab %s not found", tabID)
		}
		if !strings.Contains(updated.Content, "updated") {
			t.Errorf("tab %s content = %q, want updated content", tabID, updated.Content)
		}
	}
}
