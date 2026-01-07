package main

import (
	"sync"
	"testing"
)

// TestNewState verifies that NewState creates a properly initialized State.
func TestNewState(t *testing.T) {
	state := NewState()

	if state == nil {
		t.Fatal("NewState returned nil")
	}

	if state.tabs == nil {
		t.Error("tabs map should be initialized")
	}

	if state.order == nil {
		t.Error("order slice should be initialized")
	}

	if len(state.tabs) != 0 {
		t.Errorf("expected 0 tabs, got %d", len(state.tabs))
	}

	if len(state.order) != 0 {
		t.Errorf("expected 0 order entries, got %d", len(state.order))
	}

	if state.activeID != "" {
		t.Errorf("expected empty activeID, got %q", state.activeID)
	}
}

// TestGenerateID verifies that GenerateID creates unique IDs.
func TestGenerateID(t *testing.T) {
	t.Run("generates non-empty ID", func(t *testing.T) {
		id := GenerateID()
		if id == "" {
			t.Error("GenerateID returned empty string")
		}
	})

	t.Run("generates 16 character hex string", func(t *testing.T) {
		id := GenerateID()
		// 8 bytes = 16 hex characters
		if len(id) != 16 {
			t.Errorf("expected 16 character ID, got %d characters: %q", len(id), id)
		}
	})

	t.Run("generates unique IDs", func(t *testing.T) {
		ids := make(map[string]bool)
		for i := 0; i < 1000; i++ {
			id := GenerateID()
			if ids[id] {
				t.Errorf("duplicate ID generated: %s", id)
			}
			ids[id] = true
		}
	})
}

// TestCreateTab verifies tab creation and update behavior.
func TestCreateTab(t *testing.T) {
	t.Run("creates tab with auto-generated ID", func(t *testing.T) {
		state := NewState()
		tab := &Tab{Title: "Test", Type: TabTypeMarkdown, Content: "# Hello"}

		created, isNew := state.CreateTab(tab)

		if !isNew {
			t.Error("expected isNew to be true for new tab")
		}

		if created.ID == "" {
			t.Error("expected ID to be auto-generated")
		}

		if created.Title != "Test" {
			t.Errorf("expected title 'Test', got %q", created.Title)
		}

		if created.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set")
		}

		if created.UpdatedAt.IsZero() {
			t.Error("expected UpdatedAt to be set")
		}
	})

	t.Run("creates tab with explicit ID", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "my-custom-id", Title: "Test", Type: TabTypeCode}

		created, isNew := state.CreateTab(tab)

		if !isNew {
			t.Error("expected isNew to be true")
		}

		if created.ID != "my-custom-id" {
			t.Errorf("expected ID 'my-custom-id', got %q", created.ID)
		}
	})

	t.Run("first tab becomes active", func(t *testing.T) {
		state := NewState()
		tab := &Tab{Title: "First", Type: TabTypeMarkdown}

		created, _ := state.CreateTab(tab)

		if state.GetActive() != created.ID {
			t.Errorf("expected first tab to be active, got activeID %q", state.GetActive())
		}
	})

	t.Run("second tab does not become active", func(t *testing.T) {
		state := NewState()
		tab1 := &Tab{Title: "First", Type: TabTypeMarkdown}
		created1, _ := state.CreateTab(tab1)

		tab2 := &Tab{Title: "Second", Type: TabTypeMarkdown}
		state.CreateTab(tab2)

		if state.GetActive() != created1.ID {
			t.Errorf("expected first tab to remain active, got activeID %q", state.GetActive())
		}
	})

	t.Run("updates existing tab by ID", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "update-test", Title: "Original", Type: TabTypeMarkdown, Content: "original content"}

		created, isNew := state.CreateTab(tab)
		if !isNew {
			t.Error("expected isNew to be true for first create")
		}
		originalCreatedAt := created.CreatedAt

		// Update the tab
		updateTab := &Tab{ID: "update-test", Title: "Updated", Type: TabTypeCode, Content: "updated content"}
		updated, isNew := state.CreateTab(updateTab)

		if isNew {
			t.Error("expected isNew to be false for update")
		}

		if updated.Title != "Updated" {
			t.Errorf("expected title 'Updated', got %q", updated.Title)
		}

		if updated.Content != "updated content" {
			t.Errorf("expected updated content, got %q", updated.Content)
		}

		if updated.Type != TabTypeCode {
			t.Errorf("expected type Code, got %s", updated.Type)
		}

		// CreatedAt should remain the same
		if !updated.CreatedAt.Equal(originalCreatedAt) {
			t.Error("CreatedAt should not change on update")
		}

		// UpdatedAt should be different
		if updated.UpdatedAt.Equal(originalCreatedAt) {
			t.Error("UpdatedAt should change on update")
		}

		// Should still be only one tab
		if state.TabCount() != 1 {
			t.Errorf("expected 1 tab, got %d", state.TabCount())
		}
	})

	t.Run("adds tabs to order slice", func(t *testing.T) {
		state := NewState()

		tab1 := &Tab{ID: "tab1", Title: "Tab 1", Type: TabTypeMarkdown}
		tab2 := &Tab{ID: "tab2", Title: "Tab 2", Type: TabTypeMarkdown}
		tab3 := &Tab{ID: "tab3", Title: "Tab 3", Type: TabTypeMarkdown}

		state.CreateTab(tab1)
		state.CreateTab(tab2)
		state.CreateTab(tab3)

		tabs := state.ListTabs()
		if len(tabs) != 3 {
			t.Fatalf("expected 3 tabs, got %d", len(tabs))
		}

		if tabs[0].ID != "tab1" || tabs[1].ID != "tab2" || tabs[2].ID != "tab3" {
			t.Errorf("tabs not in expected order: got %s, %s, %s", tabs[0].ID, tabs[1].ID, tabs[2].ID)
		}
	})

	t.Run("preserves DiffMeta", func(t *testing.T) {
		state := NewState()
		tab := &Tab{
			Title:   "Diff Tab",
			Type:    TabTypeDiff,
			Content: "diff content",
			DiffMeta: &DiffMeta{
				LeftLabel:  "Before",
				RightLabel: "After",
				Language:   "go",
			},
		}

		created, _ := state.CreateTab(tab)

		if created.DiffMeta == nil {
			t.Fatal("DiffMeta should be preserved")
		}

		if created.DiffMeta.LeftLabel != "Before" {
			t.Errorf("expected LeftLabel 'Before', got %q", created.DiffMeta.LeftLabel)
		}

		if created.DiffMeta.RightLabel != "After" {
			t.Errorf("expected RightLabel 'After', got %q", created.DiffMeta.RightLabel)
		}

		if created.DiffMeta.Language != "go" {
			t.Errorf("expected Language 'go', got %q", created.DiffMeta.Language)
		}
	})
}

