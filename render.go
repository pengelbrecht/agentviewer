// Package main provides file reading and content type detection.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// FileAccessConfig holds security configuration for file access.
type FileAccessConfig struct {
	// AllowedDirs is a list of directories that files can be read from.
	// If empty, all directories are allowed (default).
	// Paths must be absolute.
	AllowedDirs []string

	// LogAccess enables logging of file access attempts.
	LogAccess bool

	// Logger is the logger to use for file access logging.
	// If nil, uses the default log package.
	Logger *log.Logger
}

// fileAccessConfig is the global file access configuration.
var fileAccessConfig = &FileAccessConfig{
	AllowedDirs: nil,   // Allow all directories by default
	LogAccess:   false, // Logging disabled by default
	Logger:      nil,
}

var fileAccessConfigMu sync.RWMutex

// SetFileAccessConfig sets the global file access configuration.
func SetFileAccessConfig(config *FileAccessConfig) {
	fileAccessConfigMu.Lock()
	defer fileAccessConfigMu.Unlock()
	fileAccessConfig = config
}

// GetFileAccessConfig returns a copy of the current file access configuration.
func GetFileAccessConfig() *FileAccessConfig {
	fileAccessConfigMu.RLock()
	defer fileAccessConfigMu.RUnlock()
	return &FileAccessConfig{
		AllowedDirs: append([]string{}, fileAccessConfig.AllowedDirs...),
		LogAccess:   fileAccessConfig.LogAccess,
		Logger:      fileAccessConfig.Logger,
	}
}

// logFileAccess logs a file access attempt if logging is enabled.
func logFileAccess(path string, allowed bool, reason string) {
	fileAccessConfigMu.RLock()
	config := fileAccessConfig
	fileAccessConfigMu.RUnlock()

	if !config.LogAccess {
		return
	}

	status := "ALLOWED"
	if !allowed {
		status = "DENIED"
	}

	msg := fmt.Sprintf("[FILE ACCESS] %s: %s", status, path)
	if reason != "" {
		msg += fmt.Sprintf(" (%s)", reason)
	}

	if config.Logger != nil {
		config.Logger.Println(msg)
	} else {
		log.Println(msg)
	}
}

// ValidatePath checks if a path is valid and secure.
// Returns the cleaned absolute path and any error.
func ValidatePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Clean the path first
	cleanPath := filepath.Clean(path)

	// Convert to absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path: %w", err)
	}

	// Check against allowed directories if configured
	fileAccessConfigMu.RLock()
	allowedDirs := fileAccessConfig.AllowedDirs
	fileAccessConfigMu.RUnlock()

	if len(allowedDirs) > 0 {
		allowed := false
		for _, dir := range allowedDirs {
			// Clean and absolutize the allowed directory
			cleanDir, err := filepath.Abs(filepath.Clean(dir))
			if err != nil {
				continue
			}
			// Check if the file path is within the allowed directory
			if isSubPath(cleanDir, absPath) {
				allowed = true
				break
			}
		}
		if !allowed {
			logFileAccess(absPath, false, "path not in allowed directories")
			return "", fmt.Errorf("access denied: path not in allowed directories: %s", absPath)
		}
	}

	return absPath, nil
}

// isSubPath checks if filePath is within or equal to dirPath.
func isSubPath(dirPath, filePath string) bool {
	// Ensure dirPath ends with separator for proper prefix check
	if !strings.HasSuffix(dirPath, string(filepath.Separator)) {
		dirPath += string(filepath.Separator)
	}
	return strings.HasPrefix(filePath+string(filepath.Separator), dirPath) ||
		strings.TrimSuffix(filePath, string(filepath.Separator)) == strings.TrimSuffix(dirPath, string(filepath.Separator))
}

