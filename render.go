// Package main provides file reading and content type detection.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadFileContent reads a file and returns its content.
func ReadFileContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DetectContentType determines the content type based on file extension and content.
func DetectContentType(filename, content string) TabType {
	if filename != "" {
		ext := strings.ToLower(filepath.Ext(filename))
		switch ext {
		case ".md", ".markdown":
			return TabTypeMarkdown
		case ".diff", ".patch":
			return TabTypeDiff
		}
		// Default to code for known source files
		if lang := DetectLanguage(filename, content); lang != "" {
			return TabTypeCode
		}
	}

	// Check content for diff markers
	if strings.HasPrefix(content, "diff ") ||
		strings.HasPrefix(content, "--- ") ||
		strings.HasPrefix(content, "+++ ") {
		return TabTypeDiff
	}

	// Check for markdown-like content (headers, lists, bold)
	if strings.Contains(content, "# ") ||
		strings.Contains(content, "## ") ||
		strings.Contains(content, "**") ||
		strings.Contains(content, "- ") {
		return TabTypeMarkdown
	}

	// Default to markdown for plain text
	return TabTypeMarkdown
}

// DetectLanguage determines the programming language based on file extension.
func DetectLanguage(filename, content string) string {
	if filename == "" {
		return ""
	}

	ext := strings.ToLower(filepath.Ext(filename))
	base := strings.ToLower(filepath.Base(filename))

	// Special filenames
	switch base {
	case "dockerfile":
		return "dockerfile"
	case "makefile":
		return "makefile"
	case ".gitignore", ".dockerignore":
		return "plaintext"
	}

	// Extension mapping
	extLang := map[string]string{
		".go":    "go",
		".js":    "javascript",
		".jsx":   "javascript",
		".ts":    "typescript",
		".tsx":   "typescript",
		".py":    "python",
		".rb":    "ruby",
		".rs":    "rust",
		".java":  "java",
		".kt":    "kotlin",
		".swift": "swift",
		".c":     "c",
		".cpp":   "cpp",
		".cc":    "cpp",
		".cxx":   "cpp",
		".h":     "c",
		".hpp":   "cpp",
		".cs":    "csharp",
		".php":   "php",
		".sh":    "bash",
		".bash":  "bash",
		".zsh":   "bash",
		".fish":  "fish",
		".ps1":   "powershell",
		".sql":   "sql",
		".html":  "html",
		".htm":   "html",
		".css":   "css",
		".scss":  "scss",
		".sass":  "sass",
		".less":  "less",
		".json":  "json",
		".yaml":  "yaml",
		".yml":   "yaml",
		".xml":   "xml",
		".toml":  "toml",
		".ini":   "ini",
		".cfg":   "ini",
		".conf":  "nginx",
		".lua":   "lua",
		".pl":    "perl",
		".r":     "r",
		".R":     "r",
		".m":     "matlab",
		".scala": "scala",
		".ex":    "elixir",
		".exs":   "elixir",
		".erl":   "erlang",
		".hs":    "haskell",
		".clj":   "clojure",
		".elm":   "elm",
		".vue":   "vue",
		".svelte": "svelte",
	}

	if lang, ok := extLang[ext]; ok {
		return lang
	}

	return ""
}

// CreateUnifiedDiff creates a unified diff from two file contents.
func CreateUnifiedDiff(leftPath, rightPath, leftContent, rightContent string) string {
	// Simple unified diff format
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- %s\n", leftPath))
	sb.WriteString(fmt.Sprintf("+++ %s\n", rightPath))

	leftLines := strings.Split(leftContent, "\n")
	rightLines := strings.Split(rightContent, "\n")

	// Simple line-by-line comparison (basic diff implementation)
	// In production, use a proper diff algorithm
	sb.WriteString(fmt.Sprintf("@@ -1,%d +1,%d @@\n", len(leftLines), len(rightLines)))

	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	for i := 0; i < maxLines; i++ {
		leftLine := ""
		rightLine := ""
		if i < len(leftLines) {
			leftLine = leftLines[i]
		}
		if i < len(rightLines) {
			rightLine = rightLines[i]
		}

		if leftLine == rightLine {
			sb.WriteString(" " + leftLine + "\n")
		} else {
			if i < len(leftLines) {
				sb.WriteString("-" + leftLine + "\n")
			}
			if i < len(rightLines) {
				sb.WriteString("+" + rightLine + "\n")
			}
		}
	}

	return sb.String()
}