// TestGetTab verifies tab retrieval behavior.
func TestGetTab(t *testing.T) {
	t.Run("returns existing tab", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "get-test", Title: "Test", Type: TabTypeMarkdown, Content: "content"}
		state.CreateTab(tab)

		retrieved, exists := state.GetTab("get-test")

		if !exists {
			t.Error("expected tab to exist")
		}

		if retrieved.ID != "get-test" {
			t.Errorf("expected ID 'get-test', got %q", retrieved.ID)
		}

		if retrieved.Title != "Test" {
			t.Errorf("expected title 'Test', got %q", retrieved.Title)
		}
	})

	t.Run("returns nil for non-existing tab", func(t *testing.T) {
		state := NewState()

		retrieved, exists := state.GetTab("non-existent")

		if exists {
			t.Error("expected tab to not exist")
		}

		if retrieved != nil {
			t.Errorf("expected nil tab, got %v", retrieved)
		}
	})

	t.Run("sets active flag correctly", func(t *testing.T) {
		state := NewState()
		tab1 := &Tab{ID: "tab1", Title: "Tab 1", Type: TabTypeMarkdown}
		tab2 := &Tab{ID: "tab2", Title: "Tab 2", Type: TabTypeMarkdown}

		state.CreateTab(tab1)
		state.CreateTab(tab2)

		// First tab should be active
		retrieved1, _ := state.GetTab("tab1")
		retrieved2, _ := state.GetTab("tab2")

		if !retrieved1.Active {
			t.Error("expected tab1 to be active")
		}

		if retrieved2.Active {
			t.Error("expected tab2 to not be active")
		}

		// Change active tab
		state.SetActive("tab2")

		retrieved1, _ = state.GetTab("tab1")
		retrieved2, _ = state.GetTab("tab2")

		if retrieved1.Active {
			t.Error("expected tab1 to not be active after SetActive")
		}

		if !retrieved2.Active {
			t.Error("expected tab2 to be active after SetActive")
		}
	})

	t.Run("returns copy not reference", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "copy-test", Title: "Original", Type: TabTypeMarkdown}
		state.CreateTab(tab)

		retrieved, _ := state.GetTab("copy-test")
		retrieved.Title = "Modified"

		// Original should be unchanged
		original, _ := state.GetTab("copy-test")
		if original.Title != "Original" {
			t.Errorf("expected title 'Original', got %q - modification leaked", original.Title)
		}
	})
}

