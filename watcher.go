// Package main provides file watcher functionality for agentviewer.
package main

import (
	"log"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher watches files for changes and notifies when they are modified.
// It is safe for concurrent use.
type FileWatcher struct {
	watcher *fsnotify.Watcher

	mu sync.RWMutex
	// pathToTabs maps absolute file paths to sets of tab IDs watching that path.
	// Multiple tabs can watch the same file.
	pathToTabs map[string]map[string]bool
	// tabToPath maps tab IDs to the file path they are watching.
	// Each tab watches at most one file.
	tabToPath map[string]string

	// onChange is called when a watched file changes.
	// Arguments are the file path and a list of tab IDs watching it.
	onChange func(path string, tabIDs []string)

	// done signals shutdown
	done chan struct{}
}

// NewFileWatcher creates a new FileWatcher.
// The onChange callback is invoked when any watched file is modified.
func NewFileWatcher(onChange func(path string, tabIDs []string)) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		watcher:    watcher,
		pathToTabs: make(map[string]map[string]bool),
		tabToPath:  make(map[string]string),
		onChange:   onChange,
		done:       make(chan struct{}),
	}, nil
}

// Run starts the file watcher event loop.
// This should be called in a goroutine. It blocks until Stop() is called.
func (fw *FileWatcher) Run() {
	for {
		select {
		case <-fw.done:
			return

		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			// Handle Write and Create events (file modified or recreated)
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				fw.handleChange(event.Name)
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			// Log errors but continue watching
			log.Printf("file watcher error: %v", err)
		}
	}
}

// handleChange processes a file change event.
func (fw *FileWatcher) handleChange(path string) {
	fw.mu.RLock()
	tabSet, exists := fw.pathToTabs[path]
	if !exists || len(tabSet) == 0 {
		fw.mu.RUnlock()
		return
	}

	// Copy tab IDs to avoid holding lock during callback
	tabIDs := make([]string, 0, len(tabSet))
	for tabID := range tabSet {
		tabIDs = append(tabIDs, tabID)
	}
	fw.mu.RUnlock()

	// Invoke callback outside lock
	if fw.onChange != nil {
		fw.onChange(path, tabIDs)
	}
}

// Add registers a tab to watch a file path.
// The path should be absolute. If the tab is already watching a different file,
// it will be removed from watching that file first.
func (fw *FileWatcher) Add(path, tabID string) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	// If this tab is already watching a different file, remove it first
	if oldPath, exists := fw.tabToPath[tabID]; exists && oldPath != path {
		fw.removeTabFromPathLocked(tabID, oldPath)
	}

	// Add tab to path's watch set
	if fw.pathToTabs[path] == nil {
		fw.pathToTabs[path] = make(map[string]bool)
		// First tab watching this path, start watching the file
		if err := fw.watcher.Add(path); err != nil {
			return err
		}
	}
	fw.pathToTabs[path][tabID] = true
	fw.tabToPath[tabID] = path

	return nil
}

// Remove stops watching a file for a specific tab.
// If no other tabs are watching the file, it stops watching the file entirely.
func (fw *FileWatcher) Remove(tabID string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	path, exists := fw.tabToPath[tabID]
	if !exists {
		return
	}

	fw.removeTabFromPathLocked(tabID, path)
}

// removeTabFromPathLocked removes a tab from watching a path.
// Caller must hold the lock.
func (fw *FileWatcher) removeTabFromPathLocked(tabID, path string) {
	delete(fw.tabToPath, tabID)

	if tabSet, exists := fw.pathToTabs[path]; exists {
		delete(tabSet, tabID)
		// If no more tabs are watching this path, stop watching
		if len(tabSet) == 0 {
			delete(fw.pathToTabs, path)
			fw.watcher.Remove(path)
		}
	}
}

// RemovePath stops watching a file entirely and removes all tabs watching it.
// Returns the list of tab IDs that were watching the path.
func (fw *FileWatcher) RemovePath(path string) []string {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	tabSet, exists := fw.pathToTabs[path]
	if !exists {
		return nil
	}

	tabIDs := make([]string, 0, len(tabSet))
	for tabID := range tabSet {
		tabIDs = append(tabIDs, tabID)
		delete(fw.tabToPath, tabID)
	}

	delete(fw.pathToTabs, path)
	fw.watcher.Remove(path)

	return tabIDs
}

// TabsWatching returns the list of tab IDs watching a specific path.
func (fw *FileWatcher) TabsWatching(path string) []string {
	fw.mu.RLock()
	defer fw.mu.RUnlock()

	tabSet, exists := fw.pathToTabs[path]
	if !exists {
		return nil
	}

	tabIDs := make([]string, 0, len(tabSet))
	for tabID := range tabSet {
		tabIDs = append(tabIDs, tabID)
	}
	return tabIDs
}

// PathForTab returns the path being watched by a tab, or empty string if none.
func (fw *FileWatcher) PathForTab(tabID string) string {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	return fw.tabToPath[tabID]
}

// WatchCount returns the number of unique files being watched.
func (fw *FileWatcher) WatchCount() int {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	return len(fw.pathToTabs)
}

// Stop stops the file watcher and closes all resources.
func (fw *FileWatcher) Stop() error {
	close(fw.done)
	return fw.watcher.Close()
}
