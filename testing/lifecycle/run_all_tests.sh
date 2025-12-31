#!/bin/bash
# Run all lifecycle tests

set -e

cd "$(dirname "$0")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

echo "========================================"
echo "Running All Lifecycle Tests"
echo "========================================"

# Find all run_test.sh files
TESTS=$(find . -name "run_test.sh" -not -path "./run_lifecycle_test.sh" | sort)

PASSED=0
FAILED=0
FAILED_TESTS=""

for test in $TESTS; do
    dir=$(dirname "$test")
    resource=$(basename "$dir")

    echo ""
    echo "----------------------------------------"
    echo "Testing: $resource"
    echo "----------------------------------------"

    if bash "$test"; then
        echo -e "${GREEN}✓ $resource passed${NC}"
        ((PASSED++))
    else
        echo -e "${RED}✗ $resource failed${NC}"
        ((FAILED++))
        FAILED_TESTS="$FAILED_TESTS $resource"
    fi
done

echo ""
echo "========================================"
echo "Results: $PASSED passed, $FAILED failed"
if [ $FAILED -gt 0 ]; then
    echo -e "${RED}Failed tests:$FAILED_TESTS${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
fi
