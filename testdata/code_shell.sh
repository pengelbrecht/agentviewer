#!/usr/bin/env bash
#
# Shell Script Code Test File - Tests syntax highlighting for Bash/Shell language features
# This file demonstrates various Shell scripting constructs for testing code rendering
#

# Strict mode
set -euo pipefail
IFS=$'\n\t'

# Constants
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"
readonly VERSION="1.0.0"
readonly MAX_RETRIES=3
readonly TIMEOUT=30

# Color codes
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Default values
DEBUG=${DEBUG:-false}
VERBOSE=${VERBOSE:-false}
DRY_RUN=${DRY_RUN:-false}

# Logging functions
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo -e "[${timestamp}] [${level}] ${message}"
}

debug() {
    if [[ "${DEBUG}" == "true" ]]; then
        log "DEBUG" "$@"
    fi
}

info() {
    log "INFO" "${GREEN}$*${NC}"
}

warn() {
    log "WARN" "${YELLOW}$*${NC}" >&2
}

error() {
    log "ERROR" "${RED}$*${NC}" >&2
}

# Cleanup trap
cleanup() {
    local exit_code=$?
    debug "Cleaning up (exit code: ${exit_code})"

    # Remove temp files
    if [[ -n "${TEMP_DIR:-}" && -d "${TEMP_DIR}" ]]; then
        rm -rf "${TEMP_DIR}"
    fi

    exit "$exit_code"
}

trap cleanup EXIT
trap 'error "Script interrupted"; exit 130' INT TERM

# Utility functions
die() {
    error "$@"
    exit 1
}

require_command() {
    local cmd="$1"
    if ! command -v "${cmd}" &>/dev/null; then
        die "Required command not found: ${cmd}"
    fi
}

is_true() {
    local value="${1:-}"
    case "${value,,}" in
        true|yes|1|on) return 0 ;;
        *) return 1 ;;
    esac
}

# String manipulation
trim() {
    local var="$*"
    # Remove leading whitespace
    var="${var#"${var%%[![:space:]]*}"}"
    # Remove trailing whitespace
    var="${var%"${var##*[![:space:]]}"}"
    printf '%s' "$var"
}

to_upper() {
    echo "${1^^}"
}

to_lower() {
    echo "${1,,}"
}

# Array operations
declare -a ITEMS=()
declare -A CONFIG=()

add_item() {
    ITEMS+=("$1")
}

get_item() {
    local index="$1"
    if [[ $index -lt ${#ITEMS[@]} ]]; then
        echo "${ITEMS[$index]}"
    fi
}

set_config() {
    local key="$1"
    local value="$2"
    CONFIG["$key"]="$value"
}

get_config() {
    local key="$1"
    local default="${2:-}"
    echo "${CONFIG[$key]:-$default}"
}

# File operations
ensure_dir() {
    local dir="$1"
    if [[ ! -d "${dir}" ]]; then
        mkdir -p "${dir}"
    fi
}

safe_copy() {
    local src="$1"
    local dst="$2"

    if [[ ! -f "${src}" ]]; then
        error "Source file not found: ${src}"
        return 1
    fi

    if [[ "${DRY_RUN}" == "true" ]]; then
        info "[DRY RUN] Would copy ${src} to ${dst}"
        return 0
    fi

    cp -v "${src}" "${dst}"
}

# HTTP operations
http_get() {
    local url="$1"
    local output="${2:-}"
    local retries="${3:-$MAX_RETRIES}"

    local attempt=1
    while [[ $attempt -le $retries ]]; do
        debug "HTTP GET ${url} (attempt ${attempt}/${retries})"

        local status
        if [[ -n "${output}" ]]; then
            status=$(curl -sS -o "${output}" -w "%{http_code}" --connect-timeout "${TIMEOUT}" "${url}")
        else
            status=$(curl -sS -o /dev/null -w "%{http_code}" --connect-timeout "${TIMEOUT}" "${url}")
        fi

        if [[ "${status}" == "200" ]]; then
            return 0
        fi

        warn "HTTP request failed with status ${status}"
        ((attempt++))
        sleep $((2 ** attempt))
    done

    error "All ${retries} attempts failed for ${url}"
    return 1
}

# Process handling
run_with_timeout() {
    local timeout="$1"
    shift
    local cmd=("$@")

    timeout --signal=KILL "${timeout}" "${cmd[@]}" || {
        local exit_code=$?
        if [[ $exit_code -eq 124 ]]; then
            error "Command timed out after ${timeout}"
        fi
        return $exit_code
    }
}

# Background job management
declare -a PIDS=()

start_background() {
    local name="$1"
    shift

    "$@" &
    local pid=$!
    PIDS+=("$pid")

    debug "Started background job '${name}' with PID ${pid}"
}

wait_for_jobs() {
    local failed=0
    for pid in "${PIDS[@]}"; do
        wait "$pid" || ((failed++))
    done
    PIDS=()
    return $failed
}

# Parameter expansion examples
param_expansion_demo() {
    local var="Hello World"

    # Length
    echo "Length: ${#var}"

    # Substring
    echo "Substring: ${var:0:5}"

    # Default value
    echo "Default: ${UNDEFINED_VAR:-default_value}"

    # Error if undefined
    : "${REQUIRED_VAR:?Variable REQUIRED_VAR must be set}"

    # Pattern substitution
    echo "Replace first: ${var/o/0}"
    echo "Replace all: ${var//o/0}"
    echo "Remove prefix: ${var#Hello }"
    echo "Remove suffix: ${var% World}"

    # Case modification
    echo "Uppercase: ${var^^}"
    echo "Lowercase: ${var,,}"
    echo "Capitalize: ${var^}"
}

# Conditional expressions
conditionals_demo() {
    local file="/tmp/test.txt"
    local num=42
    local str="hello"

    # File tests
    if [[ -f "${file}" ]]; then
        echo "File exists"
    elif [[ -d "${file}" ]]; then
        echo "Directory exists"
    else
        echo "Does not exist"
    fi

    # String tests
    [[ -z "${str}" ]] && echo "Empty string"
    [[ -n "${str}" ]] && echo "Non-empty string"
    [[ "${str}" == "hello" ]] && echo "Strings match"
    [[ "${str}" =~ ^hel ]] && echo "Regex match"

    # Numeric comparisons
    if (( num > 40 )); then
        echo "Greater than 40"
    fi

    # Compound conditions
    if [[ -n "${str}" && $num -gt 0 ]]; then
        echo "Both conditions true"
    fi

    # Ternary-like
    local result
    (( num > 50 )) && result="big" || result="small"
    echo "Result: ${result}"
}

# Loops
loops_demo() {
    # For loop with sequence
    for i in {1..5}; do
        echo "Iteration $i"
    done

    # For loop with array
    local arr=(one two three)
    for item in "${arr[@]}"; do
        echo "Item: ${item}"
    done

    # C-style for loop
    for ((i = 0; i < 3; i++)); do
        echo "Index: $i"
    done

    # While loop
    local count=0
    while [[ $count -lt 3 ]]; do
        echo "Count: $count"
        ((count++))
    done

    # Until loop
    until [[ $count -eq 0 ]]; do
        echo "Countdown: $count"
        ((count--))
    done

    # Read loop
    while IFS= read -r line; do
        echo "Line: ${line}"
    done < <(echo -e "first\nsecond\nthird")
}

# Case statement
case_demo() {
    local input="${1:-}"

    case "${input}" in
        start|run)
            echo "Starting..."
            ;;
        stop|halt)
            echo "Stopping..."
            ;;
        restart)
            echo "Restarting..."
            ;;
        [0-9]*)
            echo "Numeric input: ${input}"
            ;;
        *)
            echo "Unknown command: ${input}"
            return 1
            ;;
    esac
}

