package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
		expected TabType
	}{
		// File extension based detection
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
			name:     "markdown file mixed case extension",
			filename: "file.Md",
			content:  "",
			expected: TabTypeMarkdown,
		},
		{
			name:     "diff file",
			filename: "changes.diff",
			content:  "",
			expected: TabTypeDiff,
		},
		{
			name:     "diff file uppercase",
			filename: "CHANGES.DIFF",
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
			name:     "patch file uppercase",
			filename: "FIX.PATCH",
			content:  "",
			expected: TabTypeDiff,
		},
		{
			name:     "mermaid file .mmd extension",
			filename: "diagram.mmd",
			content:  "graph TD\n    A --> B",
			expected: TabTypeMermaid,
		},
		{
			name:     "mermaid file .mermaid extension",
			filename: "flowchart.mermaid",
			content:  "sequenceDiagram\n    Alice->>Bob: Hello",
			expected: TabTypeMermaid,
		},
		{
			name:     "mermaid file uppercase MMD",
			filename: "DIAGRAM.MMD",
			content:  "",
			expected: TabTypeMermaid,
		},
		{
			name:     "mermaid file uppercase MERMAID",
			filename: "CHART.MERMAID",
			content:  "",
			expected: TabTypeMermaid,
		},
		{
			name:     "mermaid file mixed case",
			filename: "flow.Mmd",
			content:  "",
			expected: TabTypeMermaid,
		},

		// Code files - extension overrides content
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
			name:     "javascript file",
			filename: "app.js",
			content:  "const x = 1;",
			expected: TabTypeCode,
		},
		{
			name:     "typescript file",
			filename: "app.ts",
			content:  "interface Foo {}",
			expected: TabTypeCode,
		},
		{
			name:     "rust file",
			filename: "main.rs",
			content:  "fn main() {}",
			expected: TabTypeCode,
		},
		{
			name:     "json file",
			filename: "config.json",
			content:  `{"key": "value"}`,
			expected: TabTypeCode,
		},
		{
			name:     "yaml file",
			filename: "config.yaml",
			content:  "key: value",
			expected: TabTypeCode,
		},
		{
			name:     "yml file",
			filename: "config.yml",
			content:  "key: value",
			expected: TabTypeCode,
		},
		{
			name:     "html file",
			filename: "index.html",
			content:  "<!DOCTYPE html>",
			expected: TabTypeCode,
		},
		{
			name:     "css file",
			filename: "style.css",
			content:  "body { margin: 0; }",
			expected: TabTypeCode,
		},
		{
			name:     "shell script",
			filename: "script.sh",
			content:  "#!/bin/bash",
			expected: TabTypeCode,
		},

		// Content-based detection (no filename)
		{
			name:     "diff content without extension - diff prefix",
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
			name:     "diff content with leading whitespace not detected as diff",
			filename: "",
			content:  "  diff --git a/file.go b/file.go",
			expected: TabTypeMarkdown, // HasPrefix fails with leading whitespace
		},

		// Markdown content patterns
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
			name:     "markdown content by pattern - h3 header",
			filename: "",
			content:  "some text\n### Subsection",
			expected: TabTypeMarkdown,
		},
		{
			name:     "markdown content by pattern - bold",
			filename: "",
			content:  "This is **bold** text",
			expected: TabTypeMarkdown,
		},
		{
			name:     "markdown content by pattern - unordered list",
			filename: "",
			content:  "- Item one\n- Item two",
			expected: TabTypeMarkdown,
		},
		{
			name:     "markdown content by pattern - nested list",
			filename: "",
			content:  "- Item one\n  - Nested item",
			expected: TabTypeMarkdown,
		},
		{
			name:     "plain text defaults to markdown",
			filename: "",
			content:  "Just some plain text without any markers",
			expected: TabTypeMarkdown,
		},
		{
			name:     "empty content defaults to markdown",
			filename: "",
			content:  "",
			expected: TabTypeMarkdown,
		},
		{
			name:     "whitespace only defaults to markdown",
			filename: "",
			content:  "   \n\t\n   ",
			expected: TabTypeMarkdown,
		},

		// Edge cases - filename takes priority over content
		{
			name:     "markdown filename overrides code-like content",
			filename: "README.md",
			content:  "package main\nfunc main() {}",
			expected: TabTypeMarkdown,
		},
		{
			name:     "diff filename overrides markdown content",
			filename: "changes.diff",
			content:  "# This looks like markdown",
			expected: TabTypeDiff,
		},
		{
			name:     "code filename with markdown content",
			filename: "main.go",
			content:  "# This is a header",
			expected: TabTypeCode,
		},
		{
			name:     "code filename with diff-like content",
			filename: "script.py",
			content:  "--- old\n+++ new",
			expected: TabTypeCode,
		},

		// Unknown extensions - content-based fallback
		{
			name:     "unknown extension with diff content",
			filename: "file.xyz",
			content:  "diff --git a/b",
			expected: TabTypeDiff,
		},
		{
			name:     "unknown extension with markdown content",
			filename: "file.xyz",
			content:  "# Markdown header",
			expected: TabTypeMarkdown,
		},
		{
			name:     "unknown extension with plain content",
			filename: "file.xyz",
			content:  "Just text",
			expected: TabTypeMarkdown, // defaults to markdown
		},

		// Path handling
		{
			name:     "full path - markdown",
			filename: "/path/to/README.md",
			content:  "",
			expected: TabTypeMarkdown,
		},
		{
			name:     "full path - code",
			filename: "/home/user/project/main.go",
			content:  "",
			expected: TabTypeCode,
		},
		{
			name:     "full path with dots in directory",
			filename: "/path/to/.hidden/file.py",
			content:  "",
			expected: TabTypeCode,
		},
		{
			name:     "relative path",
			filename: "./src/main.go",
			content:  "",
			expected: TabTypeCode,
		},

		// Special cases
		{
			name:     "Dockerfile detected as code",
			filename: "Dockerfile",
			content:  "FROM alpine",
			expected: TabTypeCode,
		},
		{
			name:     "Makefile detected as code",
			filename: "Makefile",
			content:  "all: build",
			expected: TabTypeCode,
		},
		{
			name:     "gitignore detected as code",
			filename: ".gitignore",
			content:  "*.log",
			expected: TabTypeCode,
		},
		{
			name:     "dockerignore detected as code",
			filename: ".dockerignore",
			content:  "node_modules",
			expected: TabTypeCode,
		},

		// Unicode content
		{
			name:     "markdown with unicode",
			filename: "",
			content:  "# 日本語ヘッダー",
			expected: TabTypeMarkdown,
		},
		{
			name:     "code file with unicode content",
			filename: "file.go",
			content:  `const greeting = "こんにちは"`,
			expected: TabTypeCode,
		},

		// Multi-line content detection
		{
			name:     "markdown header on second line",
			filename: "",
			content:  "Some intro text\n# Main Header",
			expected: TabTypeMarkdown,
		},
		{
			name:     "diff marker on second line not detected",
			filename: "",
			content:  "Description\ndiff --git a/b",
			expected: TabTypeMarkdown, // HasPrefix only checks start
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

func TestDetectContentType_ExtensionPriority(t *testing.T) {
	// Verify that file extension always takes priority over content heuristics
	// when the extension is recognized

	t.Run("markdown extension beats code content", func(t *testing.T) {
		result := DetectContentType("README.md", "package main")
		if result != TabTypeMarkdown {
			t.Errorf("Expected markdown for .md file, got %v", result)
		}
	})

	t.Run("diff extension beats markdown content", func(t *testing.T) {
		result := DetectContentType("changes.diff", "# Header")
		if result != TabTypeDiff {
			t.Errorf("Expected diff for .diff file, got %v", result)
		}
	})

	t.Run("code extension beats diff content", func(t *testing.T) {
		result := DetectContentType("file.go", "--- old\n+++ new")
		if result != TabTypeCode {
			t.Errorf("Expected code for .go file, got %v", result)
		}
	})

	t.Run("mermaid extension beats markdown content", func(t *testing.T) {
		result := DetectContentType("diagram.mmd", "# Header\n## Another header")
		if result != TabTypeMermaid {
			t.Errorf("Expected mermaid for .mmd file, got %v", result)
		}
	})

	t.Run("mermaid extension beats code content", func(t *testing.T) {
		result := DetectContentType("flow.mermaid", "package main\nfunc main() {}")
		if result != TabTypeMermaid {
			t.Errorf("Expected mermaid for .mermaid file, got %v", result)
		}
	})
}

func TestDetectContentType_EmptyAndEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
		expected TabType
	}{
		{"both empty", "", "", TabTypeMarkdown},
		{"empty filename, newlines only", "", "\n\n\n", TabTypeMarkdown},
		{"just dots in filename", "...", "", TabTypeMarkdown},
		{"extension only", ".md", "", TabTypeMarkdown},
		{"no extension with spaces", "README", "", TabTypeMarkdown},
		{"multiple extensions - last wins", "file.tar.gz", "", TabTypeMarkdown}, // .gz not recognized
		{"multiple extensions - md wins", "file.backup.md", "", TabTypeMarkdown},
		{"case insensitive - MARKDOWN", "FILE.MARKDOWN", "", TabTypeMarkdown},
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

func TestDetectLanguage_CaseInsensitive(t *testing.T) {
	// File extensions should be case-insensitive
	tests := []struct {
		filename string
		expected string
	}{
		{"FILE.GO", "go"},
		{"File.Go", "go"},
		{"SCRIPT.PY", "python"},
		{"Script.Py", "python"},
		{"APP.JS", "javascript"},
		{"Component.JSX", "javascript"},
		{"Module.TS", "typescript"},
		{"FILE.JAVA", "java"},
		{"MAIN.RS", "rust"},
		{"Config.JSON", "json"},
		{"Config.YAML", "yaml"},
		{"Style.CSS", "css"},
		{"Index.HTML", "html"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := DetectLanguage(tt.filename, "")
			if result != tt.expected {
				t.Errorf("DetectLanguage(%q, \"\") = %q, want %q (case insensitive)",
					tt.filename, result, tt.expected)
			}
		})
	}
}

