# GitHub Flavored Markdown Test File (HOT RELOADED!)

This file demonstrates various GFM features for testing agentviewer's markdown rendering.

## Task Lists

### Project Checklist
- [x] Initialize project structure
- [x] Set up dependency management
- [ ] Write unit tests
- [ ] Complete documentation
- [ ] Deploy to production

### Nested Task Lists
- [ ] Backend Development
  - [x] Set up database schema
  - [x] Create API endpoints
  - [ ] Add authentication
    - [x] JWT token generation
    - [ ] OAuth integration
- [ ] Frontend Development
  - [ ] Design UI components
  - [ ] Implement state management

## Tables

### Basic Table

| Name | Age | Role |
|------|-----|------|
| Alice | 28 | Developer |
| Bob | 35 | Designer |
| Charlie | 42 | Manager |

### Aligned Columns

| Left Aligned | Center Aligned | Right Aligned |
|:-------------|:--------------:|--------------:|
| L1 | C1 | R1 |
| Left | Center | Right |
| Data | More Data | 12345 |

### Complex Table with Code

| Language | Example | Description |
|----------|---------|-------------|
| Go | `fmt.Println()` | Print to stdout |
| Python | `print()` | Print function |
| JavaScript | `console.log()` | Console output |

## Autolinks

Visit https://github.com for more info.

Email: user@example.com

## Strikethrough

~~This text is crossed out~~ but this is not.

~~Multiword strikethrough works too~~

## Extended Autolinks

- Website: www.example.com
- Email: support@example.org
- Issue reference: #42

## Code Blocks

### Fenced Code with Syntax Highlighting

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

```python
def hello():
    print("Hello from Python")

if __name__ == "__main__":
    hello()
```

```javascript
const greet = (name) => {
    console.log(`Hello, ${name}!`);
};

greet("World");
```

### Inline Code

Use `go build` to compile and `go test` to run tests.

## Blockquotes

> This is a blockquote.
> It can span multiple lines.
>
> > Nested blockquotes are supported too.

## Horizontal Rules

---

***

___

## Lists

### Unordered List
- Item one
- Item two
  - Nested item
  - Another nested
    - Deeply nested
- Item three

### Ordered List
1. First item
2. Second item
   1. Sub-item 2.1
   2. Sub-item 2.2
3. Third item

## Links and Images

### Links
[GitHub](https://github.com)
[Relative link](./sample.go)
[Reference link][ref1]

[ref1]: https://example.com "Example Site"

### Images (reference)
![Alt text](https://via.placeholder.com/150 "Placeholder")

## Emphasis

*Italic text* and _also italic_

**Bold text** and __also bold__

***Bold and italic*** and ___also bold italic___

## Footnotes

Here is a sentence with a footnote[^1].

[^1]: This is the footnote content.

## Definition Lists (Extended)

Term 1
: Definition for term 1

Term 2
: Definition for term 2
: Another definition for term 2

## Escaping

\*Not italic\*
\`Not code\`
\[Not a link\]
