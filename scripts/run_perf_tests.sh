#!/bin/bash

set -e

echo "🚀 Running OmniTranscripts Performance Tests"
echo "================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.23+ to run tests."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
print_status "Go version: $GO_VERSION"

# Clean previous test results
print_status "Cleaning previous test results..."
rm -rf test_results/
mkdir -p test_results/

# Run unit tests with coverage
print_status "Running unit tests with coverage..."
go test -v -race -coverprofile=test_results/coverage.out ./... 2>&1 | tee test_results/unit_tests.log

if [ ${PIPESTATUS[0]} -eq 0 ]; then
    print_success "Unit tests passed!"

    # Generate coverage report
    go tool cover -html=test_results/coverage.out -o test_results/coverage.html
    COVERAGE=$(go tool cover -func=test_results/coverage.out | grep total | awk '{print $3}')
    print_status "Code coverage: $COVERAGE"
else
    print_error "Unit tests failed!"
    exit 1
fi

# Run benchmark tests
print_status "Running benchmark tests..."
go test -bench=. -benchmem -benchtime=5s ./... 2>&1 | tee test_results/benchmarks.log

if [ ${PIPESTATUS[0]} -eq 0 ]; then
    print_success "Benchmark tests completed!"
else
    print_warning "Some benchmark tests had issues (this might be expected for external dependencies)"
fi

# Run CPU profiling benchmark
print_status "Running CPU profiling benchmark..."
go test -bench=BenchmarkAPI_HealthEndpoint -cpuprofile=test_results/cpu.prof -benchtime=10s ./handlers 2>&1 | tee test_results/cpu_profile.log

# Run memory profiling benchmark
print_status "Running memory profiling benchmark..."
go test -bench=BenchmarkMemoryUsage -memprofile=test_results/mem.prof -benchtime=10s ./lib 2>&1 | tee test_results/mem_profile.log

# Run load test (if not in short mode)
if [ "$1" != "short" ]; then
    print_status "Running load tests..."
    go test -v -run TestAPI_LoadTest ./handlers 2>&1 | tee test_results/load_test.log

    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        print_success "Load tests passed!"
    else
        print_error "Load tests failed!"
    fi
else
    print_warning "Skipping load tests (short mode)"
fi

# Generate performance report
print_status "Generating performance report..."

cat > test_results/performance_report.md << EOF
# OmniTranscripts Performance Report

**Generated on:** $(date)
**Go Version:** $GO_VERSION

## Test Coverage
**Coverage:** $COVERAGE

## Benchmark Results

### API Performance
\`\`\`
$(grep -A 10 "BenchmarkAPI" test_results/benchmarks.log || echo "No API benchmarks found")
\`\`\`

### Job Queue Performance
\`\`\`
$(grep -A 10 "BenchmarkJobQueue" test_results/benchmarks.log || echo "No job queue benchmarks found")
\`\`\`

### Memory Usage
\`\`\`
$(grep -A 10 "BenchmarkMemoryUsage" test_results/benchmarks.log || echo "No memory benchmarks found")
\`\`\`

### Transcription Pipeline
\`\`\`
$(grep -A 10 "BenchmarkGetVideoDuration\|BenchmarkValidateURL\|BenchmarkTranscriptParsing" test_results/benchmarks.log || echo "No transcription benchmarks found")
\`\`\`

## Load Test Results
\`\`\`
$(grep -A 20 "Load Test Results" test_results/load_test.log 2>/dev/null || echo "Load tests not run or failed")
\`\`\`

## Files Generated
- \`coverage.out\` - Coverage data
- \`coverage.html\` - HTML coverage report
- \`cpu.prof\` - CPU profile data
- \`mem.prof\` - Memory profile data
- \`*.log\` - Detailed test logs

## How to View Profiles
\`\`\`bash
# CPU profile
go tool pprof test_results/cpu.prof

# Memory profile
go tool pprof test_results/mem.prof
\`\`\`

EOF

print_success "Performance report generated: test_results/performance_report.md"

# Summary
print_status "Performance test summary:"
echo "  📊 Coverage: $COVERAGE"
echo "  📁 Results: test_results/"
echo "  🌐 Coverage report: test_results/coverage.html"
echo "  📈 Performance report: test_results/performance_report.md"

# Check for performance regressions (basic checks)
print_status "Checking for performance issues..."

# Check if any benchmarks are unusually slow
SLOW_BENCHMARKS=$(grep "BenchmarkAPI_HealthEndpoint" test_results/benchmarks.log | awk '{if($3 > 1000000) print $1 " is slow: " $3 " ns/op"}')
if [ -n "$SLOW_BENCHMARKS" ]; then
    print_warning "Slow benchmarks detected:"
    echo "$SLOW_BENCHMARKS"
else
    print_success "No slow benchmarks detected"
fi

# Check for high memory usage
HIGH_MEMORY=$(grep -E "BenchmarkMemoryUsage.*B/op" test_results/benchmarks.log | awk '{if($5 > 10000) print $1 " uses high memory: " $4 " B/op"}')
if [ -n "$HIGH_MEMORY" ]; then
    print_warning "High memory usage detected:"
    echo "$HIGH_MEMORY"
else
    print_success "Memory usage looks good"
fi

print_success "Performance testing completed! 🎉"