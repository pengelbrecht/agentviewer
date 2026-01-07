# Advanced Task Lists Testing

This file tests various task list scenarios for GFM rendering.

## Basic Task List

- [ ] Unchecked item
- [x] Checked item
- [ ] Another unchecked

## Mixed Content Task List

- [x] **Bold task** completed
- [ ] *Italic task* pending
- [x] Task with `inline code`
- [ ] Task with [link](https://example.com)

## Nested Task Lists

- [ ] Main task 1
  - [ ] Sub-task 1.1
  - [x] Sub-task 1.2 (done)
  - [ ] Sub-task 1.3
    - [x] Sub-sub-task 1.3.1
    - [ ] Sub-sub-task 1.3.2
- [x] Main task 2 (all subtasks done)
  - [x] Sub-task 2.1
  - [x] Sub-task 2.2

## Task List with Descriptions

- [ ] **Setup development environment**
  Configure local development tools and dependencies.

- [x] **Initialize repository**
  Create Git repository and initial commit.

- [ ] **Write documentation**
  Create README and API documentation.

## Sprint Planning Task List

### Sprint 1 - Foundation
- [x] Project scaffolding
- [x] CI/CD pipeline setup
- [x] Database schema design
- [ ] Core API implementation

### Sprint 2 - Features
- [ ] User authentication
  - [ ] Login endpoint
  - [ ] Registration endpoint
  - [ ] Password reset
- [ ] Profile management
- [ ] Settings page

### Sprint 3 - Polish
- [ ] Performance optimization
- [ ] Security audit
- [ ] Documentation update
- [ ] Release preparation

## Task List with Code Blocks

- [ ] Implement the following function:
  ```go
  func Process(data []byte) error {
      // TODO: implement
      return nil
  }
  ```
- [x] Add unit tests
- [ ] Update documentation

## Task List in Blockquote

> **Review Checklist:**
> - [x] Code follows style guide
> - [x] Tests are passing
> - [ ] Documentation updated
> - [ ] CHANGELOG entry added

## Bug Tracking Task List

- [ ] Bug #101: Fix null pointer exception
  - [x] Reproduce issue
  - [x] Identify root cause
  - [ ] Implement fix
  - [ ] Add regression test

- [x] Bug #102: Memory leak in cache
  - [x] Profile memory usage
  - [x] Fix leak
  - [x] Verify fix

## Feature Development Task List

### Feature: Dark Mode
- [x] Design system colors
- [x] Create theme context
- [ ] Implement toggle component
- [ ] Persist preference
- [ ] Add system preference detection

### Feature: Export to PDF
- [ ] Research PDF libraries
- [ ] Implement basic export
- [ ] Add styling options
- [ ] Support page breaks

## Daily Standup Tasks

**Monday**
- [x] Review PR #45
- [x] Fix failing tests
- [ ] Meeting with design team

**Tuesday**
- [ ] Implement new endpoint
- [ ] Update API docs
- [ ] Code review

## Task Dependencies

1. - [x] ~~Database setup~~ ✓
2. - [ ] API endpoints (depends on #1)
3. - [ ] Frontend integration (depends on #2)
4. - [ ] End-to-end tests (depends on #3)

## Edge Cases

- [ ]Whitespace handling (no space after bracket)
- [x] Normal checked item
- [ ]  Extra space handling
- [X] Capital X checked (should render as checked)
- [  ] Two spaces inside (may not render as task)
- Empty item below:
- [ ]
