#!/bin/bash
set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Starting test coverage generation...${NC}"

mkdir -p coverage

# Function to run tests for a module
run_tests() {
    local module_path=$1
    local module_name=$2
    
    echo -e "${GREEN}Testing ${module_name}...${NC}"
    
    cd "${module_path}"
    
    # Run tests with coverage
    go test -v -coverprofile="../coverage/${module_name}.out" -covermode=atomic ./...
    
    cd - > /dev/null
}

# Test pkg (shared code)
run_tests "./pkg" "pkg"

# Test API
run_tests "./api" "api"

# Test Fetcher
run_tests "./fetcher" "fetcher"

# Merge coverage files for SonarQube
echo -e "${BLUE}Merging coverage files...${NC}"
echo "mode: atomic" > coverage/coverage.out

tail -n +2 coverage/pkg.out >> coverage/coverage.out 2>/dev/null || true
tail -n +2 coverage/api.out >> coverage/coverage.out 2>/dev/null || true
tail -n +2 coverage/fetcher.out >> coverage/coverage.out 2>/dev/null || true

# Remove test files from report.
for file in coverage/pkg.out coverage/api.out coverage/fetcher.out; do
    if [ -f "$file" ]; then
        tail -n +2 "$file" | grep -v "_test.go" >> coverage/coverage.out || true
    fi
done

COVERAGE=$(go tool cover -func=coverage/coverage.out | grep total | awk '{print $3}')

echo -e "${GREEN}Total coverage: ${COVERAGE}${NC}"
echo -e "${GREEN}Coverage report generated: ./coverage/coverage.out${NC}"

rm -f coverage/pkg.out coverage/api.out coverage/fetcher.out

exit 0