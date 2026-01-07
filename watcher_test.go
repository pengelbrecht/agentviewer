package main

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestNewFileWatcher verifies FileWatcher creation.
func TestNewFileWatcher(t *testing.T) {
	t.Run("creates watcher with callback", func(t *testing.T) {
		fw, err := NewFileWatcher(func(path string, tabIDs []string) {})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}
		defer fw.Stop()

		if fw.WatchCount() != 0 {
			t.Errorf("expected 0 watched files, got %d", fw.WatchCount())
		}
	})

	t.Run("creates watcher with callbacks struct", func(t *testing.T) {
		fw, err := NewFileWatcherWithCallbacks(FileWatcherCallbacks{
			OnChange: func(path string, tabIDs []string) {},
			OnDelete: func(path string, tabIDs []string) {},
		})
		if err != nil {
			t.Fatalf("NewFileWatcherWithCallbacks failed: %v", err)
		}
		defer fw.Stop()

		if fw.WatchCount() != 0 {
			t.Errorf("expected 0 watched files, got %d", fw.WatchCount())
		}
	})
}

// TestFileWatcherAddRemove tests Add and Remove operations.
func TestFileWatcherAddRemove(t *testing.T) {
	t.Run("add single tab for path", func(t *testing.T) {
		fw, err := NewFileWatcher(func(path string, tabIDs []string) {})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}
		defer fw.Stop()

		// Create temp file
		tmpFile := createTempFile(t, "test content")
		defer os.Remove(tmpFile)

		err = fw.Add(tmpFile, "tab1")
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}

		if fw.WatchCount() != 1 {
			t.Errorf("expected 1 watched file, got %d", fw.WatchCount())
		}

		if fw.PathForTab("tab1") != tmpFile {
			t.Errorf("expected path %q, got %q", tmpFile, fw.PathForTab("tab1"))
		}

		tabs := fw.TabsWatching(tmpFile)
		if len(tabs) != 1 || tabs[0] != "tab1" {
			t.Errorf("expected [tab1], got %v", tabs)
		}
	})

	t.Run("add multiple tabs for same path", func(t *testing.T) {
		fw, err := NewFileWatcher(func(path string, tabIDs []string) {})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}
		defer fw.Stop()

		tmpFile := createTempFile(t, "test content")
		defer os.Remove(tmpFile)

		fw.Add(tmpFile, "tab1")
		fw.Add(tmpFile, "tab2")
		fw.Add(tmpFile, "tab3")

		// Should still be only 1 watched file
		if fw.WatchCount() != 1 {
			t.Errorf("expected 1 watched file, got %d", fw.WatchCount())
		}

		tabs := fw.TabsWatching(tmpFile)
		if len(tabs) != 3 {
			t.Errorf("expected 3 tabs watching, got %d", len(tabs))
		}
	})

	t.Run("remove single tab", func(t *testing.T) {
		fw, err := NewFileWatcher(func(path string, tabIDs []string) {})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}
		defer fw.Stop()

		tmpFile := createTempFile(t, "test content")
		defer os.Remove(tmpFile)

		fw.Add(tmpFile, "tab1")
		fw.Remove("tab1")

		if fw.WatchCount() != 0 {
			t.Errorf("expected 0 watched files, got %d", fw.WatchCount())
		}

		if fw.PathForTab("tab1") != "" {
			t.Errorf("expected empty path for removed tab, got %q", fw.PathForTab("tab1"))
		}
	})

	t.Run("remove one of multiple tabs keeps watch", func(t *testing.T) {
		fw, err := NewFileWatcher(func(path string, tabIDs []string) {})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}
		defer fw.Stop()

		tmpFile := createTempFile(t, "test content")
		defer os.Remove(tmpFile)

		fw.Add(tmpFile, "tab1")
		fw.Add(tmpFile, "tab2")
		fw.Remove("tab1")

		// File should still be watched by tab2
		if fw.WatchCount() != 1 {
			t.Errorf("expected 1 watched file, got %d", fw.WatchCount())
		}

		tabs := fw.TabsWatching(tmpFile)
		if len(tabs) != 1 || tabs[0] != "tab2" {
			t.Errorf("expected [tab2], got %v", tabs)
		}
	})

	t.Run("remove last tab stops watching", func(t *testing.T) {
		fw, err := NewFileWatcher(func(path string, tabIDs []string) {})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}
		defer fw.Stop()

		tmpFile := createTempFile(t, "test content")
		defer os.Remove(tmpFile)

		fw.Add(tmpFile, "tab1")
		fw.Add(tmpFile, "tab2")
		fw.Remove("tab1")
		fw.Remove("tab2")

		if fw.WatchCount() != 0 {
			t.Errorf("expected 0 watched files, got %d", fw.WatchCount())
		}
	})

	t.Run("remove non-existent tab is no-op", func(t *testing.T) {
		fw, err := NewFileWatcher(func(path string, tabIDs []string) {})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}
		defer fw.Stop()

		// Should not panic
		fw.Remove("non-existent")
	})

	t.Run("add tab watching different file removes from old", func(t *testing.T) {
		fw, err := NewFileWatcher(func(path string, tabIDs []string) {})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}
		defer fw.Stop()

		tmpFile1 := createTempFile(t, "content1")
		defer os.Remove(tmpFile1)
		tmpFile2 := createTempFile(t, "content2")
		defer os.Remove(tmpFile2)

		fw.Add(tmpFile1, "tab1")
		fw.Add(tmpFile2, "tab1") // Same tab, different file

		// Should now only watch file2
		if fw.WatchCount() != 1 {
			t.Errorf("expected 1 watched file, got %d", fw.WatchCount())
		}

		if fw.PathForTab("tab1") != tmpFile2 {
			t.Errorf("expected path %q, got %q", tmpFile2, fw.PathForTab("tab1"))
		}

		if len(fw.TabsWatching(tmpFile1)) != 0 {
			t.Error("expected no tabs watching file1")
		}

		tabs := fw.TabsWatching(tmpFile2)
		if len(tabs) != 1 || tabs[0] != "tab1" {
			t.Errorf("expected [tab1] for file2, got %v", tabs)
		}
	})
}

