# Comprehensive Table Testing

This file focuses specifically on table rendering in various configurations.

## Simple Table

| A | B | C |
|---|---|---|
| 1 | 2 | 3 |

## Column Alignment

| Default | Left | Center | Right |
|---------|:-----|:------:|------:|
| d1 | l1 | c1 | r1 |
| default | left | center | right |
| text | aligned | middle | 1234 |

## Table with Empty Cells

| Header 1 | Header 2 | Header 3 |
|----------|----------|----------|
| Data |  | Data |
|  | Data |  |
| Data | Data | Data |

## Table with Long Content

| Short | Medium Length | Very Long Content Column That Tests Wrapping |
|-------|---------------|---------------------------------------------|
| A | Some text | This is a much longer piece of text that might need to wrap in the table cell |
| B | More text | Another long entry with detailed information about something |
| C | Text | Short |

## Table with Code and Formatting

| Syntax | Example | Output |
|--------|---------|--------|
| **Bold** | `**text**` | **text** |
| *Italic* | `*text*` | *text* |
| ~~Strike~~ | `~~text~~` | ~~text~~ |
| `Code` | `` `code` `` | `code` |
| [Link](#) | `[text](#)` | [text](#) |

## Table with Numbers

| ID | Value | Percentage | Currency |
|---:|------:|-----------:|---------:|
| 1 | 100 | 10.5% | $99.99 |
| 2 | 250 | 25.0% | $249.00 |
| 3 | 1000 | 100.0% | $999.99 |
| 4 | 42 | 4.2% | $41.58 |

## Wide Table

| Col1 | Col2 | Col3 | Col4 | Col5 | Col6 | Col7 | Col8 | Col9 | Col10 |
|------|------|------|------|------|------|------|------|------|-------|
| A1 | A2 | A3 | A4 | A5 | A6 | A7 | A8 | A9 | A10 |
| B1 | B2 | B3 | B4 | B5 | B6 | B7 | B8 | B9 | B10 |
| C1 | C2 | C3 | C4 | C5 | C6 | C7 | C8 | C9 | C10 |

## Table with Special Characters

| Symbol | Name | Unicode |
|--------|------|---------|
| & | Ampersand | U+0026 |
| < | Less than | U+003C |
| > | Greater than | U+003E |
| " | Quote | U+0022 |
| ' | Apostrophe | U+0027 |
| \| | Pipe | U+007C |

## Table with Links

| Service | Link | Status |
|---------|------|--------|
| GitHub | [github.com](https://github.com) | Active |
| GitLab | [gitlab.com](https://gitlab.com) | Active |
| Bitbucket | [bitbucket.org](https://bitbucket.org) | Active |

## Minimal Tables

| Single |
|--------|
| Cell |

---

| Two | Columns |
|-----|---------|
| A | B |

## Table without Header Separator Style

| Name | Value |
|-|-|
| Min | Separator |

## Escaped Pipe in Table

| Expression | Result |
|------------|--------|
| a \| b | OR operation |
| x \| y \| z | Multiple OR |
