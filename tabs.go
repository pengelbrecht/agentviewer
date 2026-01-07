// Package main provides tab state management.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// TabType represents the type of content in a tab.
type TabType string

const (
	TabTypeMarkdown TabType = "markdown"
	TabTypeCode     TabType = "code"
	TabTypeDiff     TabType = "diff"
)

// Tab represents a single tab in the viewer.
type Tab struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Type       TabType   `json:"type"`
	Content    string    `json:"content"`
	Language   string    `json:"language,omitempty"`
	DiffMeta   *DiffMeta `json:"diff,omitempty"`
	SourcePath string    `json:"sourcePath,omitempty"` // File path for auto-reload; only set when created from file
	Active     bool      `json:"active,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// DiffMeta holds metadata for diff tabs.
type DiffMeta struct {
	LeftLabel  string `json:"leftLabel,omitempty"`
	RightLabel string `json:"rightLabel,omitempty"`
	Language   string `json:"language,omitempty"`
}

// maxClosedTabs is the maximum number of recently closed tabs to keep in memory.
const maxClosedTabs = 10

// State manages all tabs and their ordering.
type State struct {
	mu         sync.RWMutex
	tabs       map[string]*Tab
	order      []string
	activeID   string
	closedTabs []*Tab // Recently closed tabs (stack, most recent last)
}

// NewState creates a new State instance.
func NewState() *State {
	return &State{
		tabs:  make(map[string]*Tab),
		order: make([]string, 0),
	}
}

// GenerateID creates a unique tab ID.
func GenerateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// CreateTab creates a new tab or updates an existing one.
// Returns the tab and whether it was newly created (true) or updated (false).
func (s *State) CreateTab(tab *Tab) (*Tab, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tab.ID == "" {
		tab.ID = GenerateID()
	}

	now := time.Now()
	existing, exists := s.tabs[tab.ID]

	if exists {
		// Update existing tab
		existing.Title = tab.Title
		existing.Type = tab.Type
		existing.Content = tab.Content
		existing.Language = tab.Language
		existing.DiffMeta = tab.DiffMeta
		// Only update SourcePath if provided (don't overwrite with empty)
		if tab.SourcePath != "" {
			existing.SourcePath = tab.SourcePath
		}
		existing.UpdatedAt = now
		return existing, false
	}

	// Create new tab
	tab.CreatedAt = now
	tab.UpdatedAt = now
	s.tabs[tab.ID] = tab
	s.order = append(s.order, tab.ID)

	// If this is the first tab, make it active
	if len(s.tabs) == 1 {
		s.activeID = tab.ID
	}

	return tab, true
}

// GetTab returns a tab by ID.
func (s *State) GetTab(id string) (*Tab, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tab, exists := s.tabs[id]
	if exists {
		// Return a copy with active status
		tabCopy := *tab
		tabCopy.Active = (s.activeID == id)
		return &tabCopy, true
	}
	return nil, false
}

// DeleteTab removes a tab by ID, storing it for potential reopen.
func (s *State) DeleteTab(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	tab, exists := s.tabs[id]
	if !exists {
		return false
	}

	// Store a copy of the tab for potential reopen
	tabCopy := *tab
	s.closedTabs = append(s.closedTabs, &tabCopy)

	// Trim closed tabs if we exceed the limit
	if len(s.closedTabs) > maxClosedTabs {
		s.closedTabs = s.closedTabs[len(s.closedTabs)-maxClosedTabs:]
	}

	delete(s.tabs, id)

	// Remove from order
	for i, oid := range s.order {
		if oid == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}

	// If we deleted the active tab, activate another one
	if s.activeID == id {
		if len(s.order) > 0 {
			s.activeID = s.order[0]
		} else {
			s.activeID = ""
		}
	}

	return true
}

// ListTabs returns all tabs in order.
func (s *State) ListTabs() []*Tab {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tabs := make([]*Tab, 0, len(s.order))
	for _, id := range s.order {
		if tab, exists := s.tabs[id]; exists {
			tabCopy := *tab
			tabCopy.Active = (s.activeID == id)
			tabs = append(tabs, &tabCopy)
		}
	}
	return tabs
}

// SetActive sets the active tab.
func (s *State) SetActive(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tabs[id]; !exists {
		return false
	}

	s.activeID = id
	return true
}

// GetActive returns the active tab ID.
func (s *State) GetActive() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeID
}

// Clear removes all tabs.
func (s *State) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tabs = make(map[string]*Tab)
	s.order = make([]string, 0)
	s.activeID = ""
}

// TabCount returns the number of tabs.
func (s *State) TabCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tabs)
}

// ReopenTab restores the most recently closed tab.
// Returns the reopened tab, or nil if no closed tabs are available.
func (s *State) ReopenTab() *Tab {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.closedTabs) == 0 {
		return nil
	}

	// Pop the most recently closed tab
	tab := s.closedTabs[len(s.closedTabs)-1]
	s.closedTabs = s.closedTabs[:len(s.closedTabs)-1]

	// Check if a tab with this ID already exists (edge case)
	if _, exists := s.tabs[tab.ID]; exists {
		// Generate a new ID to avoid collision
		tab.ID = GenerateID()
	}

	// Update timestamps
	now := time.Now()
	tab.UpdatedAt = now

	// Re-add to state
	s.tabs[tab.ID] = tab
	s.order = append(s.order, tab.ID)
	s.activeID = tab.ID

	// Return a copy
	tabCopy := *tab
	tabCopy.Active = true
	return &tabCopy
}

// ClosedTabCount returns the number of tabs available for reopen.
func (s *State) ClosedTabCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.closedTabs)
}