// ReadFileContent reads a file and returns its content.
// It validates that the path exists, is a regular file (not a directory),
// and is readable. It also enforces security restrictions from FileAccessConfig.
// Returns the content as a string.
func ReadFileContent(path string) (string, error) {
	// Validate and clean the path (includes security checks)
	cleanPath, err := ValidatePath(path)
	if err != nil {
		return "", err
	}

	// Get file info to validate the path
	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			logFileAccess(cleanPath, false, "file not found")
			return "", fmt.Errorf("file not found: %s", cleanPath)
		}
		if os.IsPermission(err) {
			logFileAccess(cleanPath, false, "permission denied")
			return "", fmt.Errorf("permission denied: %s", cleanPath)
		}
		logFileAccess(cleanPath, false, "cannot access")
		return "", fmt.Errorf("cannot access file: %w", err)
	}

	// Ensure it's a regular file, not a directory or other type
	if info.IsDir() {
		logFileAccess(cleanPath, false, "is a directory")
		return "", fmt.Errorf("path is a directory, not a file: %s", cleanPath)
	}
	if !info.Mode().IsRegular() {
		logFileAccess(cleanPath, false, "not a regular file")
		return "", fmt.Errorf("path is not a regular file: %s", cleanPath)
	}

	// Read the file content
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		if os.IsPermission(err) {
			logFileAccess(cleanPath, false, "permission denied reading")
			return "", fmt.Errorf("permission denied reading file: %s", cleanPath)
		}
		logFileAccess(cleanPath, false, "read failed")
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Log successful access
	logFileAccess(cleanPath, true, "")

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
		case ".mmd", ".mermaid":
			return TabTypeMermaid
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
		".go":     "go",
		".js":     "javascript",
		".jsx":    "javascript",
		".ts":     "typescript",
		".tsx":    "typescript",
		".py":     "python",
		".rb":     "ruby",
		".rs":     "rust",
		".java":   "java",
		".kt":     "kotlin",
		".swift":  "swift",
		".c":      "c",
		".cpp":    "cpp",
		".cc":     "cpp",
		".cxx":    "cpp",
		".h":      "c",
		".hpp":    "cpp",
		".cs":     "csharp",
		".php":    "php",
		".sh":     "bash",
		".bash":   "bash",
		".zsh":    "bash",
		".fish":   "fish",
		".ps1":    "powershell",
		".sql":    "sql",
		".html":   "html",
		".htm":    "html",
		".css":    "css",
		".scss":   "scss",
		".sass":   "sass",
		".less":   "less",
		".json":   "json",
		".yaml":   "yaml",
		".yml":    "yaml",
		".xml":    "xml",
		".toml":   "toml",
		".ini":    "ini",
		".cfg":    "ini",
		".conf":   "nginx",
		".lua":    "lua",
		".pl":     "perl",
		".r":      "r",
		".R":      "r",
		".m":      "matlab",
		".scala":  "scala",
		".ex":     "elixir",
		".exs":    "elixir",
		".erl":    "erlang",
		".hs":     "haskell",
		".clj":    "clojure",
		".elm":    "elm",
		".vue":    "vue",
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

// ParseUnifiedDiff parses a unified diff string (git diff output) into structured DiffResult.
// Supports standard unified diff format with --- and +++ headers and @@ hunk markers.
func ParseUnifiedDiff(diffText string) (*DiffResult, error) {
	result := &DiffResult{
		Hunks:   make([]DiffHunk, 0),
		Unified: diffText,
	}

	lines := strings.Split(diffText, "\n")
	if len(lines) == 0 {
		return result, nil
	}

	var currentHunk *DiffHunk
	oldLineNum := 0
	newLineNum := 0

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Parse file headers: --- a/path or --- path
		if strings.HasPrefix(line, "--- ") {
			path := parseFilePath(line[4:])
			result.LeftPath = path
			continue
		}

		// Parse file headers: +++ b/path or +++ path
		if strings.HasPrefix(line, "+++ ") {
			path := parseFilePath(line[4:])
			result.RightPath = path
			continue
		}

		// Parse hunk header: @@ -start,count +start,count @@ [optional context]
		if strings.HasPrefix(line, "@@ ") {
			if currentHunk != nil {
				result.Hunks = append(result.Hunks, *currentHunk)
			}

			hunk, err := parseHunkHeader(line)
			if err != nil {
				return nil, fmt.Errorf("parsing hunk header at line %d: %w", i+1, err)
			}
			currentHunk = hunk
			oldLineNum = hunk.OldStart
			newLineNum = hunk.NewStart
			continue
		}

		// Skip diff command line, index line, and other metadata
		if strings.HasPrefix(line, "diff ") ||
			strings.HasPrefix(line, "index ") ||
			strings.HasPrefix(line, "new file") ||
			strings.HasPrefix(line, "deleted file") ||
			strings.HasPrefix(line, "similarity") ||
			strings.HasPrefix(line, "rename ") ||
			strings.HasPrefix(line, "Binary ") {
			continue
		}

		// Parse diff lines within a hunk
		if currentHunk != nil {
			if len(line) == 0 {
				// Empty line in a diff could be a context line (space was trimmed)
				// or end of hunk. Treat as context if we're still expecting lines.
				expectedLines := currentHunk.OldLines + currentHunk.NewLines - len(currentHunk.Lines)
				if expectedLines > 0 {
					dl := DiffLine{
						Type:    "context",
						Content: "\n",
						OldNum:  oldLineNum,
						NewNum:  newLineNum,
					}
					currentHunk.Lines = append(currentHunk.Lines, dl)
					oldLineNum++
					newLineNum++
				}
				continue
			}

			prefix := line[0]
			content := ""
			if len(line) > 1 {
				content = line[1:]
			}
			// Ensure content has a newline for consistency
			if !strings.HasSuffix(content, "\n") {
				content = content + "\n"
			}

			switch prefix {
			case ' ':
				// Context line
				dl := DiffLine{
					Type:    "context",
					Content: content,
					OldNum:  oldLineNum,
					NewNum:  newLineNum,
				}
				currentHunk.Lines = append(currentHunk.Lines, dl)
				oldLineNum++
				newLineNum++
			case '-':
				// Deletion
				dl := DiffLine{
					Type:    "delete",
					Content: content,
					OldNum:  oldLineNum,
				}
				currentHunk.Lines = append(currentHunk.Lines, dl)
				oldLineNum++
			case '+':
				// Addition
				dl := DiffLine{
					Type:    "add",
					Content: content,
					NewNum:  newLineNum,
				}
				currentHunk.Lines = append(currentHunk.Lines, dl)
				newLineNum++
			case '\\':
				// "\ No newline at end of file" - skip but remove trailing newline from previous
				if len(currentHunk.Lines) > 0 {
					lastLine := &currentHunk.Lines[len(currentHunk.Lines)-1]
					lastLine.Content = strings.TrimSuffix(lastLine.Content, "\n")
				}
			}
		}
	}

	// Don't forget the last hunk
	if currentHunk != nil {
		result.Hunks = append(result.Hunks, *currentHunk)
	}

	return result, nil
}

