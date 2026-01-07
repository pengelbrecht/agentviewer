// Package main provides file reading and content type detection.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
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

// DiffLine represents a single line in a diff with its type and content.
type DiffLine struct {
	Type    string `json:"type"` // "context", "add", "delete"
	Content string `json:"content"`
	OldNum  int    `json:"oldNum,omitempty"` // Line number in left file (0 if not applicable)
	NewNum  int    `json:"newNum,omitempty"` // Line number in right file (0 if not applicable)
}

// DiffHunk represents a section of changes in a diff.
type DiffHunk struct {
	OldStart int        `json:"oldStart"`
	OldLines int        `json:"oldLines"`
	NewStart int        `json:"newStart"`
	NewLines int        `json:"newLines"`
	Lines    []DiffLine `json:"lines"`
}

// DiffResult represents the complete diff result with structured data.
type DiffResult struct {
	LeftPath  string     `json:"leftPath"`
	RightPath string     `json:"rightPath"`
	Hunks     []DiffHunk `json:"hunks"`
	Unified   string     `json:"unified"` // Traditional unified diff format
}

// CreateUnifiedDiff creates a unified diff from two file contents.
// Uses Myers diff algorithm via diffmatchpatch library.
func CreateUnifiedDiff(leftPath, rightPath, leftContent, rightContent string) string {
	result := ComputeDiff(leftPath, rightPath, leftContent, rightContent)
	return result.Unified
}

// ComputeDiff computes a diff between two file contents and returns structured results.
// Uses Myers diff algorithm via diffmatchpatch library for accurate line-based diffing.
func ComputeDiff(leftPath, rightPath, leftContent, rightContent string) *DiffResult {
	dmp := diffmatchpatch.New()

	// Compute line-based diff using the built-in method
	leftLines, rightLines, lineArray := dmp.DiffLinesToChars(leftContent, rightContent)
	diffs := dmp.DiffMain(leftLines, rightLines, false)
	diffs = dmp.DiffCharsToLines(diffs, lineArray)
	diffs = dmp.DiffCleanupSemantic(diffs)

	// Build unified diff string and hunks
	result := &DiffResult{
		LeftPath:  leftPath,
		RightPath: rightPath,
		Hunks:     make([]DiffHunk, 0),
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- %s\t%s\n", leftPath, time.Now().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("+++ %s\t%s\n", rightPath, time.Now().Format(time.RFC3339)))

	// Convert diffs to line-based hunks
	hunks := buildHunks(diffs)
	result.Hunks = hunks

	// Generate unified diff format
	for _, hunk := range hunks {
		sb.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
			hunk.OldStart, hunk.OldLines, hunk.NewStart, hunk.NewLines))
		for _, line := range hunk.Lines {
			switch line.Type {
			case "context":
				sb.WriteString(" " + line.Content)
			case "delete":
				sb.WriteString("-" + line.Content)
			case "add":
				sb.WriteString("+" + line.Content)
			}
		}
	}

	result.Unified = sb.String()
	return result
}

