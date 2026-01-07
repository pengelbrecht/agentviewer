//go:build tools

// Package main includes tool dependencies for development and testing.
package main

import (
	// chromedp for e2e browser testing
	_ "github.com/chromedp/chromedp"
)