// parseFilePath extracts the file path from a --- or +++ line.
// Handles formats like "a/path/to/file", "b/path/to/file", or just "path/to/file".
// Also handles timestamp suffixes.
func parseFilePath(s string) string {
	// Trim whitespace
	s = strings.TrimSpace(s)

	// Remove timestamp (e.g., "2024-01-15 10:30:00.000000000 +0000")
	if idx := strings.Index(s, "\t"); idx != -1 {
		s = s[:idx]
	}

	// Handle /dev/null
	if s == "/dev/null" {
		return s
	}

	// Strip a/ or b/ prefix if present (git diff format)
	if strings.HasPrefix(s, "a/") || strings.HasPrefix(s, "b/") {
		return s[2:]
	}

	return s
}

// parseHunkHeader parses @@ -old_start,old_count +new_start,new_count @@ context
func parseHunkHeader(line string) (*DiffHunk, error) {
	// Format: @@ -old_start,old_count +new_start,new_count @@ [context]
	// Or: @@ -old_start +new_start @@ (count of 1 is implicit)

	if !strings.HasPrefix(line, "@@ ") {
		return nil, fmt.Errorf("invalid hunk header: %s", line)
	}

	// Find the closing @@
	endIdx := strings.Index(line[3:], " @@")
	if endIdx == -1 {
		// Try without trailing space
		endIdx = strings.Index(line[3:], "@@")
		if endIdx == -1 {
			return nil, fmt.Errorf("missing closing @@ in hunk header: %s", line)
		}
	}

	rangeStr := line[3 : 3+endIdx]

	// Parse the range part: -old_start,old_count +new_start,new_count
	parts := strings.Fields(rangeStr)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid hunk range: %s", rangeStr)
	}

	oldPart := parts[0] // -start,count or -start
	newPart := parts[1] // +start,count or +start

	oldStart, oldCount, err := parseHunkRange(oldPart, "-")
	if err != nil {
		return nil, fmt.Errorf("invalid old range %s: %w", oldPart, err)
	}

	newStart, newCount, err := parseHunkRange(newPart, "+")
	if err != nil {
		return nil, fmt.Errorf("invalid new range %s: %w", newPart, err)
	}

	return &DiffHunk{
		OldStart: oldStart,
		OldLines: oldCount,
		NewStart: newStart,
		NewLines: newCount,
		Lines:    make([]DiffLine, 0),
	}, nil
}

// parseHunkRange parses a range like "-10,5" or "+20,3" or "-10" (count=1 implicit)
func parseHunkRange(s, prefix string) (start, count int, err error) {
	if !strings.HasPrefix(s, prefix) {
		return 0, 0, fmt.Errorf("expected prefix %s", prefix)
	}

	numStr := s[1:] // Remove prefix

	if idx := strings.Index(numStr, ","); idx != -1 {
		// Has count
		_, err = fmt.Sscanf(numStr, "%d,%d", &start, &count)
		if err != nil {
			return 0, 0, err
		}
	} else {
		// No count, implicit 1
		_, err = fmt.Sscanf(numStr, "%d", &start)
		if err != nil {
			return 0, 0, err
		}
		count = 1
	}

	// Handle edge case: count of 0 means empty (file created/deleted)
	if start == 0 && count == 0 {
		start = 1
	}

	return start, count, nil
}
