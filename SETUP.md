# Mac Go Development Setup

Minimal setup for CLI/LLM-driven Go development on macOS.

## Install Go

```bash
brew install go
```

## Configure Environment

Add to your shell profile (`~/.zshrc` or `~/.bashrc`):

```bash
export GOPATH="$HOME/go"
export GOBIN="$GOPATH/bin"
export PATH="$GOBIN:$PATH"
```

Then reload:

```bash
source ~/.zshrc  # or ~/.bashrc
```

## Verify Installation

```bash
go version
go env GOPATH GOBIN GOROOT
```

Expected output shows Go version 1.22+ and correctly set paths.

## Quick Test

```bash
# In project directory
go mod tidy
go build .
```

If these complete without errors, your setup is ready.