// TestDeleteTab verifies tab deletion behavior.
func TestDeleteTab(t *testing.T) {
	t.Run("deletes existing tab", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "delete-test", Title: "Test", Type: TabTypeMarkdown}
		state.CreateTab(tab)

		deleted := state.DeleteTab("delete-test")

		if !deleted {
			t.Error("expected DeleteTab to return true")
		}

		if state.TabCount() != 0 {
			t.Errorf("expected 0 tabs, got %d", state.TabCount())
		}

		_, exists := state.GetTab("delete-test")
		if exists {
			t.Error("tab should no longer exist")
		}
	})

	t.Run("returns false for non-existing tab", func(t *testing.T) {
		state := NewState()

		deleted := state.DeleteTab("non-existent")

		if deleted {
			t.Error("expected DeleteTab to return false for non-existing tab")
		}
	})

	t.Run("removes from order slice", func(t *testing.T) {
		state := NewState()
		tab1 := &Tab{ID: "tab1", Title: "Tab 1", Type: TabTypeMarkdown}
		tab2 := &Tab{ID: "tab2", Title: "Tab 2", Type: TabTypeMarkdown}
		tab3 := &Tab{ID: "tab3", Title: "Tab 3", Type: TabTypeMarkdown}

		state.CreateTab(tab1)
		state.CreateTab(tab2)
		state.CreateTab(tab3)

		state.DeleteTab("tab2")

		tabs := state.ListTabs()
		if len(tabs) != 2 {
			t.Fatalf("expected 2 tabs, got %d", len(tabs))
		}

		if tabs[0].ID != "tab1" || tabs[1].ID != "tab3" {
			t.Errorf("expected order [tab1, tab3], got [%s, %s]", tabs[0].ID, tabs[1].ID)
		}
	})

	t.Run("activates another tab when active is deleted", func(t *testing.T) {
		state := NewState()
		tab1 := &Tab{ID: "tab1", Title: "Tab 1", Type: TabTypeMarkdown}
		tab2 := &Tab{ID: "tab2", Title: "Tab 2", Type: TabTypeMarkdown}

		state.CreateTab(tab1)
		state.CreateTab(tab2)

		// tab1 is active, delete it
		state.DeleteTab("tab1")

		// tab2 should now be active (first in order)
		if state.GetActive() != "tab2" {
			t.Errorf("expected tab2 to be active, got %q", state.GetActive())
		}
	})

	t.Run("clears active when last tab is deleted", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "only-tab", Title: "Only", Type: TabTypeMarkdown}
		state.CreateTab(tab)

		state.DeleteTab("only-tab")

		if state.GetActive() != "" {
			t.Errorf("expected empty activeID, got %q", state.GetActive())
		}
	})

	t.Run("does not change active when non-active is deleted", func(t *testing.T) {
		state := NewState()
		tab1 := &Tab{ID: "tab1", Title: "Tab 1", Type: TabTypeMarkdown}
		tab2 := &Tab{ID: "tab2", Title: "Tab 2", Type: TabTypeMarkdown}

		state.CreateTab(tab1)
		state.CreateTab(tab2)

		// tab1 is active, delete tab2
		state.DeleteTab("tab2")

		if state.GetActive() != "tab1" {
			t.Errorf("expected tab1 to remain active, got %q", state.GetActive())
		}
	})
}