// buildHunks converts diffmatchpatch diffs into hunks with context lines.
func buildHunks(diffs []diffmatchpatch.Diff) []DiffHunk {
	const contextLines = 3 // Number of context lines before and after changes

	// First, build a flat list of all lines with their types
	var allLines []DiffLine
	oldLineNum := 1
	newLineNum := 1

	for _, diff := range diffs {
		lines := strings.Split(diff.Text, "\n")
		// Handle trailing newline - the last element after split might be empty
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
			// Re-add newlines to each line except the last one will be handled
			for i := range lines {
				lines[i] = lines[i] + "\n"
			}
		} else {
			// Add newline to all but last line
			for i := 0; i < len(lines)-1; i++ {
				lines[i] = lines[i] + "\n"
			}
			// Last line doesn't have newline
			if len(lines) > 0 {
				lines[len(lines)-1] = lines[len(lines)-1] + "\n"
			}
		}

		for _, line := range lines {
			var dl DiffLine
			switch diff.Type {
			case diffmatchpatch.DiffEqual:
				dl = DiffLine{Type: "context", Content: line, OldNum: oldLineNum, NewNum: newLineNum}
				oldLineNum++
				newLineNum++
			case diffmatchpatch.DiffDelete:
				dl = DiffLine{Type: "delete", Content: line, OldNum: oldLineNum}
				oldLineNum++
			case diffmatchpatch.DiffInsert:
				dl = DiffLine{Type: "add", Content: line, NewNum: newLineNum}
				newLineNum++
			}
			allLines = append(allLines, dl)
		}
	}

	// Now, group lines into hunks with context
	var hunks []DiffHunk
	var currentHunk *DiffHunk
	contextBuffer := make([]DiffLine, 0, contextLines)

	for i, line := range allLines {
		isChange := line.Type != "context"

		if isChange {
			// Start a new hunk or continue current one
			if currentHunk == nil {
				currentHunk = &DiffHunk{
					Lines: make([]DiffLine, 0),
				}
				// Add leading context from buffer
				startContext := contextBuffer
				if len(startContext) > contextLines {
					startContext = startContext[len(startContext)-contextLines:]
				}
				if len(startContext) > 0 {
					currentHunk.OldStart = startContext[0].OldNum
					currentHunk.NewStart = startContext[0].NewNum
					currentHunk.Lines = append(currentHunk.Lines, startContext...)
				} else {
					// No context, start at the change
					if line.OldNum > 0 {
						currentHunk.OldStart = line.OldNum
					} else {
						currentHunk.OldStart = 1
					}
					if line.NewNum > 0 {
						currentHunk.NewStart = line.NewNum
					} else {
						currentHunk.NewStart = 1
					}
				}
			}
			currentHunk.Lines = append(currentHunk.Lines, line)
			contextBuffer = contextBuffer[:0] // Clear context buffer
		} else {
			// Context line
			if currentHunk != nil {
				// Check if we should continue or end the hunk
				// Look ahead to see if there are more changes within context range
				hasMoreChanges := false
				for j := i + 1; j < len(allLines) && j <= i+contextLines; j++ {
					if allLines[j].Type != "context" {
						hasMoreChanges = true
						break
					}
				}

				if hasMoreChanges || len(contextBuffer) < contextLines {
					currentHunk.Lines = append(currentHunk.Lines, line)
					if !hasMoreChanges {
						contextBuffer = append(contextBuffer, line)
					}
				} else {
					// End the hunk - add trailing context
					// Calculate line counts
					calculateHunkCounts(currentHunk)
					hunks = append(hunks, *currentHunk)
					currentHunk = nil
					contextBuffer = []DiffLine{line}
				}
			} else {
				// No current hunk, just buffer context
				contextBuffer = append(contextBuffer, line)
				if len(contextBuffer) > contextLines {
					contextBuffer = contextBuffer[1:]
				}
			}
		}
	}

	// Finish any remaining hunk
	if currentHunk != nil {
		calculateHunkCounts(currentHunk)
		hunks = append(hunks, *currentHunk)
	}

	return hunks
}

// calculateHunkCounts calculates the old and new line counts for a hunk.
func calculateHunkCounts(hunk *DiffHunk) {
	oldCount := 0
	newCount := 0
	for _, line := range hunk.Lines {
		switch line.Type {
		case "context":
			oldCount++
			newCount++
		case "delete":
			oldCount++
		case "add":
			newCount++
		}
	}
	hunk.OldLines = oldCount
	hunk.NewLines = newCount
}

// CompareFiles reads two files and computes their diff.
func CompareFiles(leftPath, rightPath string) (*DiffResult, error) {
	leftContent, err := ReadFileContent(leftPath)
	if err != nil {
		return nil, fmt.Errorf("reading left file: %w", err)
	}
	rightContent, err := ReadFileContent(rightPath)
	if err != nil {
		return nil, fmt.Errorf("reading right file: %w", err)
	}
	return ComputeDiff(leftPath, rightPath, leftContent, rightContent), nil
}