func TestDetectLanguage_SpecialFilenames(t *testing.T) {
	// Test special filenames that are case-sensitive
	tests := []struct {
		filename string
		expected string
	}{
		// Exact matches
		{"Dockerfile", "dockerfile"},
		{"Makefile", "makefile"},
		{".gitignore", "plaintext"},
		{".dockerignore", "plaintext"},
		// Case variations should NOT match (special names are case-sensitive in current impl)
		{"dockerfile", "dockerfile"},
		{"makefile", "makefile"},
		// Paths with special filenames
		{"/path/to/Dockerfile", "dockerfile"},
		{"./project/Makefile", "makefile"},
		{"src/.gitignore", "plaintext"},
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

func TestDetectLanguage_Paths(t *testing.T) {
	// Test that paths are handled correctly (extension from filename only)
	tests := []struct {
		filename string
		expected string
	}{
		{"/home/user/project/main.go", "go"},
		{"./src/app.js", "javascript"},
		{"../config/settings.yaml", "yaml"},
		{"/path/to/.hidden/file.py", "python"},
		{"./node_modules/@scope/pkg/index.ts", "typescript"},
		{"C:\\Users\\name\\project\\main.rs", "rust"},
		{"/path.with.dots/file.go", "go"},
		{"./dir.name/subdir.name/file.java", "java"},
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

func TestDetectLanguage_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{"empty string", "", ""},
		{"just extension", ".go", "go"},
		{"just dot", ".", ""},
		{"double dot", "..", ""},
		{"hidden file without extension", ".hidden", ""},
		{"hidden file with extension", ".hidden.go", "go"},
		{"multiple extensions - last wins", "file.tar.gz", ""},            // .gz not recognized
		{"multiple extensions - known last", "file.min.js", "javascript"}, // .js is recognized
		{"trailing dot", "file.", ""},
		{"multiple trailing dots", "file...", ""},
		{"spaces in filename", "my file.go", "go"},
		{"unicode in filename", "ファイル.go", "go"},
		{"very long extension", "file.verylongextension", ""},
		{"no extension", "README", ""},
		{"uppercase no extension", "LICENSE", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectLanguage(tt.filename, "")
			if result != tt.expected {
				t.Errorf("DetectLanguage(%q, \"\") = %q, want %q",
					tt.filename, result, tt.expected)
			}
		})
	}
}

func TestDetectLanguage_ContentIgnored(t *testing.T) {
	// The content parameter is currently not used for language detection
	// but ensure it doesn't cause issues
	tests := []struct {
		name     string
		filename string
		content  string
		expected string
	}{
		{"go with content", "main.go", "package main", "go"},
		{"go empty content", "main.go", "", "go"},
		{"go wrong content", "main.go", "this is not go code", "go"},
		{"python with shebang", "script.py", "#!/usr/bin/env python", "python"},
		{"python wrong content", "script.py", "not python at all", "python"},
		{"unknown ext with go content", "unknown.xyz", "package main", ""}, // Content doesn't help unknown extensions
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectLanguage(tt.filename, tt.content)
			if result != tt.expected {
				t.Errorf("DetectLanguage(%q, %q) = %q, want %q",
					tt.filename, tt.content, result, tt.expected)
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
		// Check error message mentions "file not found"
		if err != nil && !strings.Contains(err.Error(), "file not found") {
			t.Errorf("Error should mention 'file not found', got: %v", err)
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

	// Test reading a directory returns error
	t.Run("read directory returns error", func(t *testing.T) {
		dirPath := filepath.Join(tmpDir, "testdir")
		err := os.Mkdir(dirPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		_, err = ReadFileContent(dirPath)
		if err == nil {
			t.Error("ReadFileContent should return error for directory")
		}
		if err != nil && !strings.Contains(err.Error(), "directory") {
			t.Errorf("Error should mention 'directory', got: %v", err)
		}
	})

	// Test reading empty file
	t.Run("read empty file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "empty.txt")
		err := os.WriteFile(path, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result, err := ReadFileContent(path)
		if err != nil {
			t.Errorf("ReadFileContent returned error for empty file: %v", err)
		}
		if result != "" {
			t.Errorf("ReadFileContent = %q, want empty string", result)
		}
	})

	// Test path normalization with trailing slashes and dots
	t.Run("path normalization", func(t *testing.T) {
		content := "normalized content"
		path := filepath.Join(tmpDir, "normalized.txt")
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Try with redundant path components
		messyPath := filepath.Join(tmpDir, ".", "normalized.txt")
		result, err := ReadFileContent(messyPath)
		if err != nil {
			t.Errorf("ReadFileContent returned error: %v", err)
		}
		if result != content {
			t.Errorf("ReadFileContent = %q, want %q", result, content)
		}
	})

	// Test reading large file
	t.Run("read large file", func(t *testing.T) {
		// Create a 1MB file
		largeContent := strings.Repeat("abcdefghij", 100000) // 1MB
		path := filepath.Join(tmpDir, "large.txt")
		err := os.WriteFile(path, []byte(largeContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create large test file: %v", err)
		}

		result, err := ReadFileContent(path)
		if err != nil {
			t.Errorf("ReadFileContent returned error for large file: %v", err)
		}
		if result != largeContent {
			t.Errorf("ReadFileContent returned wrong content length: got %d, want %d",
				len(result), len(largeContent))
		}
	})

	// Test binary file with null bytes
	t.Run("read binary file", func(t *testing.T) {
		binaryContent := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE}
		path := filepath.Join(tmpDir, "binary.bin")
		err := os.WriteFile(path, binaryContent, 0644)
		if err != nil {
			t.Fatalf("Failed to create binary test file: %v", err)
		}

		result, err := ReadFileContent(path)
		if err != nil {
			t.Errorf("ReadFileContent returned error for binary file: %v", err)
		}
		if result != string(binaryContent) {
			t.Errorf("ReadFileContent returned wrong content for binary file")
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

func TestParseUnifiedDiff(t *testing.T) {
	t.Run("basic git diff", func(t *testing.T) {
		diffText := `diff --git a/file.go b/file.go
index abc123..def456 100644
--- a/file.go
+++ b/file.go
@@ -1,3 +1,3 @@
 package main
-old line
+new line
 func main() {}`

		result, err := ParseUnifiedDiff(diffText)
		if err != nil {
			t.Fatalf("ParseUnifiedDiff failed: %v", err)
		}

		if result.LeftPath != "file.go" {
			t.Errorf("Expected LeftPath 'file.go', got %q", result.LeftPath)
		}
		if result.RightPath != "file.go" {
			t.Errorf("Expected RightPath 'file.go', got %q", result.RightPath)
		}
		if len(result.Hunks) != 1 {
			t.Fatalf("Expected 1 hunk, got %d", len(result.Hunks))
		}

		hunk := result.Hunks[0]
		if hunk.OldStart != 1 || hunk.OldLines != 3 {
			t.Errorf("Expected old range 1,3 got %d,%d", hunk.OldStart, hunk.OldLines)
		}
		if hunk.NewStart != 1 || hunk.NewLines != 3 {
			t.Errorf("Expected new range 1,3 got %d,%d", hunk.NewStart, hunk.NewLines)
		}

		// Check we have context, delete, add, context
		if len(hunk.Lines) != 4 {
			t.Errorf("Expected 4 lines, got %d", len(hunk.Lines))
		}

		hasDelete := false
		hasAdd := false
		for _, line := range hunk.Lines {
			if line.Type == "delete" && contains(line.Content, "old line") {
				hasDelete = true
			}
			if line.Type == "add" && contains(line.Content, "new line") {
				hasAdd = true
			}
		}
		if !hasDelete {
			t.Error("Expected delete line")
		}
		if !hasAdd {
			t.Error("Expected add line")
		}
	})

	t.Run("diff without a/ b/ prefix", func(t *testing.T) {
		diffText := `--- old.txt	2024-01-15 10:30:00.000000000 +0000
+++ new.txt	2024-01-15 10:31:00.000000000 +0000
@@ -1,2 +1,2 @@
 hello
-world
+Go`

		result, err := ParseUnifiedDiff(diffText)
		if err != nil {
			t.Fatalf("ParseUnifiedDiff failed: %v", err)
		}

		if result.LeftPath != "old.txt" {
			t.Errorf("Expected LeftPath 'old.txt', got %q", result.LeftPath)
		}
		if result.RightPath != "new.txt" {
			t.Errorf("Expected RightPath 'new.txt', got %q", result.RightPath)
		}
	})

	t.Run("multiple hunks", func(t *testing.T) {
		diffText := `--- a/file.go
+++ b/file.go
@@ -1,3 +1,3 @@
 line1
-old1
+new1
 line3
@@ -10,3 +10,3 @@
 line10
-old2
+new2
 line12`

		result, err := ParseUnifiedDiff(diffText)
		if err != nil {
			t.Fatalf("ParseUnifiedDiff failed: %v", err)
		}

		if len(result.Hunks) != 2 {
			t.Fatalf("Expected 2 hunks, got %d", len(result.Hunks))
		}

		if result.Hunks[0].OldStart != 1 {
			t.Errorf("First hunk OldStart should be 1, got %d", result.Hunks[0].OldStart)
		}
		if result.Hunks[1].OldStart != 10 {
			t.Errorf("Second hunk OldStart should be 10, got %d", result.Hunks[1].OldStart)
		}
	})

	t.Run("new file", func(t *testing.T) {
		diffText := `diff --git a/newfile.go b/newfile.go
new file mode 100644
index 0000000..abc123
--- /dev/null
+++ b/newfile.go
@@ -0,0 +1,3 @@
+package main
+
+func main() {}`

		result, err := ParseUnifiedDiff(diffText)
		if err != nil {
			t.Fatalf("ParseUnifiedDiff failed: %v", err)
		}

		if result.LeftPath != "/dev/null" {
			t.Errorf("Expected LeftPath '/dev/null', got %q", result.LeftPath)
		}
		if result.RightPath != "newfile.go" {
			t.Errorf("Expected RightPath 'newfile.go', got %q", result.RightPath)
		}

		if len(result.Hunks) != 1 {
			t.Fatalf("Expected 1 hunk, got %d", len(result.Hunks))
		}

		// All lines should be additions
		for _, line := range result.Hunks[0].Lines {
			if line.Type != "add" {
				t.Errorf("Expected all lines to be 'add', got %q", line.Type)
			}
		}
	})

	t.Run("deleted file", func(t *testing.T) {
		diffText := `diff --git a/oldfile.go b/oldfile.go
deleted file mode 100644
index abc123..0000000
--- a/oldfile.go
+++ /dev/null
@@ -1,3 +0,0 @@
-package main
-
-func main() {}`

		result, err := ParseUnifiedDiff(diffText)
		if err != nil {
			t.Fatalf("ParseUnifiedDiff failed: %v", err)
		}

		if result.LeftPath != "oldfile.go" {
			t.Errorf("Expected LeftPath 'oldfile.go', got %q", result.LeftPath)
		}
		if result.RightPath != "/dev/null" {
			t.Errorf("Expected RightPath '/dev/null', got %q", result.RightPath)
		}

		// All lines should be deletions
		for _, line := range result.Hunks[0].Lines {
			if line.Type != "delete" {
				t.Errorf("Expected all lines to be 'delete', got %q", line.Type)
			}
		}
	})

	t.Run("line numbers tracking", func(t *testing.T) {
		diffText := `--- a/file.go
+++ b/file.go
@@ -5,4 +5,5 @@
 context1
-deleted
+added1
+added2
 context2`

		result, err := ParseUnifiedDiff(diffText)
		if err != nil {
			t.Fatalf("ParseUnifiedDiff failed: %v", err)
		}

		hunk := result.Hunks[0]

		// First context line should be at old=5, new=5
		if hunk.Lines[0].OldNum != 5 || hunk.Lines[0].NewNum != 5 {
			t.Errorf("First context line numbers wrong: old=%d, new=%d",
				hunk.Lines[0].OldNum, hunk.Lines[0].NewNum)
		}

		// Deleted line should be at old=6, new=0
		if hunk.Lines[1].OldNum != 6 || hunk.Lines[1].NewNum != 0 {
			t.Errorf("Deleted line numbers wrong: old=%d, new=%d",
				hunk.Lines[1].OldNum, hunk.Lines[1].NewNum)
		}

		// First added line should be at old=0, new=6
		if hunk.Lines[2].OldNum != 0 || hunk.Lines[2].NewNum != 6 {
			t.Errorf("First added line numbers wrong: old=%d, new=%d",
				hunk.Lines[2].OldNum, hunk.Lines[2].NewNum)
		}

		// Second added line should be at old=0, new=7
		if hunk.Lines[3].OldNum != 0 || hunk.Lines[3].NewNum != 7 {
			t.Errorf("Second added line numbers wrong: old=%d, new=%d",
				hunk.Lines[3].OldNum, hunk.Lines[3].NewNum)
		}

		// Last context line should be at old=7, new=8
		if hunk.Lines[4].OldNum != 7 || hunk.Lines[4].NewNum != 8 {
			t.Errorf("Last context line numbers wrong: old=%d, new=%d",
				hunk.Lines[4].OldNum, hunk.Lines[4].NewNum)
		}
	})

	t.Run("empty diff", func(t *testing.T) {
		result, err := ParseUnifiedDiff("")
		if err != nil {
			t.Fatalf("ParseUnifiedDiff failed: %v", err)
		}
		if len(result.Hunks) != 0 {
			t.Errorf("Expected 0 hunks for empty diff, got %d", len(result.Hunks))
		}
	})

	t.Run("no newline at end of file marker", func(t *testing.T) {
		diffText := `--- a/file.go
+++ b/file.go
@@ -1,2 +1,2 @@
 line1
-old
\ No newline at end of file
+new`

		result, err := ParseUnifiedDiff(diffText)
		if err != nil {
			t.Fatalf("ParseUnifiedDiff failed: %v", err)
		}

		// The deleted line should not have trailing newline
		for _, line := range result.Hunks[0].Lines {
			if line.Type == "delete" && contains(line.Content, "old") {
				if contains(line.Content, "\n") {
					t.Error("Deleted line should not have newline when '\\' marker present")
				}
			}
		}
	})

	t.Run("implicit count of 1", func(t *testing.T) {
		diffText := `--- a/file.go
+++ b/file.go
@@ -5 +5 @@
-old
+new`

		result, err := ParseUnifiedDiff(diffText)
		if err != nil {
			t.Fatalf("ParseUnifiedDiff failed: %v", err)
		}

		if len(result.Hunks) != 1 {
			t.Fatalf("Expected 1 hunk, got %d", len(result.Hunks))
		}

		hunk := result.Hunks[0]
		if hunk.OldLines != 1 || hunk.NewLines != 1 {
			t.Errorf("Expected implicit count of 1, got old=%d, new=%d",
				hunk.OldLines, hunk.NewLines)
		}
	})

	t.Run("hunk with function context", func(t *testing.T) {
		diffText := `--- a/file.go
+++ b/file.go
@@ -10,3 +10,3 @@ func processData() {
 context
-old
+new`

		result, err := ParseUnifiedDiff(diffText)
		if err != nil {
			t.Fatalf("ParseUnifiedDiff failed: %v", err)
		}

		if len(result.Hunks) != 1 {
			t.Fatalf("Expected 1 hunk, got %d", len(result.Hunks))
		}
	})

	t.Run("preserves unified diff text", func(t *testing.T) {
		diffText := `--- a/file.go
+++ b/file.go
@@ -1,2 +1,2 @@
 line1
-old
+new`

		result, err := ParseUnifiedDiff(diffText)
		if err != nil {
			t.Fatalf("ParseUnifiedDiff failed: %v", err)
		}

		if result.Unified != diffText {
			t.Error("Unified field should preserve original diff text")
		}
	})
}

func TestParseFilePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"a/path/to/file.go", "path/to/file.go"},
		{"b/path/to/file.go", "path/to/file.go"},
		{"path/to/file.go", "path/to/file.go"},
		{"/dev/null", "/dev/null"},
		{"a/file.go\t2024-01-15 10:30:00", "file.go"},
		{"  b/file.go  ", "file.go"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseFilePath(tt.input)
			if result != tt.expected {
				t.Errorf("parseFilePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseHunkHeader(t *testing.T) {
	tests := []struct {
		name                                   string
		input                                  string
		expectErr                              bool
		oldStart, oldLines, newStart, newLines int
	}{
		{
			name:     "standard format",
			input:    "@@ -10,5 +20,6 @@",
			oldStart: 10, oldLines: 5, newStart: 20, newLines: 6,
		},
		{
			name:     "with function context",
			input:    "@@ -10,5 +20,6 @@ func main() {",
			oldStart: 10, oldLines: 5, newStart: 20, newLines: 6,
		},
		{
			name:     "implicit count of 1",
			input:    "@@ -10 +20 @@",
			oldStart: 10, oldLines: 1, newStart: 20, newLines: 1,
		},
		{
			name:     "new file with count 0",
			input:    "@@ -0,0 +1,3 @@",
			oldStart: 1, oldLines: 0, newStart: 1, newLines: 3,
		},
		{
			name:      "invalid - no closing @@",
			input:     "@@ -10,5 +20,6",
			expectErr: true,
		},
		{
			name:      "invalid - missing range",
			input:     "@@ @@",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hunk, err := parseHunkHeader(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if hunk.OldStart != tt.oldStart || hunk.OldLines != tt.oldLines ||
				hunk.NewStart != tt.newStart || hunk.NewLines != tt.newLines {
				t.Errorf("Got -%d,%d +%d,%d, want -%d,%d +%d,%d",
					hunk.OldStart, hunk.OldLines, hunk.NewStart, hunk.NewLines,
					tt.oldStart, tt.oldLines, tt.newStart, tt.newLines)
			}
		})
	}
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

func TestValidatePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	// Create a subdirectory with a file
	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)
	subFile := filepath.Join(subDir, "file.txt")
	os.WriteFile(subFile, []byte("sub"), 0644)

	// Save original config and restore after test
	originalConfig := GetFileAccessConfig()
	defer SetFileAccessConfig(originalConfig)

	tests := []struct {
		name        string
		path        string
		allowedDirs []string
		expectErr   bool
		errContains string
	}{
		{
			name:        "valid absolute path",
			path:        testFile,
			allowedDirs: nil,
			expectErr:   false,
		},
		{
			name:        "empty path",
			path:        "",
			allowedDirs: nil,
			expectErr:   true,
			errContains: "cannot be empty",
		},
		{
			name:        "allowed directory - file inside",
			path:        testFile,
			allowedDirs: []string{tmpDir},
			expectErr:   false,
		},
		{
			name:        "allowed directory - file outside",
			path:        "/etc/passwd",
			allowedDirs: []string{tmpDir},
			expectErr:   true,
			errContains: "not in allowed directories",
		},
		{
			name:        "multiple allowed directories - first match",
			path:        testFile,
			allowedDirs: []string{tmpDir, "/opt"},
			expectErr:   false,
		},
		{
			name:        "multiple allowed directories - second match",
			path:        testFile,
			allowedDirs: []string{"/opt", tmpDir},
			expectErr:   false,
		},
		{
			name:        "path traversal blocked by allowed dirs",
			path:        filepath.Join(subDir, "..", "..", "etc", "passwd"),
			allowedDirs: []string{tmpDir},
			expectErr:   true,
			errContains: "not in allowed directories",
		},
		{
			name:        "relative path within allowed dir normalized correctly",
			path:        filepath.Join(subDir, "..", "test.txt"),
			allowedDirs: []string{tmpDir},
			expectErr:   false, // This resolves to testFile which is in allowed dir
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetFileAccessConfig(&FileAccessConfig{
				AllowedDirs: tt.allowedDirs,
				LogAccess:   false,
			})

			result, err := ValidatePath(tt.path)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == "" {
					t.Error("Expected non-empty result path")
				}
			}
		})
	}
}

func TestIsSubPath(t *testing.T) {
	tests := []struct {
		dirPath  string
		filePath string
		expected bool
	}{
		{"/home/user", "/home/user/file.txt", true},
		{"/home/user", "/home/user/subdir/file.txt", true},
		{"/home/user", "/home/user", true},
		{"/home/user/", "/home/user", true},
		{"/home/user", "/home/user/", true},
		{"/home/user", "/home/userX/file.txt", false},
		{"/home/user", "/home/other/file.txt", false},
		{"/home/user", "/etc/passwd", false},
		{"/home/user/dir", "/home/user", false},
	}

	for _, tt := range tests {
		name := tt.dirPath + " contains " + tt.filePath
		t.Run(name, func(t *testing.T) {
			result := isSubPath(tt.dirPath, tt.filePath)
			if result != tt.expected {
				t.Errorf("isSubPath(%q, %q) = %v, want %v",
					tt.dirPath, tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestFileAccessConfig(t *testing.T) {
	// Save original config
	originalConfig := GetFileAccessConfig()
	defer SetFileAccessConfig(originalConfig)

	t.Run("set and get config", func(t *testing.T) {
		newConfig := &FileAccessConfig{
			AllowedDirs: []string{"/tmp", "/home"},
			LogAccess:   true,
		}
		SetFileAccessConfig(newConfig)

		got := GetFileAccessConfig()
		if len(got.AllowedDirs) != 2 {
			t.Errorf("Expected 2 allowed dirs, got %d", len(got.AllowedDirs))
		}
		if !got.LogAccess {
			t.Error("Expected LogAccess to be true")
		}
	})

	t.Run("config copy is independent", func(t *testing.T) {
		SetFileAccessConfig(&FileAccessConfig{
			AllowedDirs: []string{"/original"},
			LogAccess:   true,
		})

		got := GetFileAccessConfig()
		got.AllowedDirs = append(got.AllowedDirs, "/modified")
		got.LogAccess = false

		// Original should be unchanged
		current := GetFileAccessConfig()
		if len(current.AllowedDirs) != 1 {
			t.Errorf("Original config was modified, expected 1 dir, got %d", len(current.AllowedDirs))
		}
		if !current.LogAccess {
			t.Error("Original config LogAccess was modified")
		}
	})
}

func TestReadFileContentWithSecurityConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	allowedFile := filepath.Join(tmpDir, "allowed", "file.txt")
	os.MkdirAll(filepath.Dir(allowedFile), 0755)
	os.WriteFile(allowedFile, []byte("allowed content"), 0644)

	restrictedFile := filepath.Join(tmpDir, "restricted", "file.txt")
	os.MkdirAll(filepath.Dir(restrictedFile), 0755)
	os.WriteFile(restrictedFile, []byte("restricted content"), 0644)

	// Save original config
	originalConfig := GetFileAccessConfig()
	defer SetFileAccessConfig(originalConfig)

	t.Run("access allowed file", func(t *testing.T) {
		SetFileAccessConfig(&FileAccessConfig{
			AllowedDirs: []string{filepath.Join(tmpDir, "allowed")},
			LogAccess:   false,
		})

		content, err := ReadFileContent(allowedFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if content != "allowed content" {
			t.Errorf("Expected 'allowed content', got %q", content)
		}
	})

	t.Run("access restricted file", func(t *testing.T) {
		SetFileAccessConfig(&FileAccessConfig{
			AllowedDirs: []string{filepath.Join(tmpDir, "allowed")},
			LogAccess:   false,
		})

		_, err := ReadFileContent(restrictedFile)
		if err == nil {
			t.Error("Expected error for restricted file")
		}
		if !strings.Contains(err.Error(), "not in allowed directories") {
			t.Errorf("Expected 'not in allowed directories' error, got: %v", err)
		}
	})

	t.Run("path traversal blocked by allowed dirs", func(t *testing.T) {
		SetFileAccessConfig(&FileAccessConfig{
			AllowedDirs: []string{filepath.Join(tmpDir, "allowed")},
			LogAccess:   false,
		})

		// Try to access with path traversal that escapes allowed dir
		traversalPath := filepath.Join(tmpDir, "allowed", "..", "restricted", "file.txt")
		_, err := ReadFileContent(traversalPath)
		if err == nil {
			t.Error("Expected error for path traversal")
		}
		if !strings.Contains(err.Error(), "not in allowed directories") {
			t.Errorf("Expected 'not in allowed directories' error, got: %v", err)
		}
	})

	t.Run("no restrictions allows all", func(t *testing.T) {
		SetFileAccessConfig(&FileAccessConfig{
			AllowedDirs: nil,
			LogAccess:   false,
		})

		// Both files should be accessible
		_, err1 := ReadFileContent(allowedFile)
		_, err2 := ReadFileContent(restrictedFile)

		if err1 != nil {
			t.Errorf("Unexpected error for allowed file: %v", err1)
		}
		if err2 != nil {
			t.Errorf("Unexpected error for restricted file: %v", err2)
		}
	})
}

func TestFileAccessLogging(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	// Save original config
	originalConfig := GetFileAccessConfig()
	defer SetFileAccessConfig(originalConfig)

	t.Run("logging enabled captures access", func(t *testing.T) {
		var logBuffer strings.Builder
		logger := log.New(&logBuffer, "", 0)

		SetFileAccessConfig(&FileAccessConfig{
			AllowedDirs: nil,
			LogAccess:   true,
			Logger:      logger,
		})

		_, err := ReadFileContent(testFile)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, "ALLOWED") {
			t.Errorf("Expected log to contain 'ALLOWED', got: %s", logOutput)
		}
		if !strings.Contains(logOutput, testFile) {
			t.Errorf("Expected log to contain file path, got: %s", logOutput)
		}
	})

	t.Run("logging captures denied access", func(t *testing.T) {
		var logBuffer strings.Builder
		logger := log.New(&logBuffer, "", 0)

		SetFileAccessConfig(&FileAccessConfig{
			AllowedDirs: []string{"/nonexistent"},
			LogAccess:   true,
			Logger:      logger,
		})

		_, _ = ReadFileContent(testFile)

		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, "DENIED") {
			t.Errorf("Expected log to contain 'DENIED', got: %s", logOutput)
		}
	})

	t.Run("logging disabled produces no output", func(t *testing.T) {
		var logBuffer strings.Builder
		logger := log.New(&logBuffer, "", 0)

		SetFileAccessConfig(&FileAccessConfig{
			AllowedDirs: nil,
			LogAccess:   false,
			Logger:      logger,
		})

		_, _ = ReadFileContent(testFile)

		if logBuffer.Len() > 0 {
			t.Errorf("Expected no log output when logging disabled, got: %s", logBuffer.String())
		}
	})
}

func TestDetectContentType_CSV(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected TabType
	}{
		{"csv lowercase", "data.csv", TabTypeCSV},
		{"csv uppercase", "DATA.CSV", TabTypeCSV},
		{"csv mixed case", "Data.Csv", TabTypeCSV},
		{"csv with path", "/path/to/data.csv", TabTypeCSV},
		{"csv relative path", "./reports/sales.csv", TabTypeCSV},
		{"csv complex filename", "2024-01-report_data.csv", TabTypeCSV},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectContentType(tt.filename, "")
			if result != tt.expected {
				t.Errorf("DetectContentType(%q, \"\") = %v, want %v",
					tt.filename, result, tt.expected)
			}
		})
	}
}

func TestDetectContentType_Images(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected TabType
	}{
		{"png lowercase", "image.png", TabTypeImage},
		{"png uppercase", "IMAGE.PNG", TabTypeImage},
		{"jpg lowercase", "photo.jpg", TabTypeImage},
		{"jpg uppercase", "PHOTO.JPG", TabTypeImage},
		{"jpeg lowercase", "photo.jpeg", TabTypeImage},
		{"jpeg uppercase", "PHOTO.JPEG", TabTypeImage},
		{"gif lowercase", "animation.gif", TabTypeImage},
		{"gif uppercase", "ANIMATION.GIF", TabTypeImage},
		{"svg lowercase", "icon.svg", TabTypeImage},
		{"svg uppercase", "ICON.SVG", TabTypeImage},
		{"webp lowercase", "image.webp", TabTypeImage},
		{"webp uppercase", "IMAGE.WEBP", TabTypeImage},
		{"full path png", "/path/to/image.png", TabTypeImage},
		{"full path jpg", "/home/user/photos/vacation.jpg", TabTypeImage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectContentType(tt.filename, "")
			if result != tt.expected {
				t.Errorf("DetectContentType(%q, \"\") = %v, want %v",
					tt.filename, result, tt.expected)
			}
		})
	}
}

func TestIsImageFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		// Image files
		{"image.png", true},
		{"photo.jpg", true},
		{"photo.jpeg", true},
		{"animation.gif", true},
		{"icon.svg", true},
		{"image.webp", true},
		// Case insensitive
		{"IMAGE.PNG", true},
		{"PHOTO.JPG", true},
		{"PHOTO.JPEG", true},
		// Full paths
		{"/path/to/image.png", true},
		{"./relative/photo.jpg", true},
		// Non-image files
		{"document.pdf", false},
		{"script.js", false},
		{"style.css", false},
		{"readme.md", false},
		{"file.txt", false},
		{"archive.zip", false},
		// Edge cases
		{"", false},
		{".", false},
		{".png", true},
		{"noextension", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := IsImageFile(tt.filename)
			if result != tt.expected {
				t.Errorf("IsImageFile(%q) = %v, want %v",
					tt.filename, result, tt.expected)
			}
		})
	}
}

func TestGetImageMIMEType(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"image.png", "image/png"},
		{"IMAGE.PNG", "image/png"},
		{"photo.jpg", "image/jpeg"},
		{"photo.jpeg", "image/jpeg"},
		{"PHOTO.JPEG", "image/jpeg"},
		{"animation.gif", "image/gif"},
		{"icon.svg", "image/svg+xml"},
		{"image.webp", "image/webp"},
		// Full paths
		{"/path/to/image.png", "image/png"},
		{"./relative/photo.jpg", "image/jpeg"},
		// Unknown returns octet-stream
		{"file.txt", "application/octet-stream"},
		{"unknown.xyz", "application/octet-stream"},
		{"", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := GetImageMIMEType(tt.filename)
			if result != tt.expected {
				t.Errorf("GetImageMIMEType(%q) = %q, want %q",
					tt.filename, result, tt.expected)
			}
		})
	}
}

