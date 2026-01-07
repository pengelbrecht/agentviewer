package main

import (
	"testing"
)

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