// TestFileWatcherRemovePath tests removing all watches for a path.
func TestFileWatcherRemovePath(t *testing.T) {
	t.Run("removes path and returns affected tabs", func(t *testing.T) {
		fw, err := NewFileWatcher(func(path string, tabIDs []string) {})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}
		defer fw.Stop()

		tmpFile := createTempFile(t, "test content")
		defer os.Remove(tmpFile)

		fw.Add(tmpFile, "tab1")
		fw.Add(tmpFile, "tab2")
		fw.Add(tmpFile, "tab3")

		removedTabs := fw.RemovePath(tmpFile)

		if len(removedTabs) != 3 {
			t.Errorf("expected 3 removed tabs, got %d", len(removedTabs))
		}

		if fw.WatchCount() != 0 {
			t.Errorf("expected 0 watched files, got %d", fw.WatchCount())
		}
	})

	t.Run("remove non-existent path returns nil", func(t *testing.T) {
		fw, err := NewFileWatcher(func(path string, tabIDs []string) {})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}
		defer fw.Stop()

		tabs := fw.RemovePath("/non/existent/path")
		if tabs != nil {
			t.Errorf("expected nil, got %v", tabs)
		}
	})
}

// TestFileWatcherClear tests clearing all watches.
func TestFileWatcherClear(t *testing.T) {
	t.Run("clear removes all watches", func(t *testing.T) {
		fw, err := NewFileWatcher(func(path string, tabIDs []string) {})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}
		defer fw.Stop()

		tmpFile1 := createTempFile(t, "content1")
		defer os.Remove(tmpFile1)
		tmpFile2 := createTempFile(t, "content2")
		defer os.Remove(tmpFile2)

		fw.Add(tmpFile1, "tab1")
		fw.Add(tmpFile2, "tab2")

		if fw.WatchCount() != 2 {
			t.Fatalf("expected 2 watched files, got %d", fw.WatchCount())
		}

		fw.Clear()

		if fw.WatchCount() != 0 {
			t.Errorf("expected 0 watched files after clear, got %d", fw.WatchCount())
		}

		if fw.PathForTab("tab1") != "" {
			t.Errorf("expected empty path for tab1 after clear, got %q", fw.PathForTab("tab1"))
		}
	})
}