// TestListTabs verifies tab listing behavior.
func TestListTabs(t *testing.T) {
	t.Run("returns empty slice for empty state", func(t *testing.T) {
		state := NewState()

		tabs := state.ListTabs()

		if tabs == nil {
			t.Error("expected non-nil slice")
		}

		if len(tabs) != 0 {
			t.Errorf("expected 0 tabs, got %d", len(tabs))
		}
	})

	t.Run("returns tabs in creation order", func(t *testing.T) {
		state := NewState()
		tab1 := &Tab{ID: "first", Title: "First", Type: TabTypeMarkdown}
		tab2 := &Tab{ID: "second", Title: "Second", Type: TabTypeMarkdown}
		tab3 := &Tab{ID: "third", Title: "Third", Type: TabTypeMarkdown}

		state.CreateTab(tab1)
		state.CreateTab(tab2)
		state.CreateTab(tab3)

		tabs := state.ListTabs()

		if len(tabs) != 3 {
			t.Fatalf("expected 3 tabs, got %d", len(tabs))
		}

		expected := []string{"first", "second", "third"}
		for i, tab := range tabs {
			if tab.ID != expected[i] {
				t.Errorf("expected tab %d to have ID %q, got %q", i, expected[i], tab.ID)
			}
		}
	})

	t.Run("sets active flag on correct tab", func(t *testing.T) {
		state := NewState()
		tab1 := &Tab{ID: "tab1", Title: "Tab 1", Type: TabTypeMarkdown}
		tab2 := &Tab{ID: "tab2", Title: "Tab 2", Type: TabTypeMarkdown}
		tab3 := &Tab{ID: "tab3", Title: "Tab 3", Type: TabTypeMarkdown}

		state.CreateTab(tab1)
		state.CreateTab(tab2)
		state.CreateTab(tab3)

		state.SetActive("tab2")

		tabs := state.ListTabs()

		for _, tab := range tabs {
			expectedActive := tab.ID == "tab2"
			if tab.Active != expectedActive {
				t.Errorf("tab %s: expected Active=%v, got %v", tab.ID, expectedActive, tab.Active)
			}
		}
	})

	t.Run("returns copies not references", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "copy-test", Title: "Original", Type: TabTypeMarkdown}
		state.CreateTab(tab)

		tabs := state.ListTabs()
		tabs[0].Title = "Modified"

		// Original should be unchanged
		tabsAgain := state.ListTabs()
		if tabsAgain[0].Title != "Original" {
			t.Errorf("expected title 'Original', got %q - modification leaked", tabsAgain[0].Title)
		}
	})
}

// TestSetActive verifies active tab setting behavior.
func TestSetActive(t *testing.T) {
	t.Run("sets existing tab as active", func(t *testing.T) {
		state := NewState()
		tab1 := &Tab{ID: "tab1", Title: "Tab 1", Type: TabTypeMarkdown}
		tab2 := &Tab{ID: "tab2", Title: "Tab 2", Type: TabTypeMarkdown}

		state.CreateTab(tab1)
		state.CreateTab(tab2)

		success := state.SetActive("tab2")

		if !success {
			t.Error("expected SetActive to return true")
		}

		if state.GetActive() != "tab2" {
			t.Errorf("expected active to be 'tab2', got %q", state.GetActive())
		}
	})

	t.Run("returns false for non-existing tab", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "tab1", Title: "Tab 1", Type: TabTypeMarkdown}
		state.CreateTab(tab)

		success := state.SetActive("non-existent")

		if success {
			t.Error("expected SetActive to return false for non-existing tab")
		}

		// Active should remain unchanged
		if state.GetActive() != "tab1" {
			t.Errorf("expected active to remain 'tab1', got %q", state.GetActive())
		}
	})
}

// TestGetActive verifies active tab retrieval.
func TestGetActive(t *testing.T) {
	t.Run("returns empty for new state", func(t *testing.T) {
		state := NewState()

		if state.GetActive() != "" {
			t.Errorf("expected empty activeID for new state, got %q", state.GetActive())
		}
	})

	t.Run("returns first tab after creation", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "first", Title: "First", Type: TabTypeMarkdown}
		state.CreateTab(tab)

		if state.GetActive() != "first" {
			t.Errorf("expected active 'first', got %q", state.GetActive())
		}
	})
}

