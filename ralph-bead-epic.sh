#!/bin/bash
# Ralph Pattern: Finish a Beads Epic
#
# Runs Claude in an autonomous loop to complete all tasks in a Beads epic.
# Claude operates bd directly - the script just loops until COMPLETE.
#
# Beads: https://github.com/steveyegge/beads
#
# Usage: ./ralph-bead-epic.sh <max-iterations> [epic-id]
#
# If epic-id is omitted, automatically selects the highest priority ready epic.
#
# Prerequisites:
# - Claude CLI installed
# - Beads CLI installed (bd)

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <iterations> [epic-id]"
    echo ""
    echo "Examples:"
    echo "  $0 20 bd-a3f8          # Complete up to 20 tasks in epic bd-a3f8"
    echo "  $0 50 bd-c7e2          # Complete up to 50 tasks in epic bd-c7e2"
    echo "  $0 30                  # Auto-select highest priority ready epic"
    exit 1
fi

MAX_ITERATIONS=$1

if [ -z "$2" ]; then
    echo "No epic specified, finding highest priority ready epic..."
    EPIC_ID=$(bd ready --type=epic --limit=1 --sort=priority --json 2>/dev/null | jq -r '.[0].id // empty')
    if [ -z "$EPIC_ID" ]; then
        echo "Error: No ready epics found. Create one with: bd create --type=epic --title=\"...\""
        exit 1
    fi
    echo "Selected: $EPIC_ID"
else
    EPIC_ID="$2"
fi

echo "Starting Ralph loop for Beads epic..."
echo "Max iterations: $MAX_ITERATIONS"
echo "Epic: $EPIC_ID"
echo ""

for ((i=1; i<=MAX_ITERATIONS; i++)); do
    echo "=== Iteration $i of $MAX_ITERATIONS ==="

    result=$(claude --dangerously-skip-permissions --print "
BEADS EPIC RALPH - ITERATION $i

Epic: $EPIC_ID

PROCESS:
1. Run 'bd ready' to see unblocked tasks for epic $EPIC_ID.
2. If no tasks remain in this epic, output <promise>COMPLETE</promise> and stop.
3. Pick the highest priority ready task from this epic.
4. Implement it fully - code, tests, verification.
5. Run tests to ensure nothing broke.
6. Mark complete: bd done <task-id>
7. Commit changes: git add -A && git commit -m \"feat(<task-id>): <brief description>\"
8. Add iteration note: bd note $EPIC_ID \"Iteration $i: <task-id> | <what you did> | <any blockers or learnings>\"
9. Summarize what you did.

RULES:
- Complete ONE task per iteration.
- Always leave tests passing.
- Only work on tasks under epic $EPIC_ID.
- Use bd commands directly (bd ready, bd show, bd done, bd note, bd block).
- Always run bd note after completing a task - this is required, not optional.
- Only output <promise>COMPLETE</promise> when bd ready shows no remaining tasks for this epic.
")

    echo "$result"

    if [[ "$result" == *"<promise>COMPLETE</promise>"* ]]; then
        echo ""
        echo "=== Epic completed after $i iterations ==="
        exit 0
    fi

    echo ""
    sleep 2
done

echo ""
echo "=== Max iterations ($MAX_ITERATIONS) reached ==="
exit 1
