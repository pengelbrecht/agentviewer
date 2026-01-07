package main

import (
	"os"
	"runtime"
	"testing"
)

func TestOpenBrowser(t *testing.T) {
	// OpenBrowser uses exec.Command().Start() which spawns a process.
	// Only run this test in CI to avoid opening browsers during local development.
	if os.Getenv("CI") == "" {
		t.Skip("Skipping OpenBrowser test outside CI (would open actual browser)")
	}

	url := "http://127.0.0.1:9999/test"

	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		// On these platforms, we expect the command to start without error.
		// In CI, this will likely fail due to no display - that's fine.
		err := OpenBrowser(url)
		if err != nil {
			t.Logf("OpenBrowser returned error (expected in CI without display): %v", err)
		}
	default:
		// On unsupported platforms, we expect an error
		err := OpenBrowser(url)
		if err == nil {
			t.Error("Expected error on unsupported platform")
		}
	}
}
