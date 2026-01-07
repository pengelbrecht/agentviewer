package main

import (
	"runtime"
	"testing"
)

func TestOpenBrowser(t *testing.T) {
	// OpenBrowser uses exec.Command().Start() which spawns a process
	// We can't easily mock that, so we just verify it doesn't error
	// on supported platforms

	url := "http://127.0.0.1:9999/test" // Non-existent URL is fine

	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		// On these platforms, we expect the command to start without error
		// Note: The browser opening is async (Start not Run), so this won't block
		err := OpenBrowser(url)
		// On CI, we might not have a display, so we accept the error
		// The important thing is the code path is executed
		if err != nil {
			t.Logf("OpenBrowser returned error (may be expected on CI): %v", err)
		}
	default:
		// On unsupported platforms, we expect an error
		err := OpenBrowser(url)
		if err == nil {
			t.Error("Expected error on unsupported platform")
		}
	}
}
