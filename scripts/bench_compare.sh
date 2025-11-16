#!/usr/bin/env bash
# bench_compare.sh - Compare benchmark results to detect performance regressions
#
# Usage:
#   ./scripts/bench_compare.sh [baseline.txt] [current.txt]
#
# If no arguments provided, it will:
#   1. Run benchmarks and save to current.txt
#   2. Compare with baseline.txt if it exists
#
# To create a new baseline:
#   ./scripts/bench_compare.sh --save-baseline
#
# Example workflow:
#   # Create baseline before making changes
#   ./scripts/bench_compare.sh --save-baseline
#
#   # Make changes to parser...
#
#   # Compare current performance to baseline
#   ./scripts/bench_compare.sh

set -euo pipefail

# Configuration
BENCH_DIR="benchmarks"
BASELINE_FILE="${BENCH_DIR}/baseline.txt"
CURRENT_FILE="${BENCH_DIR}/current.txt"
BENCHTIME="${BENCHTIME:-1s}"  # Can be overridden via environment variable

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create benchmark directory if it doesn't exist
mkdir -p "${BENCH_DIR}"

# Function to run benchmarks
run_benchmarks() {
    local output_file=$1
    echo -e "${BLUE}Running parser benchmarks (benchtime=${BENCHTIME})...${NC}"
    go test -bench=BenchmarkParser -benchmem -benchtime="${BENCHTIME}" \
        ./internal/parser | tee "${output_file}"
}

# Function to save baseline
save_baseline() {
    echo -e "${BLUE}Saving new baseline...${NC}"
    run_benchmarks "${BASELINE_FILE}"
    echo -e "${GREEN}Baseline saved to ${BASELINE_FILE}${NC}"
}

# Function to compare benchmarks using benchstat (if available)
compare_with_benchstat() {
    local baseline=$1
    local current=$2

    if ! command -v benchstat &> /dev/null; then
        echo -e "${YELLOW}benchstat not found. Install with: go install golang.org/x/perf/cmd/benchstat@latest${NC}"
        return 1
    fi

    echo -e "\n${BLUE}=== Benchmark Comparison ===${NC}\n"
    benchstat "${baseline}" "${current}"
    return 0
}

# Function to do simple comparison without benchstat
simple_compare() {
    local baseline=$1
    local current=$2

    echo -e "\n${BLUE}=== Simple Benchmark Comparison ===${NC}\n"
    echo -e "${YELLOW}(Install benchstat for detailed statistical comparison)${NC}\n"

    # Extract benchmark results (lines starting with "Benchmark")
    local baseline_results=$(grep "^Benchmark" "${baseline}" || true)
    local current_results=$(grep "^Benchmark" "${current}" || true)

    if [[ -z "${baseline_results}" ]] || [[ -z "${current_results}" ]]; then
        echo -e "${RED}Error: Could not extract benchmark results${NC}"
        return 1
    fi

    # Display side-by-side comparison
    echo "Baseline:"
    echo "${baseline_results}"
    echo ""
    echo "Current:"
    echo "${current_results}"
    echo ""

    # Simple percentage calculation for major benchmarks
    echo -e "${BLUE}Performance Changes:${NC}"
    # Parse benchmark output format: "BenchmarkName-N    iterations    time ns/op"
    # Example: "BenchmarkParser-8    1000000    1234 ns/op"
    # The regex captures: (1) benchmark name including CPU count, (2) time in ns
    while IFS= read -r line; do
        if [[ $line =~ ^(Benchmark[^[:space:]]+)[[:space:]]+[0-9]+[[:space:]]+([0-9]+)[[:space:]]ns/op ]]; then
            bench_name="${BASH_REMATCH[1]}"
            baseline_time="${BASH_REMATCH[2]}"

            current_line=$(echo "${current_results}" | grep "^${bench_name}" || true)
            if [[ $current_line =~ [[:space:]]([0-9]+)[[:space:]]ns/op ]]; then
                current_time="${BASH_REMATCH[1]}"

                # Calculate percentage change
                change=$(awk "BEGIN {printf \"%.2f\", (($current_time - $baseline_time) / $baseline_time) * 100}")

                # Color code the output
                if awk "BEGIN {exit !($change > 10)}"; then
                    color="${RED}"
                    symbol="▲"
                elif awk "BEGIN {exit !($change < -10)}"; then
                    color="${GREEN}"
                    symbol="▼"
                else
                    color="${NC}"
                    symbol="≈"
                fi

                printf "  %-40s %s %6.2f%%${NC}\n" "${bench_name}" "${symbol}" "${change}"
            fi
        fi
    done <<< "${baseline_results}"
}

# Main logic
case "${1:-}" in
    --save-baseline)
        save_baseline
        ;;
    --help|-h)
        cat << EOF
bench_compare.sh - Compare benchmark results

Usage:
    ./scripts/bench_compare.sh [OPTIONS] [baseline.txt] [current.txt]

Options:
    --save-baseline     Run benchmarks and save as new baseline
    --help, -h          Show this help message

Environment Variables:
    BENCHTIME          Benchmark duration (default: 1s)
                      Examples: 100ms, 500ms, 2s, 5s

Examples:
    # Save new baseline
    ./scripts/bench_compare.sh --save-baseline

    # Run and compare to baseline
    ./scripts/bench_compare.sh

    # Compare two specific files
    ./scripts/bench_compare.sh old.txt new.txt

    # Run with custom benchtime
    BENCHTIME=5s ./scripts/bench_compare.sh --save-baseline

Benchmark Results:
    Results are saved in ./benchmarks/ directory
    - baseline.txt: Reference baseline for comparison
    - current.txt:  Most recent benchmark run

Install benchstat for detailed statistical analysis:
    go install golang.org/x/perf/cmd/benchstat@latest
EOF
        ;;
    *)
        # Determine baseline and current files
        if [[ $# -eq 2 ]]; then
            BASELINE_FILE=$1
            CURRENT_FILE=$2
        elif [[ $# -eq 0 ]]; then
            # Run current benchmarks
            run_benchmarks "${CURRENT_FILE}"
        else
            echo -e "${RED}Error: Invalid arguments${NC}"
            echo "Usage: $0 [baseline.txt] [current.txt]"
            echo "       $0 --save-baseline"
            echo "       $0 --help"
            exit 1
        fi

        # Check if baseline exists
        if [[ ! -f "${BASELINE_FILE}" ]]; then
            echo -e "${YELLOW}No baseline found at ${BASELINE_FILE}${NC}"
            echo -e "${YELLOW}Run with --save-baseline to create one${NC}"
            exit 0
        fi

        # Compare results
        if ! compare_with_benchstat "${BASELINE_FILE}" "${CURRENT_FILE}"; then
            simple_compare "${BASELINE_FILE}" "${CURRENT_FILE}"
        fi
        ;;
esac