func TestReadImageAsDataURL(t *testing.T) {
	tmpDir := t.TempDir()

	// Save original config and restore after test
	originalConfig := GetFileAccessConfig()
	defer SetFileAccessConfig(originalConfig)
	SetFileAccessConfig(&FileAccessConfig{AllowedDirs: nil, LogAccess: false})

	t.Run("read png file", func(t *testing.T) {
		// Create a minimal valid PNG file (8x8 red square)
		// PNG header + IHDR chunk (minimal)
		pngData := []byte{
			0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
			0x00, 0x00, 0x00, 0x0D, // IHDR chunk length
			0x49, 0x48, 0x44, 0x52, // IHDR
			0x00, 0x00, 0x00, 0x01, // width: 1
			0x00, 0x00, 0x00, 0x01, // height: 1
			0x08, 0x02, // bit depth 8, color type 2 (RGB)
			0x00, 0x00, 0x00, // compression, filter, interlace
			0x90, 0x77, 0x53, 0xDE, // CRC
			0x00, 0x00, 0x00, 0x0C, // IDAT chunk length
			0x49, 0x44, 0x41, 0x54, // IDAT
			0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F, 0x00, 0x05, 0xFE, 0x02, 0xFE, // compressed data
			0xA3, 0x6D, 0xED, 0xC6, // CRC
			0x00, 0x00, 0x00, 0x00, // IEND chunk length
			0x49, 0x45, 0x4E, 0x44, // IEND
			0xAE, 0x42, 0x60, 0x82, // CRC
		}

		path := filepath.Join(tmpDir, "test.png")
		err := os.WriteFile(path, pngData, 0644)
		if err != nil {
			t.Fatalf("Failed to create test PNG: %v", err)
		}

		result, err := ReadImageAsDataURL(path)
		if err != nil {
			t.Fatalf("ReadImageAsDataURL failed: %v", err)
		}

		if !strings.HasPrefix(result, "data:image/png;base64,") {
			t.Errorf("Expected data URL to start with 'data:image/png;base64,', got prefix: %s",
				result[:min(len(result), 30)])
		}
	})

	t.Run("read jpg file", func(t *testing.T) {
		// Minimal JPEG data (SOI + EOI markers)
		jpgData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0xFF, 0xD9}

		path := filepath.Join(tmpDir, "test.jpg")
		err := os.WriteFile(path, jpgData, 0644)
		if err != nil {
			t.Fatalf("Failed to create test JPG: %v", err)
		}

		result, err := ReadImageAsDataURL(path)
		if err != nil {
			t.Fatalf("ReadImageAsDataURL failed: %v", err)
		}

		if !strings.HasPrefix(result, "data:image/jpeg;base64,") {
			t.Errorf("Expected data URL to start with 'data:image/jpeg;base64,', got prefix: %s",
				result[:min(len(result), 30)])
		}
	})

	t.Run("read svg file", func(t *testing.T) {
		svgData := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
			<circle cx="50" cy="50" r="40" fill="red"/>
		</svg>`

		path := filepath.Join(tmpDir, "test.svg")
		err := os.WriteFile(path, []byte(svgData), 0644)
		if err != nil {
			t.Fatalf("Failed to create test SVG: %v", err)
		}

		result, err := ReadImageAsDataURL(path)
		if err != nil {
			t.Fatalf("ReadImageAsDataURL failed: %v", err)
		}

		if !strings.HasPrefix(result, "data:image/svg+xml;base64,") {
			t.Errorf("Expected data URL to start with 'data:image/svg+xml;base64,', got prefix: %s",
				result[:min(len(result), 30)])
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "nonexistent.png")
		_, err := ReadImageAsDataURL(path)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})

	t.Run("directory instead of file", func(t *testing.T) {
		dirPath := filepath.Join(tmpDir, "testdir.png")
		os.MkdirAll(dirPath, 0755)

		_, err := ReadImageAsDataURL(dirPath)
		if err == nil {
			t.Error("Expected error for directory")
		}
		if !strings.Contains(err.Error(), "directory") {
			t.Errorf("Expected 'directory' error, got: %v", err)
		}
	})

	t.Run("non-image file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(path, []byte("hello"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		_, err = ReadImageAsDataURL(path)
		if err == nil {
			t.Error("Expected error for non-image file")
		}
		if !strings.Contains(err.Error(), "not a recognized image format") {
			t.Errorf("Expected 'not a recognized image format' error, got: %v", err)
		}
	})

	t.Run("empty path", func(t *testing.T) {
		_, err := ReadImageAsDataURL("")
		if err == nil {
			t.Error("Expected error for empty path")
		}
	})
}
