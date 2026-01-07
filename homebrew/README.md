# Homebrew Formula for agentviewer

This directory contains the Homebrew formula template for agentviewer.

## Setting Up the Homebrew Tap

1. **Create a new repository** named `homebrew-agentviewer` at:
   https://github.com/pengelbrecht/homebrew-agentviewer

2. **Clone the repository** and copy the formula:
   ```bash
   git clone https://github.com/pengelbrecht/homebrew-agentviewer
   cd homebrew-agentviewer
   mkdir Formula
   cp /path/to/agentviewer/homebrew/Formula/agentviewer.rb Formula/
   ```

3. **Update SHA256 checksums** after a release:
   ```bash
   # Download binaries and compute SHA256
   curl -sL https://github.com/pengelbrecht/agentviewer/releases/download/v0.1.0/agentviewer-darwin-arm64 | shasum -a 256
   curl -sL https://github.com/pengelbrecht/agentviewer/releases/download/v0.1.0/agentviewer-darwin-amd64 | shasum -a 256
   curl -sL https://github.com/pengelbrecht/agentviewer/releases/download/v0.1.0/agentviewer-linux-arm64 | shasum -a 256
   curl -sL https://github.com/pengelbrecht/agentviewer/releases/download/v0.1.0/agentviewer-linux-amd64 | shasum -a 256
   ```

4. **Commit and push** the formula:
   ```bash
   git add Formula/agentviewer.rb
   git commit -m "Add agentviewer formula"
   git push
   ```

## Installing via Homebrew

Once the tap repository is set up:

```bash
brew tap pengelbrecht/agentviewer
brew install agentviewer
```

Or in one command:

```bash
brew install pengelbrecht/agentviewer/agentviewer
```

## Updating the Formula

After each release:

1. Update the `version` in the formula
2. Recalculate SHA256 checksums for all binaries
3. Replace the placeholder values in the formula
4. Commit and push to the tap repository

### Automation Script

A helper script to update the formula:

```bash
#!/bin/bash
VERSION="${1:-0.1.0}"
BASE_URL="https://github.com/pengelbrecht/agentviewer/releases/download/v${VERSION}"

echo "Fetching checksums for v${VERSION}..."

DARWIN_ARM64=$(curl -sL "${BASE_URL}/agentviewer-darwin-arm64" | shasum -a 256 | cut -d' ' -f1)
DARWIN_AMD64=$(curl -sL "${BASE_URL}/agentviewer-darwin-amd64" | shasum -a 256 | cut -d' ' -f1)
LINUX_ARM64=$(curl -sL "${BASE_URL}/agentviewer-linux-arm64" | shasum -a 256 | cut -d' ' -f1)
LINUX_AMD64=$(curl -sL "${BASE_URL}/agentviewer-linux-amd64" | shasum -a 256 | cut -d' ' -f1)

echo "darwin-arm64: ${DARWIN_ARM64}"
echo "darwin-amd64: ${DARWIN_AMD64}"
echo "linux-arm64:  ${LINUX_ARM64}"
echo "linux-amd64:  ${LINUX_AMD64}"
```

## Testing the Formula Locally

```bash
# Install from local file
brew install --build-from-source ./Formula/agentviewer.rb

# Run the test block
brew test agentviewer

# Audit the formula
brew audit --strict --online agentviewer
```
