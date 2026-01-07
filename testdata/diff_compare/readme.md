# Two-File Comparison Test Data

This directory contains pairs of files for testing the agentviewer's two-file diff comparison feature.

## File Pairs

Each pair consists of an "old" (v1) and "new" (v2) version:

| Old File | New File | Description |
|----------|----------|-------------|
| simple_v1.txt | simple_v2.txt | Basic text changes (add/modify/delete lines) |
| config_v1.json | config_v2.json | JSON configuration changes (nested objects) |
| code_v1.go | code_v2.go | Go code refactoring (simple to database pattern) |
| code_v1.py | code_v2.py | Python code changes (class to dataclass, repository pattern) |
| html_v1.html | html_v2.html | HTML template changes (basic to modern with CSS) |
| empty_v1.txt | empty_v2.txt | Empty to non-empty file |
| whitespace_v1.txt | whitespace_v2.txt | Whitespace-only changes (tabs to spaces) |
| unicode_v1.txt | unicode_v2.txt | Unicode content (extended characters, emoji) |
| large_v1.txt | large_v2.txt | Large file (100+ lines, multiple sections) |
| sql_v1.sql | sql_v2.sql | SQL schema changes (basic to CTE/window functions) |
| types_v1.ts | types_v2.ts | TypeScript type evolution (sync to async, generics) |
| docker_v1.yaml | docker_v2.yaml | Docker Compose (basic to production-ready)

## Usage with agentviewer API

```bash
curl -X POST http://localhost:8080/api/tabs \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Code Review",
    "type": "diff",
    "diff": {
      "left": "testdata/diff_compare/code_v1.go",
      "right": "testdata/diff_compare/code_v2.go",
      "leftLabel": "Before",
      "rightLabel": "After",
      "language": "go"
    }
  }'
```