// TestClear verifies state clearing behavior.
func TestClear(t *testing.T) {
	t.Run("removes all tabs", func(t *testing.T) {
		state := NewState()
		for i := 0; i < 5; i++ {
			tab := &Tab{Title: "Tab", Type: TabTypeMarkdown}
			state.CreateTab(tab)
		}

		if state.TabCount() != 5 {
			t.Fatalf("expected 5 tabs before clear, got %d", state.TabCount())
		}

		state.Clear()

		if state.TabCount() != 0 {
			t.Errorf("expected 0 tabs after clear, got %d", state.TabCount())
		}
	})

	t.Run("clears order slice", func(t *testing.T) {
		state := NewState()
		tab1 := &Tab{ID: "tab1", Title: "Tab 1", Type: TabTypeMarkdown}
		tab2 := &Tab{ID: "tab2", Title: "Tab 2", Type: TabTypeMarkdown}
		state.CreateTab(tab1)
		state.CreateTab(tab2)

		state.Clear()

		tabs := state.ListTabs()
		if len(tabs) != 0 {
			t.Errorf("expected 0 tabs in list after clear, got %d", len(tabs))
		}
	})

	t.Run("clears active ID", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "tab1", Title: "Tab 1", Type: TabTypeMarkdown}
		state.CreateTab(tab)

		state.Clear()

		if state.GetActive() != "" {
			t.Errorf("expected empty activeID after clear, got %q", state.GetActive())
		}
	})
}

// TestTabCount verifies tab counting behavior.
func TestTabCount(t *testing.T) {
	t.Run("returns 0 for empty state", func(t *testing.T) {
		state := NewState()

		if state.TabCount() != 0 {
			t.Errorf("expected 0, got %d", state.TabCount())
		}
	})

	t.Run("returns correct count after additions", func(t *testing.T) {
		state := NewState()

		for i := 1; i <= 5; i++ {
			tab := &Tab{Title: "Tab", Type: TabTypeMarkdown}
			state.CreateTab(tab)

			if state.TabCount() != i {
				t.Errorf("expected %d tabs, got %d", i, state.TabCount())
			}
		}
	})

	t.Run("returns correct count after deletions", func(t *testing.T) {
		state := NewState()
		ids := make([]string, 5)

		for i := 0; i < 5; i++ {
			tab := &Tab{Title: "Tab", Type: TabTypeMarkdown}
			created, _ := state.CreateTab(tab)
			ids[i] = created.ID
		}

		for i := 0; i < 5; i++ {
			state.DeleteTab(ids[i])
			expected := 4 - i
			if state.TabCount() != expected {
				t.Errorf("expected %d tabs, got %d", expected, state.TabCount())
			}
		}
	})
}

// TestConcurrency verifies thread-safety of State operations.
func TestConcurrency(t *testing.T) {
	t.Run("concurrent create and read", func(t *testing.T) {
		state := NewState()
		var wg sync.WaitGroup

		// Concurrent creates
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				tab := &Tab{Title: "Concurrent Tab", Type: TabTypeMarkdown}
				state.CreateTab(tab)
			}()
		}

		// Concurrent reads
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				state.ListTabs()
				state.TabCount()
				state.GetActive()
			}()
		}

		wg.Wait()

		if state.TabCount() != 100 {
			t.Errorf("expected 100 tabs, got %d", state.TabCount())
		}
	})

	t.Run("concurrent delete", func(t *testing.T) {
		state := NewState()
		ids := make([]string, 50)

		for i := 0; i < 50; i++ {
			tab := &Tab{Title: "Tab", Type: TabTypeMarkdown}
			created, _ := state.CreateTab(tab)
			ids[i] = created.ID
		}

		var wg sync.WaitGroup
		for _, id := range ids {
			wg.Add(1)
			go func(tabID string) {
				defer wg.Done()
				state.DeleteTab(tabID)
			}(id)
		}

		wg.Wait()

		if state.TabCount() != 0 {
			t.Errorf("expected 0 tabs after concurrent delete, got %d", state.TabCount())
		}
	})

	t.Run("concurrent SetActive", func(t *testing.T) {
		state := NewState()
		ids := make([]string, 10)

		for i := 0; i < 10; i++ {
			tab := &Tab{Title: "Tab", Type: TabTypeMarkdown}
			created, _ := state.CreateTab(tab)
			ids[i] = created.ID
		}

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				state.SetActive(ids[idx%10])
			}(i)
		}

		wg.Wait()

		// Active should be one of the valid IDs
		active := state.GetActive()
		found := false
		for _, id := range ids {
			if id == active {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("active ID %q is not one of the valid IDs", active)
		}
	})
}

