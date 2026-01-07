# Claude Plugin Marketplace Publishing Guide

Research findings for publishing agentviewer to Claude Code marketplaces.

## Summary

Claude Code has multiple distribution channels for plugins:

1. **Official Anthropic Marketplace** - Curated, requires PR submission
2. **GitHub Distribution** - Direct installation from any repo
3. **Community Registries** - Third-party directories
4. **Self-Hosted Marketplaces** - Organization-specific

## Publishing Options

### Option 1: Official Anthropic Marketplace (Recommended)

**Repository**: [anthropics/claude-plugins-official](https://github.com/anthropics/claude-plugins-official)

**Status**: Accepting third-party submissions via Pull Request

**How it works**:
- External plugins are placed in `/external_plugins/` directory
- Submit a PR adding your plugin
- Must meet quality and security standards
- Review process appears active (59 open PRs, 47 closed)

**Submission process** (inferred from existing PRs):
1. Fork `anthropics/claude-plugins-official`
2. Create `external_plugins/agentviewer/` directory
3. Add plugin structure:
   ```
   external_plugins/agentviewer/
   ├── .claude-plugin/
   │   └── plugin.json
   ├── skills/
   │   └── agentviewer/
   │       └── SKILL.md
   └── README.md
   ```
4. Submit PR with:
   - Clear description of plugin functionality
   - Security considerations
   - Installation requirements
5. Wait for Anthropic review

**Current plugins in official marketplace**: ~30+ including
- Language servers (TypeScript, Python, Go, Rust, etc.)
- MCP integrations (GitHub, GitLab, Slack, Supabase, etc.)
- Development tools (PR review, code review, commit commands)

**User installation** (once accepted):
```bash
/plugin install agentviewer@claude-plugins-official
```

### Option 2: GitHub Distribution (Self-Hosted)

**No approval needed** - Users install directly from our repo.

**How it works**:
- Users add our repository as a marketplace
- Plugin available immediately

**Setup** (already complete):
- `.claude-plugin/plugin.json` - Plugin manifest ✅
- `.claude-plugin/marketplace.json` - Marketplace definition ✅
- `skills/agentviewer/SKILL.md` - Agent skill ✅

**User installation**:
```bash
# Add marketplace
/plugin marketplace add peterengelbrecht/agentviewer

# Install plugin
/plugin install agentviewer@agentviewer-marketplace
```

Or directly:
```bash
/plugin add github:peterengelbrecht/agentviewer
```

### Option 3: Community Registry (claude-plugins.dev)

**Repository**: [Kamalnrf/claude-plugins](https://github.com/Kamalnrf/claude-plugins)

**Benefits**:
- Community-maintained directory
- CLI tool for simplified installation
- Cross-platform compatibility (Claude, Cursor, VS Code, etc.)

**Submission**: Submit PR to their registry repository

### Option 4: npm Publishing

**Status**: Not yet fully implemented in Claude Code

Per official docs: "npm sources are not yet fully implemented"

**Future option when available**:
```bash
npm publish claude-plugin-agentviewer
/plugin add claude-plugin-agentviewer
```

## Plugin Requirements

### Required Files

**plugin.json** (minimum):
```json
{
  "name": "agentviewer",
  "description": "Display rich content in browser viewer for AI agents",
  "version": "0.1.0"
}
```

**Complete plugin.json** (recommended):
```json
{
  "name": "agentviewer",
  "description": "Display rich content (markdown, code, diffs, diagrams) in a browser-based tabbed viewer for AI agents",
  "version": "0.1.0",
  "author": {
    "name": "Peter Engelbrecht",
    "email": "email@example.com"
  },
  "repository": "https://github.com/peterengelbrecht/agentviewer",
  "license": "MIT",
  "homepage": "https://github.com/peterengelbrecht/agentviewer",
  "keywords": ["viewer", "markdown", "code", "diff", "mermaid", "agent-tools"],
  "category": "productivity"
}
```

### Quality Standards (for official marketplace)

Based on existing plugins:

1. **Documentation**
   - Clear README with installation instructions
   - Usage examples
   - API reference if applicable

2. **Security**
   - No arbitrary code execution without user consent
   - Clear trust boundaries documented
   - Dependencies vetted

3. **Functionality**
   - Plugin must work as described
   - Error handling present
   - Graceful degradation when server unavailable

4. **Best Practices**
   - Use `${CLAUDE_PLUGIN_ROOT}` for internal paths
   - Follow semantic versioning
   - Minimize dependencies

### Reserved Names (cannot use)

- `claude-code-marketplace`
- `claude-code-plugins`
- `claude-plugins-official`
- `anthropic-marketplace`
- `anthropic-plugins`
- `agent-skills`

## Recommended Strategy

### Phase 1: GitHub Distribution (Now)

Agentviewer already supports direct GitHub installation:
```bash
/plugin add github:peterengelbrecht/agentviewer
```

This works today with no approval needed.

### Phase 2: Official Marketplace Submission

1. Ensure plugin meets quality standards
2. Write comprehensive submission PR description
3. Include:
   - Plugin description and use cases
   - Security model (localhost only, no network access)
   - Installation requirements (Go binary)
   - Screenshots or demo if possible
4. Submit PR to `anthropics/claude-plugins-official`
5. Respond to reviewer feedback

### Phase 3: Community Visibility

- Submit to community registries (claude-plugins.dev)
- Add to awesome-claude-plugins list
- Document in README

## Marketplace Entry Example

For official marketplace submission, entry in `marketplace.json`:

```json
{
  "name": "agentviewer",
  "description": "Display rich content (markdown, code, diffs, diagrams) in a browser viewer for AI agents",
  "version": "0.1.0",
  "author": {
    "name": "Peter Engelbrecht"
  },
  "source": "./external_plugins/agentviewer",
  "category": "productivity",
  "homepage": "https://github.com/peterengelbrecht/agentviewer",
  "tags": ["viewer", "markdown", "code", "diff", "mermaid", "rich-content"]
}
```

## Next Steps

1. ✅ Plugin structure complete
2. ✅ GitHub distribution ready
3. [ ] Prepare official marketplace submission materials
4. [ ] Create PR to anthropics/claude-plugins-official
5. [ ] Submit to community registries
6. [ ] Monitor for npm support availability

## References

- [Claude Code Plugin Documentation](https://code.claude.com/docs/en/plugins)
- [Official Marketplace Repository](https://github.com/anthropics/claude-plugins-official)
- [Plugin Marketplace Documentation](https://code.claude.com/docs/en/plugin-marketplaces)
- [Community Registry](https://claude-plugins.dev/)