// TestFileWatcherOnChange tests the change callback fires on file modification.
func TestFileWatcherOnChange(t *testing.T) {
	t.Run("callback fires on file write", func(t *testing.T) {
		var mu sync.Mutex
		var callbackPath string
		var callbackTabs []string
		callbackCh := make(chan struct{})

		fw, err := NewFileWatcher(func(path string, tabIDs []string) {
			mu.Lock()
			callbackPath = path
			callbackTabs = tabIDs
			mu.Unlock()
			close(callbackCh)
		})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}
		defer fw.Stop()

		// Start the watcher event loop
		go fw.Run()

		tmpFile := createTempFile(t, "initial content")
		defer os.Remove(tmpFile)

		fw.Add(tmpFile, "test-tab")

		// Modify the file
		time.Sleep(50 * time.Millisecond) // Let watcher settle
		err = os.WriteFile(tmpFile, []byte("modified content"), 0644)
		if err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		// Wait for callback with timeout
		select {
		case <-callbackCh:
			// OK
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for change callback")
		}

		mu.Lock()
		defer mu.Unlock()

		if callbackPath != tmpFile {
			t.Errorf("expected path %q, got %q", tmpFile, callbackPath)
		}

		if len(callbackTabs) != 1 || callbackTabs[0] != "test-tab" {
			t.Errorf("expected [test-tab], got %v", callbackTabs)
		}
	})

	t.Run("callback fires for all tabs watching file", func(t *testing.T) {
		var mu sync.Mutex
		var callbackTabs []string
		callbackCh := make(chan struct{})

		fw, err := NewFileWatcher(func(path string, tabIDs []string) {
			mu.Lock()
			callbackTabs = tabIDs
			mu.Unlock()
			close(callbackCh)
		})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}
		defer fw.Stop()

		go fw.Run()

		tmpFile := createTempFile(t, "initial content")
		defer os.Remove(tmpFile)

		fw.Add(tmpFile, "tab1")
		fw.Add(tmpFile, "tab2")
		fw.Add(tmpFile, "tab3")

		time.Sleep(50 * time.Millisecond)
		os.WriteFile(tmpFile, []byte("modified"), 0644)

		select {
		case <-callbackCh:
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for callback")
		}

		mu.Lock()
		defer mu.Unlock()

		if len(callbackTabs) != 3 {
			t.Errorf("expected 3 tabs in callback, got %d", len(callbackTabs))
		}
	})

	t.Run("no callback when no tabs watching", func(t *testing.T) {
		callbackCh := make(chan struct{})

		fw, err := NewFileWatcher(func(path string, tabIDs []string) {
			close(callbackCh)
		})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}
		defer fw.Stop()

		go fw.Run()

		tmpFile := createTempFile(t, "initial content")
		defer os.Remove(tmpFile)

		// Don't add any tabs

		time.Sleep(50 * time.Millisecond)
		os.WriteFile(tmpFile, []byte("modified"), 0644)

		select {
		case <-callbackCh:
			t.Error("callback should not have fired")
		case <-time.After(300 * time.Millisecond):
			// OK - no callback is expected
		}
	})
}

// TestFileWatcherDebouncing tests that rapid events are coalesced.
func TestFileWatcherDebouncing(t *testing.T) {
	t.Run("multiple rapid writes result in single callback", func(t *testing.T) {
		var mu sync.Mutex
		callbackCount := 0
		done := make(chan struct{})

		fw, err := NewFileWatcher(func(path string, tabIDs []string) {
			mu.Lock()
			callbackCount++
			mu.Unlock()
			// Signal on first callback
			select {
			case done <- struct{}{}:
			default:
			}
		})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}
		defer fw.Stop()

		go fw.Run()

		tmpFile := createTempFile(t, "initial")
		defer os.Remove(tmpFile)

		fw.Add(tmpFile, "tab1")
		time.Sleep(50 * time.Millisecond)

		// Write multiple times rapidly (faster than debounce delay)
		for i := 0; i < 5; i++ {
			os.WriteFile(tmpFile, []byte("content "+string(rune('a'+i))), 0644)
			time.Sleep(20 * time.Millisecond) // Faster than debounce delay
		}

		// Wait for debounce to settle
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for callback")
		}

		// Give a bit more time for any additional callbacks
		time.Sleep(300 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()

		// Should have exactly 1 callback due to debouncing
		if callbackCount != 1 {
			t.Errorf("expected 1 callback due to debouncing, got %d", callbackCount)
		}
	})
}