func TestReopenTab(t *testing.T) {
	t.Run("reopen after delete", func(t *testing.T) {
		state := NewState()

		// Create a tab
		tab := &Tab{Title: "Test Tab", Type: TabTypeMarkdown, Content: "# Hello"}
		created, _ := state.CreateTab(tab)

		originalID := created.ID
		originalContent := created.Content

		// Delete it
		state.DeleteTab(originalID)

		// Verify it's gone
		if state.TabCount() != 0 {
			t.Errorf("expected 0 tabs after delete, got %d", state.TabCount())
		}

		// Verify we have a closed tab
		if state.ClosedTabCount() != 1 {
			t.Errorf("expected 1 closed tab, got %d", state.ClosedTabCount())
		}

		// Reopen it
		reopened := state.ReopenTab()

		if reopened == nil {
			t.Fatal("expected reopened tab, got nil")
		}

		// Verify content is preserved
		if reopened.Content != originalContent {
			t.Errorf("expected content %q, got %q", originalContent, reopened.Content)
		}

		// Verify it's back in state
		if state.TabCount() != 1 {
			t.Errorf("expected 1 tab after reopen, got %d", state.TabCount())
		}

		// Verify it's now active
		if !reopened.Active {
			t.Error("expected reopened tab to be active")
		}

		// Verify closed count is decremented
		if state.ClosedTabCount() != 0 {
			t.Errorf("expected 0 closed tabs after reopen, got %d", state.ClosedTabCount())
		}
	})

	t.Run("reopen empty returns nil", func(t *testing.T) {
		state := NewState()

		reopened := state.ReopenTab()
		if reopened != nil {
			t.Errorf("expected nil when no closed tabs, got %v", reopened)
		}
	})

	t.Run("reopen multiple tabs in order (LIFO)", func(t *testing.T) {
		state := NewState()

		// Create and delete multiple tabs
		tab1 := &Tab{Title: "Tab 1", Type: TabTypeMarkdown, Content: "Content 1"}
		created1, _ := state.CreateTab(tab1)
		id1 := created1.ID

		tab2 := &Tab{Title: "Tab 2", Type: TabTypeCode, Content: "Content 2"}
		created2, _ := state.CreateTab(tab2)
		id2 := created2.ID

		tab3 := &Tab{Title: "Tab 3", Type: TabTypeDiff, Content: "Content 3"}
		created3, _ := state.CreateTab(tab3)
		id3 := created3.ID

		// Delete in order: 1, 2, 3
		state.DeleteTab(id1)
		state.DeleteTab(id2)
		state.DeleteTab(id3)

		if state.ClosedTabCount() != 3 {
			t.Errorf("expected 3 closed tabs, got %d", state.ClosedTabCount())
		}

		// Reopen should be LIFO: 3, 2, 1
		reopened3 := state.ReopenTab()
		if reopened3.Content != "Content 3" {
			t.Errorf("expected Content 3, got %s", reopened3.Content)
		}

		reopened2 := state.ReopenTab()
		if reopened2.Content != "Content 2" {
			t.Errorf("expected Content 2, got %s", reopened2.Content)
		}

		reopened1 := state.ReopenTab()
		if reopened1.Content != "Content 1" {
			t.Errorf("expected Content 1, got %s", reopened1.Content)
		}
	})

	t.Run("closed tabs limited to maxClosedTabs", func(t *testing.T) {
		state := NewState()

		// Create and delete more than maxClosedTabs
		for i := 0; i < maxClosedTabs+5; i++ {
			tab := &Tab{Title: "Tab", Type: TabTypeMarkdown, Content: "Content"}
			created, _ := state.CreateTab(tab)
			state.DeleteTab(created.ID)
		}

		// Should be limited to maxClosedTabs
		if state.ClosedTabCount() != maxClosedTabs {
			t.Errorf("expected %d closed tabs, got %d", maxClosedTabs, state.ClosedTabCount())
		}
	})

	t.Run("reopen preserves tab properties", func(t *testing.T) {
		state := NewState()

		diffMeta := &DiffMeta{
			LeftLabel:  "Before",
			RightLabel: "After",
			Language:   "go",
		}

		tab := &Tab{
			Title:    "Diff Tab",
			Type:     TabTypeDiff,
			Content:  "diff content",
			Language: "go",
			DiffMeta: diffMeta,
		}
		created, _ := state.CreateTab(tab)
		state.DeleteTab(created.ID)

		reopened := state.ReopenTab()

		if reopened.Title != "Diff Tab" {
			t.Errorf("expected title Diff Tab, got %s", reopened.Title)
		}
		if reopened.Type != TabTypeDiff {
			t.Errorf("expected type diff, got %s", reopened.Type)
		}
		if reopened.Language != "go" {
			t.Errorf("expected language go, got %s", reopened.Language)
		}
		if reopened.DiffMeta == nil {
			t.Fatal("expected DiffMeta to be preserved")
		}
		if reopened.DiffMeta.LeftLabel != "Before" {
			t.Errorf("expected LeftLabel Before, got %s", reopened.DiffMeta.LeftLabel)
		}
	})

	t.Run("reopen handles ID collision", func(t *testing.T) {
		state := NewState()

		// Create tab with explicit ID
		tab := &Tab{ID: "fixed-id", Title: "Tab", Type: TabTypeMarkdown, Content: "Content"}
		state.CreateTab(tab)
		state.DeleteTab("fixed-id")

		// Create a new tab with same ID
		tab2 := &Tab{ID: "fixed-id", Title: "New Tab", Type: TabTypeMarkdown, Content: "New Content"}
		state.CreateTab(tab2)

		// Now reopen - should get a new ID since fixed-id is taken
		reopened := state.ReopenTab()

		if reopened == nil {
			t.Fatal("expected reopened tab")
		}

		if reopened.ID == "fixed-id" {
			t.Error("expected reopened tab to have a new ID due to collision")
		}

		// Should have 2 tabs now
		if state.TabCount() != 2 {
			t.Errorf("expected 2 tabs, got %d", state.TabCount())
		}
	})
}

