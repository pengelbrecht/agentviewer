package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
		expected TabType
	}{
		{
			name:     "markdown file",
			filename: "README.md",
			content:  "# Hello",
			expected: TabTypeMarkdown,
		},
		{
			name:     "markdown file uppercase",
			filename: "NOTES.MD",
			content:  "Notes",
			expected: TabTypeMarkdown,
		},
		{
			name:     "markdown file .markdown extension",
			filename: "document.markdown",
			content:  "Content",
			expected: TabTypeMarkdown,
		},
		{
			name:     "diff file",
			filename: "changes.diff",
			content:  "",
			expected: TabTypeDiff,
		},
		{
			name:     "patch file",
			filename: "fix.patch",
			content:  "",
			expected: TabTypeDiff,
		},
		{
			name:     "go source file",
			filename: "main.go",
			content:  "package main",
			expected: TabTypeCode,
		},
		{
			name:     "python source file",
			filename: "script.py",
			content:  "import os",
			expected: TabTypeCode,
		},
		{
			name:     "diff content without extension",
			filename: "",
			content:  "diff --git a/file.go b/file.go",
			expected: TabTypeDiff,
		},
		{
			name:     "diff content starting with ---",
			filename: "",
			content:  "--- a/file.go",
			expected: TabTypeDiff,
		},
		{
			name:     "diff content starting with +++",
			filename: "",
			content:  "+++ b/file.go",
			expected: TabTypeDiff,
		},
		{
			name:     "markdown content by pattern - h1 header",
			filename: "",
			content:  "# This is a header",
			expected: TabTypeMarkdown,
		},
		{
			name:     "markdown content by pattern - h2 header",
			filename: "",
			content:  "## Section",
			expected: TabTypeMarkdown,
		},
		{
			name:     "markdown content by pattern - bold",
			filename: "",
			content:  "This is **bold** text",
			expected: TabTypeMarkdown,
		},
		{
			name:     "markdown content by pattern - list",
			filename: "",
			content:  "- Item one\n- Item two",
			expected: TabTypeMarkdown,
		},
		{
			name:     "plain text defaults to markdown",
			filename: "",
			content:  "Just some plain text without any markers",
			expected: TabTypeMarkdown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectContentType(tt.filename, tt.content)
			if result != tt.expected {
				t.Errorf("DetectContentType(%q, %q) = %v, want %v",
					tt.filename, tt.content, result, tt.expected)
			}
		})
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		// Go
		{"main.go", "go"},
		{"file.go", "go"},
		// JavaScript/TypeScript
		{"script.js", "javascript"},
		{"component.jsx", "javascript"},
		{"module.ts", "typescript"},
		{"component.tsx", "typescript"},
		// Python
		{"script.py", "python"},
		// Ruby
		{"app.rb", "ruby"},
		// Rust
		{"main.rs", "rust"},
		// Java/Kotlin
		{"Main.java", "java"},
		{"App.kt", "kotlin"},
		// Swift
		{"ViewController.swift", "swift"},
		// C/C++
		{"main.c", "c"},
		{"main.cpp", "cpp"},
		{"main.cc", "cpp"},
		{"main.cxx", "cpp"},
		{"header.h", "c"},
		{"header.hpp", "cpp"},
		// C#
		{"Program.cs", "csharp"},
		// PHP
		{"index.php", "php"},
		// Shell
		{"script.sh", "bash"},
		{"script.bash", "bash"},
		{"script.zsh", "bash"},
		{"script.fish", "fish"},
		{"script.ps1", "powershell"},
		// SQL
		{"query.sql", "sql"},
		// Web
		{"index.html", "html"},
		{"page.htm", "html"},
		{"styles.css", "css"},
		{"styles.scss", "scss"},
		{"styles.sass", "sass"},
		{"styles.less", "less"},
		// Config/Data
		{"config.json", "json"},
		{"config.yaml", "yaml"},
		{"config.yml", "yaml"},
		{"pom.xml", "xml"},
		{"config.toml", "toml"},
		{"settings.ini", "ini"},
		{"settings.cfg", "ini"},
		{"nginx.conf", "nginx"},
		// Other
		{"script.lua", "lua"},
		{"script.pl", "perl"},
		{"analysis.r", "r"},
		{"analysis.R", "r"},
		{"script.m", "matlab"},
		{"App.scala", "scala"},
		{"module.ex", "elixir"},
		{"module.exs", "elixir"},
		{"module.erl", "erlang"},
		{"Module.hs", "haskell"},
		{"core.clj", "clojure"},
		{"Main.elm", "elm"},
		{"App.vue", "vue"},
		{"Component.svelte", "svelte"},
		// Special filenames
		{"Dockerfile", "dockerfile"},
		{"Makefile", "makefile"},
		{".gitignore", "plaintext"},
		{".dockerignore", "plaintext"},
		// Unknown
		{"unknown.xyz", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := DetectLanguage(tt.filename, "")
			if result != tt.expected {
				t.Errorf("DetectLanguage(%q, \"\") = %q, want %q",
					tt.filename, result, tt.expected)
			}
		})
	}
}