// TestFileWatcherOnDelete tests the delete callback fires on file removal/rename.
func TestFileWatcherOnDelete(t *testing.T) {
	t.Run("callback fires on file delete", func(t *testing.T) {
		var mu sync.Mutex
		var deletePath string
		var deleteTabs []string
		deleteCh := make(chan struct{})

		fw, err := NewFileWatcherWithCallbacks(FileWatcherCallbacks{
			OnChange: func(path string, tabIDs []string) {},
			OnDelete: func(path string, tabIDs []string) {
				mu.Lock()
				deletePath = path
				deleteTabs = tabIDs
				mu.Unlock()
				close(deleteCh)
			},
		})
		if err != nil {
			t.Fatalf("NewFileWatcherWithCallbacks failed: %v", err)
		}
		defer fw.Stop()

		go fw.Run()

		tmpFile := createTempFile(t, "content")
		// Don't defer Remove - we're going to delete it

		fw.Add(tmpFile, "tab1")
		time.Sleep(50 * time.Millisecond)

		os.Remove(tmpFile)

		select {
		case <-deleteCh:
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for delete callback")
		}

		mu.Lock()
		defer mu.Unlock()

		if deletePath != tmpFile {
			t.Errorf("expected path %q, got %q", tmpFile, deletePath)
		}

		if len(deleteTabs) != 1 || deleteTabs[0] != "tab1" {
			t.Errorf("expected [tab1], got %v", deleteTabs)
		}
	})

	t.Run("callback fires on file rename", func(t *testing.T) {
		deleteCh := make(chan struct{})

		fw, err := NewFileWatcherWithCallbacks(FileWatcherCallbacks{
			OnChange: func(path string, tabIDs []string) {},
			OnDelete: func(path string, tabIDs []string) {
				close(deleteCh)
			},
		})
		if err != nil {
			t.Fatalf("NewFileWatcherWithCallbacks failed: %v", err)
		}
		defer fw.Stop()

		go fw.Run()

		tmpFile := createTempFile(t, "content")
		newPath := tmpFile + ".renamed"
		defer os.Remove(tmpFile)
		defer os.Remove(newPath)

		fw.Add(tmpFile, "tab1")
		time.Sleep(50 * time.Millisecond)

		os.Rename(tmpFile, newPath)

		select {
		case <-deleteCh:
			// OK
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for rename/delete callback")
		}
	})
}

// TestFileWatcherConcurrency tests thread-safety.
func TestFileWatcherConcurrency(t *testing.T) {
	t.Run("concurrent add and remove", func(t *testing.T) {
		fw, err := NewFileWatcher(func(path string, tabIDs []string) {})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}
		defer fw.Stop()

		// Create multiple temp files
		var tmpFiles []string
		for i := 0; i < 10; i++ {
			f := createTempFile(t, "content")
			tmpFiles = append(tmpFiles, f)
			defer os.Remove(f)
		}

		var wg sync.WaitGroup

		// Concurrent adds
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				tabID := "tab" + string(rune('a'+idx%26))
				fileIdx := idx % len(tmpFiles)
				fw.Add(tmpFiles[fileIdx], tabID)
			}(i)
		}

		// Concurrent removes
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				tabID := "tab" + string(rune('a'+idx%26))
				fw.Remove(tabID)
			}(i)
		}

		// Concurrent reads
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				fw.WatchCount()
				fw.PathForTab("tab" + string(rune('a'+idx%26)))
				fw.TabsWatching(tmpFiles[idx%len(tmpFiles)])
			}(i)
		}

		wg.Wait()

		// Should not panic and operations should complete
	})
}

// TestFileWatcherStop tests stopping the watcher.
func TestFileWatcherStop(t *testing.T) {
	t.Run("stop ends run loop", func(t *testing.T) {
		fw, err := NewFileWatcher(func(path string, tabIDs []string) {})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}

		runDone := make(chan struct{})
		go func() {
			fw.Run()
			close(runDone)
		}()

		// Give Run a moment to start
		time.Sleep(50 * time.Millisecond)

		fw.Stop()

		select {
		case <-runDone:
			// OK - Run() exited
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for Run to exit after Stop")
		}
	})
}

// TestFileWatcherAddNonExistentFile tests behavior with non-existent files.
func TestFileWatcherAddNonExistentFile(t *testing.T) {
	t.Run("add non-existent file returns error", func(t *testing.T) {
		fw, err := NewFileWatcher(func(path string, tabIDs []string) {})
		if err != nil {
			t.Fatalf("NewFileWatcher failed: %v", err)
		}
		defer fw.Stop()

		err = fw.Add("/non/existent/file/path", "tab1")
		if err == nil {
			t.Error("expected error when adding non-existent file")
		}
	})
}

// createTempFile creates a temporary file with the given content and returns its path.
func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile.txt")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return tmpFile
}