# Functions with local and global scope
outer_function() {
    local local_var="local"
    GLOBAL_VAR="global"

    inner_function() {
        echo "Inner: ${local_var}, ${GLOBAL_VAR}"
    }

    inner_function
}

# Arithmetic
arithmetic_demo() {
    local a=10 b=3

    echo "Add: $((a + b))"
    echo "Sub: $((a - b))"
    echo "Mul: $((a * b))"
    echo "Div: $((a / b))"
    echo "Mod: $((a % b))"
    echo "Power: $((a ** 2))"

    # Increment/decrement
    ((a++))
    ((b--))

    # Compound assignment
    ((a += 5))
    ((b *= 2))

    echo "a=${a}, b=${b}"
}

# Here documents
heredoc_demo() {
    cat <<'EOF'
This is a here document
with 'single quotes' and "double quotes"
and $variables that are NOT expanded
EOF

    cat <<EOF
This is a here document
with variable expansion: ${VERSION}
EOF

    # Here string
    while read -r word; do
        echo "Word: ${word}"
    done <<< "one two three"
}

# Usage message
usage() {
    cat <<USAGE
${SCRIPT_NAME} - A comprehensive shell script example

Usage:
    ${SCRIPT_NAME} [OPTIONS] COMMAND [ARGS...]

Options:
    -h, --help          Show this help message
    -v, --verbose       Enable verbose output
    -d, --debug         Enable debug mode
    -n, --dry-run       Show what would be done without doing it
    --version           Show version

Commands:
    run                 Run the main process
    test                Run tests
    demo                Run feature demos

Examples:
    ${SCRIPT_NAME} run
    ${SCRIPT_NAME} --verbose test
    ${SCRIPT_NAME} -d demo

USAGE
}

# Parse command line arguments
parse_args() {
    local positional=()

    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help)
                usage
                exit 0
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -d|--debug)
                DEBUG=true
                shift
                ;;
            -n|--dry-run)
                DRY_RUN=true
                shift
                ;;
            --version)
                echo "${SCRIPT_NAME} version ${VERSION}"
                exit 0
                ;;
            --)
                shift
                positional+=("$@")
                break
                ;;
            -*)
                die "Unknown option: $1"
                ;;
            *)
                positional+=("$1")
                shift
                ;;
        esac
    done

    # Restore positional parameters
    set -- "${positional[@]}"
    ARGS=("$@")
}

# Main function
main() {
    parse_args "$@"

    local command="${ARGS[0]:-}"

    case "${command}" in
        run)
            info "Running main process..."
            ;;
        test)
            info "Running tests..."
            ;;
        demo)
            info "Running demos..."
            conditionals_demo
            loops_demo
            arithmetic_demo
            heredoc_demo
            ;;
        "")
            usage
            exit 0
            ;;
        *)
            die "Unknown command: ${command}"
            ;;
    esac
}

# Entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