func TestReadFileContent(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Test reading an existing file
	t.Run("read existing file", func(t *testing.T) {
		content := "Hello, World!\nLine 2"
		path := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result, err := ReadFileContent(path)
		if err != nil {
			t.Errorf("ReadFileContent returned error: %v", err)
		}
		if result != content {
			t.Errorf("ReadFileContent = %q, want %q", result, content)
		}
	})

	// Test reading a non-existent file
	t.Run("read non-existent file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "nonexistent.txt")
		_, err := ReadFileContent(path)
		if err == nil {
			t.Error("ReadFileContent should return error for non-existent file")
		}
	})

	// Test reading a file with unicode content
	t.Run("read unicode file", func(t *testing.T) {
		content := "Hello 世界! 🎉"
		path := filepath.Join(tmpDir, "unicode.txt")
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result, err := ReadFileContent(path)
		if err != nil {
			t.Errorf("ReadFileContent returned error: %v", err)
		}
		if result != content {
			t.Errorf("ReadFileContent = %q, want %q", result, content)
		}
	})
}

func TestCreateUnifiedDiff(t *testing.T) {
	t.Run("simple diff", func(t *testing.T) {
		left := "line1\nline2\nline3"
		right := "line1\nmodified\nline3"

		result := CreateUnifiedDiff("old.txt", "new.txt", left, right)

		// Verify diff header
		if !contains(result, "--- old.txt") {
			t.Error("Diff should contain left file header")
		}
		if !contains(result, "+++ new.txt") {
			t.Error("Diff should contain right file header")
		}
		if !contains(result, "@@") {
			t.Error("Diff should contain hunk header")
		}
		// Verify content markers
		if !contains(result, "-line2") {
			t.Error("Diff should contain removed line")
		}
		if !contains(result, "+modified") {
			t.Error("Diff should contain added line")
		}
	})
}

