#!/bin/bash

# Bookmark Sync Service Test Runner Script
# 書籤同步服務測試運行腳本

set -e

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go first."
        exit 1
    fi
    log_info "Go version: $(go version)"
}

# Check dependencies
check_dependencies() {
    log_info "Checking dependencies..."

    if [ ! -f "go.mod" ]; then
        log_error "go.mod file not found. Please run from project root."
        exit 1
    fi

    log_info "Downloading dependencies..."
    go mod download
    go mod tidy

    log_success "Dependencies checked and updated"
}

# Run unit tests
run_unit_tests() {
    log_info "Running unit tests..."

    export GO_ENV=test
    export LOG_LEVEL=error

    # Run tests with coverage
    # 運行帶有覆蓋率的測試
    go test -v -race -coverprofile=coverage.out -covermode=atomic ./backend/...

    if [ $? -eq 0 ]; then
        log_success "All unit tests passed!"
    else
        log_error "Some unit tests failed!"
        exit 1
    fi
}

# Generate coverage report
generate_coverage_report() {
    log_info "Generating coverage report..."

    if [ -f "coverage.out" ]; then
        go tool cover -html=coverage.out -o coverage.html
        go tool cover -func=coverage.out | tail -1
        log_success "Coverage report generated: coverage.html"
    else
        log_warning "Coverage file not found"
    fi
}

# Run static analysis
run_static_analysis() {
    log_info "Running static analysis..."

    if command -v golangci-lint &> /dev/null; then
        golangci-lint run ./...
        if [ $? -eq 0 ]; then
            log_success "Static analysis passed!"
        else
            log_warning "Static analysis found issues"
        fi
    else
        log_warning "golangci-lint not installed, skipping static analysis"
    fi

    go vet ./...
    if [ $? -eq 0 ]; then
        log_success "go vet passed!"
    else
        log_error "go vet found issues!"
        exit 1
    fi

    unformatted=$(gofmt -l .)
    if [ -n "$unformatted" ]; then
        log_error "The following files are not formatted:"
        echo "$unformatted"
        log_info "Run 'gofmt -w .' to format them"
        exit 1
    else
        log_success "All files are properly formatted!"
    fi
}

# Show help
show_help() {
    echo "Bookmark Sync Service Test Runner"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -u, --unit          Run only unit tests"
    echo "  -c, --coverage      Generate coverage report"
    echo "  -s, --static        Run static analysis"
    echo "  -a, --all           Run all tests (default)"
    echo "  -h, --help          Show this help message"
}

# Main function
main() {
    local run_unit=false
    local run_coverage=false
    local run_static=false
    local run_all=true

    while [[ $# -gt 0 ]]; do
        case $1 in
            -u|--unit)
                run_unit=true
                run_all=false
                shift
                ;;
            -c|--coverage)
                run_coverage=true
                run_all=false
                shift
                ;;
            -s|--static)
                run_static=true
                run_all=false
                shift
                ;;
            -a|--all)
                run_all=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    check_go
    check_dependencies

    if [ "$run_all" = true ]; then
        run_unit_tests
        generate_coverage_report
        run_static_analysis
    else
        [ "$run_unit" = true ] && run_unit_tests
        [ "$run_coverage" = true ] && { run_unit_tests; generate_coverage_report; }
        [ "$run_static" = true ] && run_static_analysis
    fi

    log_success "All tests completed successfully!"
}

main "$@"