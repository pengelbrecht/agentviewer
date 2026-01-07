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

- agentviewer CLI installed and in PATH
- Run `agentviewer serve --open &` to start the server
