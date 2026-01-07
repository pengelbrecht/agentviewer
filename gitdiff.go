// Package main provides git diff execution functionality.
package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// DiffMode specifies how to compute a git diff.
type DiffMode struct {
	// Type is the diff mode: "unstaged", "staged", "head", "commit", or "range"
	Type string
	// Ref is the commit SHA for "commit" mode or range like "main..feature" for "range" mode
	Ref string
}

// ParseDiffMode parses a diffMode string into a DiffMode struct.
// Supported formats:
//   - "unstaged" (default): working directory vs index
//   - "staged": index vs HEAD
//   - "head": working directory vs HEAD
//   - "commit:<sha>": changes introduced by a specific commit
//   - "range:<from>..<to>": diff between two refs
func ParseDiffMode(s string) (DiffMode, error) {
	if s == "" {
		return DiffMode{Type: "unstaged"}, nil
	}

	s = strings.TrimSpace(s)

	// Simple modes
	switch s {
	case "unstaged":
		return DiffMode{Type: "unstaged"}, nil
	case "staged":
		return DiffMode{Type: "staged"}, nil
	case "head":
		return DiffMode{Type: "head"}, nil
	}

	// Commit mode: commit:<sha>
	if strings.HasPrefix(s, "commit:") {
		ref := strings.TrimPrefix(s, "commit:")
		ref = strings.TrimSpace(ref)
		if ref == "" {
			return DiffMode{}, fmt.Errorf("commit mode requires a SHA: commit:<sha>")
		}
		return DiffMode{Type: "commit", Ref: ref}, nil
	}

	// Range mode: range:<from>..<to>
	if strings.HasPrefix(s, "range:") {
		ref := strings.TrimPrefix(s, "range:")
		ref = strings.TrimSpace(ref)
		if ref == "" || !strings.Contains(ref, "..") {
			return DiffMode{}, fmt.Errorf("range mode requires format: range:<from>..<to>")
		}
		return DiffMode{Type: "range", Ref: ref}, nil
	}

	return DiffMode{}, fmt.Errorf("invalid diffMode: %q (valid: unstaged, staged, head, commit:<sha>, range:<from>..<to>)", s)
}

// GitDiff computes a git diff for the given file path using the specified mode.
// Returns the unified diff output as a string.
func GitDiff(path string, mode DiffMode) (string, error) {
	// Validate and clean the path
	cleanPath, err := ValidatePath(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Find the git repository root
	repoRoot, err := findGitRoot(cleanPath)
	if err != nil {
		return "", err
	}

	// Make the path relative to the repo root for git
	relPath, err := filepath.Rel(repoRoot, cleanPath)
	if err != nil {
		return "", fmt.Errorf("cannot compute relative path: %w", err)
	}

	// Build the git diff command based on mode
	args, err := buildGitDiffArgs(mode, relPath)
	if err != nil {
		return "", err
	}

	// Execute git diff
	cmd := exec.Command("git", args...)
	cmd.Dir = repoRoot

	output, err := cmd.Output()
	if err != nil {
		// Check for specific error cases
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "not a git repository") {
				return "", fmt.Errorf("not a git repository: %s", repoRoot)
			}
			if strings.Contains(stderr, "unknown revision") || strings.Contains(stderr, "bad revision") {
				return "", fmt.Errorf("invalid git reference: %s", mode.Ref)
			}
			if strings.Contains(stderr, "did not match any") {
				return "", fmt.Errorf("file not tracked by git: %s", relPath)
			}
			// For other non-zero exits, it might just mean no diff
			// Git diff exits with 1 when there are differences and --exit-code is used
			// but we're not using --exit-code, so actual errors have stderr
			if len(stderr) > 0 {
				return "", fmt.Errorf("git diff failed: %s", strings.TrimSpace(stderr))
			}
		}
		// If there's no stderr, it might be fine (empty diff)
	}

	return string(output), nil
}

// findGitRoot finds the root of the git repository containing the given path.
func findGitRoot(path string) (string, error) {
	// Get the directory containing the file
	dir := filepath.Dir(path)

	// Run git rev-parse --show-toplevel to find the repo root
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "not a git repository") {
				return "", fmt.Errorf("not a git repository: %s", dir)
			}
			return "", fmt.Errorf("git error: %s", strings.TrimSpace(stderr))
		}
		return "", fmt.Errorf("failed to find git repository: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// buildGitDiffArgs builds the git diff command arguments based on the mode.
func buildGitDiffArgs(mode DiffMode, relPath string) ([]string, error) {
	// Always include these flags for consistent unified diff output
	baseArgs := []string{"diff", "--no-color"}

	switch mode.Type {
	case "unstaged":
		// Working directory vs index (staged changes)
		// git diff -- <path>
		return append(baseArgs, "--", relPath), nil

	case "staged":
		// Index vs HEAD (what's staged for commit)
		// git diff --cached -- <path>
		return append(baseArgs, "--cached", "--", relPath), nil

	case "head":
		// Working directory vs HEAD (all uncommitted changes)
		// git diff HEAD -- <path>
		return append(baseArgs, "HEAD", "--", relPath), nil

	case "commit":
		// Changes introduced by a specific commit
		// git diff <sha>^..<sha> -- <path>
		// Or for root commit: git diff --root <sha> -- <path>
		// Using show for single commit is cleaner: git show <sha> -- <path>
		// But diff format: git diff <sha>~1 <sha> -- <path>
		// Safest: git diff <sha>^ <sha> -- <path>
		return append(baseArgs, mode.Ref+"^", mode.Ref, "--", relPath), nil

	case "range":
		// Diff between two refs
		// git diff <from>..<to> -- <path>
		return append(baseArgs, mode.Ref, "--", relPath), nil

	default:
		return nil, fmt.Errorf("unknown diff mode: %s", mode.Type)
	}
}

// IsGitRepo checks if the given path is within a git repository.
func IsGitRepo(path string) bool {
	dir := path
	// If it's a file, use its directory
	info, err := exec.Command("test", "-d", path).Output()
	_ = info
	if err != nil {
		dir = filepath.Dir(path)
	}

	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	err = cmd.Run()
	return err == nil
}

// IsTrackedByGit checks if the given file is tracked by git.
func IsTrackedByGit(path string) (bool, error) {
	cleanPath, err := ValidatePath(path)
	if err != nil {
		return false, err
	}

	repoRoot, err := findGitRoot(cleanPath)
	if err != nil {
		return false, err
	}

	relPath, err := filepath.Rel(repoRoot, cleanPath)
	if err != nil {
		return false, err
	}

	// git ls-files --error-unmatch <path>
	// Returns 0 if tracked, non-zero if not
	cmd := exec.Command("git", "ls-files", "--error-unmatch", relPath)
	cmd.Dir = repoRoot
	err = cmd.Run()
	return err == nil, nil
}
