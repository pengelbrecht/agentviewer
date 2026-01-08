# Agentviewer Claude Code Plugin

This plugin provides native integration of agentviewer with Claude Code.

## Local Installation

To install the plugin from a local clone:

```bash
# Clone the repository
git clone https://github.com/pengelbrecht/agentviewer.git
cd agentviewer

# Install the plugin locally
claude plugins add --local .
```

## Marketplace Installation

When published to a Claude Code marketplace:

```bash
claude plugins add agentviewer@agentviewer-marketplace
```

## What This Plugin Provides

- **agentviewer skill**: Display rich content (markdown, code, diffs, diagrams) in a browser viewer
- **allowed-tools**: `Read`, `Bash(curl:*)`, `Bash(agentviewer:*)`

## Usage

Once installed, Claude will automatically use the agentviewer skill when you need to display:

- Markdown documents with rendered formatting
- Code with syntax highlighting
- Git diffs with side-by-side comparison
- Mermaid diagrams and LaTeX math

The skill provides API documentation and best practices for interacting with the agentviewer server.

## Requirements

The agentviewer CLI must be installed and available in your PATH.

### Installation Options

**macOS/Linux (Homebrew):**
```bash
brew tap pengelbrecht/agentviewer
brew install agentviewer
```

**Windows (Scoop):**
```powershell
scoop bucket add agentviewer https://github.com/pengelbrecht/scoop-agentviewer
scoop install agentviewer
```

**Windows (Winget):**
```powershell
winget install pengelbrecht.agentviewer
```

**Linux (deb - Debian/Ubuntu):**
```bash
curl -LO https://github.com/pengelbrecht/agentviewer/releases/latest/download/agentviewer_amd64.deb
sudo dpkg -i agentviewer_amd64.deb
```

**Linux (rpm - Fedora/RHEL):**
```bash
curl -LO https://github.com/pengelbrecht/agentviewer/releases/latest/download/agentviewer.amd64.rpm
sudo rpm -i agentviewer.amd64.rpm
```

**Go:**
```bash
go install github.com/pengelbrecht/agentviewer@latest
```

**Binary:** Download from [GitHub Releases](https://github.com/pengelbrecht/agentviewer/releases)

### Verify Installation

```bash
agentviewer --version
```

### Start the Server

```bash
agentviewer serve --open &
```
