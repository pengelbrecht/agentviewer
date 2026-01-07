package main

import (
	"testing"
)

func TestOpenBrowser(t *testing.T) {
	// OpenBrowser spawns a real browser process via exec.Command.
	// Cannot be meaningfully tested without side effects (opening browser)
	// or mocking exec.Command. The function is simple enough to verify by inspection.
	t.Skip("OpenBrowser requires spawning real browser - not testable in isolation")
}