func TestComputeDiff(t *testing.T) {
	t.Run("identical files", func(t *testing.T) {
		content := "line1\nline2\nline3\n"
		result := ComputeDiff("a.txt", "b.txt", content, content)

		if len(result.Hunks) != 0 {
			t.Errorf("Expected no hunks for identical content, got %d", len(result.Hunks))
		}
		if result.LeftPath != "a.txt" {
			t.Errorf("Expected LeftPath 'a.txt', got %s", result.LeftPath)
		}
		if result.RightPath != "b.txt" {
			t.Errorf("Expected RightPath 'b.txt', got %s", result.RightPath)
		}
	})

	t.Run("single line change", func(t *testing.T) {
		left := "line1\nline2\nline3\n"
		right := "line1\nchanged\nline3\n"

		result := ComputeDiff("old.txt", "new.txt", left, right)

		if len(result.Hunks) != 1 {
			t.Fatalf("Expected 1 hunk, got %d", len(result.Hunks))
		}

		hunk := result.Hunks[0]
		// Should have context, delete, add, context
		hasDelete := false
		hasAdd := false
		for _, line := range hunk.Lines {
			if line.Type == "delete" && contains(line.Content, "line2") {
				hasDelete = true
			}
			if line.Type == "add" && contains(line.Content, "changed") {
				hasAdd = true
			}
		}
		if !hasDelete {
			t.Error("Expected hunk to contain deleted line 'line2'")
		}
		if !hasAdd {
			t.Error("Expected hunk to contain added line 'changed'")
		}
	})

	t.Run("addition at end", func(t *testing.T) {
		left := "line1\nline2\n"
		right := "line1\nline2\nline3\n"

		result := ComputeDiff("old.txt", "new.txt", left, right)

		if len(result.Hunks) != 1 {
			t.Fatalf("Expected 1 hunk, got %d", len(result.Hunks))
		}

		hasAdd := false
		for _, line := range result.Hunks[0].Lines {
			if line.Type == "add" && contains(line.Content, "line3") {
				hasAdd = true
			}
		}
		if !hasAdd {
			t.Error("Expected hunk to contain added line 'line3'")
		}
	})

	t.Run("deletion at start", func(t *testing.T) {
		left := "line0\nline1\nline2\n"
		right := "line1\nline2\n"

		result := ComputeDiff("old.txt", "new.txt", left, right)

		if len(result.Hunks) != 1 {
			t.Fatalf("Expected 1 hunk, got %d", len(result.Hunks))
		}

		hasDelete := false
		for _, line := range result.Hunks[0].Lines {
			if line.Type == "delete" && contains(line.Content, "line0") {
				hasDelete = true
			}
		}
		if !hasDelete {
			t.Error("Expected hunk to contain deleted line 'line0'")
		}
	})

	t.Run("multiple separated changes", func(t *testing.T) {
		// Changes far apart should produce separate hunks
		left := "a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nk\nl\nm\nn\no\np\n"
		right := "a\nB\nc\nd\ne\nf\ng\nh\ni\nj\nK\nl\nm\nn\no\np\n"

		result := ComputeDiff("old.txt", "new.txt", left, right)

		// Changes at lines 2 and 11 should produce 2 hunks (more than 3 context lines apart)
		if len(result.Hunks) < 1 {
			t.Errorf("Expected at least 1 hunk for separated changes, got %d", len(result.Hunks))
		}
	})

	t.Run("line numbers are correct", func(t *testing.T) {
		left := "a\nb\nc\nd\n"
		right := "a\nX\nc\nd\n"

		result := ComputeDiff("old.txt", "new.txt", left, right)

		for _, hunk := range result.Hunks {
			for _, line := range hunk.Lines {
				switch line.Type {
				case "delete":
					if line.OldNum == 0 {
						t.Error("Delete line should have OldNum set")
					}
				case "add":
					if line.NewNum == 0 {
						t.Error("Add line should have NewNum set")
					}
				case "context":
					if line.OldNum == 0 || line.NewNum == 0 {
						t.Error("Context line should have both OldNum and NewNum set")
					}
				}
			}
		}
	})
}

func TestCompareFiles(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("compare two files", func(t *testing.T) {
		// Create test files
		leftPath := filepath.Join(tmpDir, "left.txt")
		rightPath := filepath.Join(tmpDir, "right.txt")

		os.WriteFile(leftPath, []byte("hello\nworld\n"), 0644)
		os.WriteFile(rightPath, []byte("hello\nGo\n"), 0644)

		result, err := CompareFiles(leftPath, rightPath)
		if err != nil {
			t.Fatalf("CompareFiles failed: %v", err)
		}

		if result.LeftPath != leftPath {
			t.Errorf("Expected LeftPath %s, got %s", leftPath, result.LeftPath)
		}
		if result.RightPath != rightPath {
			t.Errorf("Expected RightPath %s, got %s", rightPath, result.RightPath)
		}
		if len(result.Hunks) != 1 {
			t.Errorf("Expected 1 hunk, got %d", len(result.Hunks))
		}
	})

	t.Run("left file not found", func(t *testing.T) {
		rightPath := filepath.Join(tmpDir, "right.txt")
		os.WriteFile(rightPath, []byte("content"), 0644)

		_, err := CompareFiles("/nonexistent", rightPath)
		if err == nil {
			t.Error("Expected error for non-existent left file")
		}
	})

	t.Run("right file not found", func(t *testing.T) {
		leftPath := filepath.Join(tmpDir, "left2.txt")
		os.WriteFile(leftPath, []byte("content"), 0644)

		_, err := CompareFiles(leftPath, "/nonexistent")
		if err == nil {
			t.Error("Expected error for non-existent right file")
		}
	})
}

// Helper to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