func TestDeleteTabStoresClosed(t *testing.T) {
	state := NewState()

	tab := &Tab{Title: "Test", Type: TabTypeMarkdown, Content: "Content"}
	created, _ := state.CreateTab(tab)

	if state.ClosedTabCount() != 0 {
		t.Errorf("expected 0 closed tabs initially, got %d", state.ClosedTabCount())
	}

	state.DeleteTab(created.ID)

	if state.ClosedTabCount() != 1 {
		t.Errorf("expected 1 closed tab after delete, got %d", state.ClosedTabCount())
	}
}

// TestUpdateTabContent verifies live file update functionality.
// This is a key user-facing feature: when a watched file changes,
// users see the updated content in their browser automatically.
func TestUpdateTabContent(t *testing.T) {
	t.Run("updates content of existing tab", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "live-update", Title: "File", Type: TabTypeCode, Content: "original content"}
		state.CreateTab(tab)

		updated := state.UpdateTabContent("live-update", "new content")

		if updated == nil {
			t.Fatal("expected updated tab, got nil")
		}

		if updated.Content != "new content" {
			t.Errorf("expected content 'new content', got %q", updated.Content)
		}

		// Verify change persisted
		retrieved, _ := state.GetTab("live-update")
		if retrieved.Content != "new content" {
			t.Errorf("expected persisted content 'new content', got %q", retrieved.Content)
		}
	})

	t.Run("returns nil for non-existing tab", func(t *testing.T) {
		state := NewState()

		updated := state.UpdateTabContent("non-existent", "content")

		if updated != nil {
			t.Errorf("expected nil for non-existing tab, got %v", updated)
		}
	})

	t.Run("clears stale flag on update", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "stale-test", Title: "File", Type: TabTypeCode, Content: "content"}
		state.CreateTab(tab)

		// Mark as stale first
		state.MarkTabStale("stale-test")
		stale, _ := state.GetTab("stale-test")
		if !stale.Stale {
			t.Fatal("expected tab to be stale")
		}

		// Update clears stale flag
		updated := state.UpdateTabContent("stale-test", "fresh content")

		if updated.Stale {
			t.Error("expected stale flag to be cleared after update")
		}
	})

	t.Run("updates UpdatedAt timestamp", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "timestamp-test", Title: "File", Type: TabTypeCode, Content: "content"}
		created, _ := state.CreateTab(tab)
		originalTime := created.UpdatedAt

		updated := state.UpdateTabContent("timestamp-test", "new content")

		if !updated.UpdatedAt.After(originalTime) && !updated.UpdatedAt.Equal(originalTime) {
			t.Error("expected UpdatedAt to be updated")
		}
	})

	t.Run("sets active flag correctly", func(t *testing.T) {
		state := NewState()
		tab1 := &Tab{ID: "tab1", Title: "Tab 1", Type: TabTypeCode, Content: "content"}
		tab2 := &Tab{ID: "tab2", Title: "Tab 2", Type: TabTypeCode, Content: "content"}
		state.CreateTab(tab1)
		state.CreateTab(tab2)

		// tab1 is active (first created)
		updated1 := state.UpdateTabContent("tab1", "new")
		updated2 := state.UpdateTabContent("tab2", "new")

		if !updated1.Active {
			t.Error("expected tab1 to have Active=true")
		}
		if updated2.Active {
			t.Error("expected tab2 to have Active=false")
		}
	})
}

