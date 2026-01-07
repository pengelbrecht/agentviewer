package main

import (
	"strings"
	"testing"
)

func TestParseDiffMode(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantType  string
		wantRef   string
		wantError bool
		errSubstr string
	}{
		// Basic modes
		{name: "empty defaults to unstaged", input: "", wantType: "unstaged"},
		{name: "unstaged", input: "unstaged", wantType: "unstaged"},
		{name: "staged", input: "staged", wantType: "staged"},
		{name: "head", input: "head", wantType: "head"},
		{name: "whitespace trimmed", input: "  staged  ", wantType: "staged"},

		// Commit mode
		{name: "commit with sha", input: "commit:abc123", wantType: "commit", wantRef: "abc123"},
		{name: "commit with full sha", input: "commit:abc123def456", wantType: "commit", wantRef: "abc123def456"},
		{name: "commit with whitespace", input: "commit:  abc123  ", wantType: "commit", wantRef: "abc123"},
		{name: "commit empty sha", input: "commit:", wantError: true, errSubstr: "requires a SHA"},
		{name: "commit only whitespace", input: "commit:   ", wantError: true, errSubstr: "requires a SHA"},

		// Range mode
		{name: "range mode", input: "range:main..feature", wantType: "range", wantRef: "main..feature"},
		{name: "range three dots", input: "range:main...feature", wantType: "range", wantRef: "main...feature"},
		{name: "range empty", input: "range:", wantError: true, errSubstr: "range mode requires format"},
		{name: "range no dots", input: "range:main", wantError: true, errSubstr: "range mode requires format"},

		// Invalid modes
		{name: "unknown mode", input: "foobar", wantError: true, errSubstr: "invalid diffMode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDiffMode(tt.input)

			if tt.wantError {
				if err == nil {
					t.Errorf("ParseDiffMode(%q) expected error containing %q, got nil", tt.input, tt.errSubstr)
					return
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("ParseDiffMode(%q) error = %q, want substring %q", tt.input, err.Error(), tt.errSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseDiffMode(%q) unexpected error: %v", tt.input, err)
				return
			}

			if got.Type != tt.wantType {
				t.Errorf("ParseDiffMode(%q).Type = %q, want %q", tt.input, got.Type, tt.wantType)
			}
			if got.Ref != tt.wantRef {
				t.Errorf("ParseDiffMode(%q).Ref = %q, want %q", tt.input, got.Ref, tt.wantRef)
			}
		})
	}
}

func TestBuildGitDiffArgs(t *testing.T) {
	tests := []struct {
		name    string
		mode    DiffMode
		relPath string
		want    []string
		wantErr bool
	}{
		{
			name:    "unstaged mode",
			mode:    DiffMode{Type: "unstaged"},
			relPath: "file.go",
			want:    []string{"diff", "--no-color", "--", "file.go"},
		},
		{
			name:    "staged mode",
			mode:    DiffMode{Type: "staged"},
			relPath: "file.go",
			want:    []string{"diff", "--no-color", "--cached", "--", "file.go"},
		},
		{
			name:    "head mode",
			mode:    DiffMode{Type: "head"},
			relPath: "file.go",
			want:    []string{"diff", "--no-color", "HEAD", "--", "file.go"},
		},
		{
			name:    "commit mode",
			mode:    DiffMode{Type: "commit", Ref: "abc123"},
			relPath: "file.go",
			want:    []string{"diff", "--no-color", "abc123^", "abc123", "--", "file.go"},
		},
		{
			name:    "range mode",
			mode:    DiffMode{Type: "range", Ref: "main..feature"},
			relPath: "file.go",
			want:    []string{"diff", "--no-color", "main..feature", "--", "file.go"},
		},
		{
			name:    "unknown mode",
			mode:    DiffMode{Type: "invalid"},
			relPath: "file.go",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildGitDiffArgs(tt.mode, tt.relPath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("buildGitDiffArgs() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("buildGitDiffArgs() unexpected error: %v", err)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("buildGitDiffArgs() = %v, want %v", got, tt.want)
				return
			}

			for i, arg := range got {
				if arg != tt.want[i] {
					t.Errorf("buildGitDiffArgs()[%d] = %q, want %q", i, arg, tt.want[i])
				}
			}
		})
	}
}
