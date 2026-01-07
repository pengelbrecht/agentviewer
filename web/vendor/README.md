# Vendor Libraries

This directory contains embedded JavaScript libraries for the frontend.

## Included Libraries

| Library | Version | Purpose |
|---------|---------|---------|
| marked.js | 15.0.5 | Markdown rendering |
| highlight.js | 11.10.0 | Syntax highlighting |
| mermaid.js | 11.4.1 | Diagram rendering |
| KaTeX | 0.16.18 | Math rendering |
| diff2html | 3.4.51 | Diff rendering |
| diff2html-ui | 3.4.51 | Diff rendering with syntax highlighting (slim bundle) |

## Files

- `marked.min.js` - Markdown parser
- `highlight.min.js` - Syntax highlighter core
- `highlight-github-dark.min.css` - Dark theme for syntax highlighting
- `mermaid.min.js` - Diagram renderer
- `katex.min.js` - Math renderer
- `katex.min.css` - KaTeX styles
- `katex-fonts/` - KaTeX font files (woff2)
- `diff2html.min.js` - Diff renderer (basic)
- `diff2html-ui-slim.min.js` - Diff renderer with highlight.js integration (slim)
- `diff2html.min.css` - Diff styles

## Usage in HTML

```html
<!-- CSS -->
<link rel="stylesheet" href="/vendor/highlight-github-dark.min.css">
<link rel="stylesheet" href="/vendor/katex.min.css">
<link rel="stylesheet" href="/vendor/diff2html.min.css">

<!-- JS -->
<script src="/vendor/marked.min.js"></script>
<script src="/vendor/highlight.min.js"></script>
<script src="/vendor/mermaid.min.js"></script>
<script src="/vendor/katex.min.js"></script>
<script src="/vendor/diff2html-ui-slim.min.js"></script>
```