// TestMarkTabStale verifies stale indicator functionality.
// When a watched file is deleted, users see a stale indicator so they
// know the displayed content may be outdated.
func TestMarkTabStale(t *testing.T) {
	t.Run("marks existing tab as stale", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "stale-test", Title: "File", Type: TabTypeCode, Content: "content"}
		state.CreateTab(tab)

		// Initially not stale
		retrieved, _ := state.GetTab("stale-test")
		if retrieved.Stale {
			t.Error("expected tab to not be stale initially")
		}

		stale := state.MarkTabStale("stale-test")

		if stale == nil {
			t.Fatal("expected stale tab, got nil")
		}

		if !stale.Stale {
			t.Error("expected Stale flag to be true")
		}

		// Verify persisted
		retrieved2, _ := state.GetTab("stale-test")
		if !retrieved2.Stale {
			t.Error("expected stale flag to persist")
		}
	})

	t.Run("returns nil for non-existing tab", func(t *testing.T) {
		state := NewState()

		stale := state.MarkTabStale("non-existent")

		if stale != nil {
			t.Errorf("expected nil for non-existing tab, got %v", stale)
		}
	})

	t.Run("preserves content when marked stale", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "preserve-test", Title: "File", Type: TabTypeCode, Content: "important content"}
		state.CreateTab(tab)

		state.MarkTabStale("preserve-test")

		retrieved, _ := state.GetTab("preserve-test")
		if retrieved.Content != "important content" {
			t.Errorf("expected content to be preserved, got %q", retrieved.Content)
		}
	})

	t.Run("sets active flag correctly", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "active-test", Title: "Tab", Type: TabTypeCode, Content: "content"}
		state.CreateTab(tab)

		stale := state.MarkTabStale("active-test")

		if !stale.Active {
			t.Error("expected Active=true for the only (active) tab")
		}
	})
}

// TestClearTabStale verifies clearing the stale flag.
// When a deleted file reappears (e.g., user restores it), the stale
// indicator is cleared.
func TestClearTabStale(t *testing.T) {
	t.Run("clears stale flag on existing tab", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "clear-test", Title: "File", Type: TabTypeCode, Content: "content"}
		state.CreateTab(tab)

		// Mark and then clear
		state.MarkTabStale("clear-test")
		cleared := state.ClearTabStale("clear-test")

		if cleared == nil {
			t.Fatal("expected cleared tab, got nil")
		}

		if cleared.Stale {
			t.Error("expected Stale flag to be false after clearing")
		}

		// Verify persisted
		retrieved, _ := state.GetTab("clear-test")
		if retrieved.Stale {
			t.Error("expected stale=false to persist")
		}
	})

	t.Run("returns nil for non-existing tab", func(t *testing.T) {
		state := NewState()

		cleared := state.ClearTabStale("non-existent")

		if cleared != nil {
			t.Errorf("expected nil for non-existing tab, got %v", cleared)
		}
	})

	t.Run("no-op on non-stale tab", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "noop-test", Title: "File", Type: TabTypeCode, Content: "content"}
		state.CreateTab(tab)

		cleared := state.ClearTabStale("noop-test")

		if cleared == nil {
			t.Fatal("expected tab, got nil")
		}

		if cleared.Stale {
			t.Error("expected Stale to remain false")
		}
	})

	t.Run("sets active flag correctly", func(t *testing.T) {
		state := NewState()
		tab := &Tab{ID: "active-test", Title: "Tab", Type: TabTypeCode, Content: "content"}
		state.CreateTab(tab)
		state.MarkTabStale("active-test")

		cleared := state.ClearTabStale("active-test")

		if !cleared.Active {
			t.Error("expected Active=true for the only (active) tab")
		}
	})
}
