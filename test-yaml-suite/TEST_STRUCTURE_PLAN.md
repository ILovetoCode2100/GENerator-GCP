# Virtuoso API CLI - Comprehensive YAML Test Suite Structure

## Overview

This document outlines the comprehensive test structure for validating all 69 CLI commands,
edge cases, error scenarios, and YAML features.

## Test Suite Organization

### Directory Structure

```
test-yaml-suite/
├── README.md                          # Test suite documentation
├── test-runner.sh                     # Main test execution script
├── config/
│   ├── test-config.yaml              # Test environment configuration
│   ├── test-data.yaml                # Shared test data definitions
│   └── test-fixtures.yaml            # Reusable test fixtures
│
├── commands/                          # Command-specific tests
│   ├── step-assert/
│   │   ├── positive/
│   │   │   ├── exists.yaml
│   │   │   ├── not-exists.yaml
│   │   │   ├── equals.yaml
│   │   │   ├── not-equals.yaml
│   │   │   ├── checked.yaml
│   │   │   ├── selected.yaml
│   │   │   ├── variable.yaml
│   │   │   ├── gt.yaml
│   │   │   ├── gte.yaml
│   │   │   ├── lt.yaml
│   │   │   ├── lte.yaml
│   │   │   └── matches.yaml
│   │   ├── edge-cases/
│   │   │   ├── special-characters.yaml
│   │   │   ├── unicode.yaml
│   │   │   ├── boundary-values.yaml
│   │   │   └── xpath-variations.yaml
│   │   └── negative/
│   │       ├── invalid-args.yaml
│   │       ├── missing-elements.yaml
│   │       └── type-mismatches.yaml
│   │
│   ├── step-interact/
│   │   ├── positive/
│   │   │   ├── click-variations.yaml
│   │   │   ├── write-operations.yaml
│   │   │   ├── keyboard-events.yaml
│   │   │   ├── mouse-operations.yaml
│   │   │   └── select-operations.yaml
│   │   ├── edge-cases/
│   │   │   ├── multi-language-input.yaml
│   │   │   ├── special-keys.yaml
│   │   │   ├── coordinate-boundaries.yaml
│   │   │   └── dynamic-elements.yaml
│   │   └── negative/
│   │       ├── invalid-selectors.yaml
│   │       ├── disabled-elements.yaml
│   │       └── timing-issues.yaml
│   │
│   ├── step-navigate/
│   │   ├── positive/
│   │   │   ├── url-navigation.yaml
│   │   │   ├── scroll-operations.yaml
│   │   │   └── anchor-navigation.yaml
│   │   ├── edge-cases/
│   │   │   ├── special-urls.yaml
│   │   │   ├── scroll-limits.yaml
│   │   │   └── protocol-variations.yaml
│   │   └── negative/
│   │       ├── invalid-urls.yaml
│   │       └── scroll-errors.yaml
│   │
│   ├── step-window/
│   │   ├── positive/
│   │   │   ├── resize-operations.yaml
│   │   │   ├── tab-switching.yaml
│   │   │   └── frame-navigation.yaml
│   │   ├── edge-cases/
│   │   │   ├── dimension-limits.yaml
│   │   │   ├── nested-frames.yaml
│   │   │   └── multiple-tabs.yaml
│   │   └── negative/
│   │       ├── invalid-dimensions.yaml
│   │       └── missing-frames.yaml
│   │
│   ├── step-data/
│   │   ├── positive/
│   │   │   ├── store-operations.yaml
│   │   │   └── cookie-management.yaml
│   │   ├── edge-cases/
│   │   │   ├── variable-naming.yaml
│   │   │   ├── cookie-attributes.yaml
│   │   │   └── data-persistence.yaml
│   │   └── negative/
│   │       ├── invalid-variables.yaml
│   │       └── cookie-errors.yaml
│   │
│   ├── step-dialog/
│   │   ├── positive/
│   │   │   ├── alert-handling.yaml
│   │   │   ├── confirm-handling.yaml
│   │   │   └── prompt-handling.yaml
│   │   ├── edge-cases/
│   │   │   ├── dialog-timing.yaml
│   │   │   └── multiple-dialogs.yaml
│   │   └── negative/
│   │       └── no-dialog-present.yaml
│   │
│   ├── step-wait/
│   │   ├── positive/
│   │   │   ├── element-waits.yaml
│   │   │   └── time-waits.yaml
│   │   ├── edge-cases/
│   │   │   ├── timeout-boundaries.yaml
│   │   │   └── dynamic-loading.yaml
│   │   └── negative/
│   │       └── timeout-failures.yaml
│   │
│   ├── step-file/
│   │   ├── positive/
│   │   │   └── url-uploads.yaml
│   │   ├── edge-cases/
│   │   │   └── url-variations.yaml
│   │   └── negative/
│   │       └── invalid-urls.yaml
│   │
│   ├── step-misc/
│   │   ├── positive/
│   │   │   ├── comments.yaml
│   │   │   └── javascript-execution.yaml
│   │   ├── edge-cases/
│   │   │   ├── script-complexity.yaml
│   │   │   └── comment-formats.yaml
│   │   └── negative/
│   │       └── script-errors.yaml
│   │
│   └── library/
│       ├── positive/
│       │   └── library-operations.yaml
│       ├── edge-cases/
│       │   └── step-ordering.yaml
│       └── negative/
│           └── invalid-references.yaml
│
├── workflows/                         # Complex multi-step scenarios
│   ├── e2e-scenarios/
│   │   ├── login-flow.yaml
│   │   ├── checkout-process.yaml
│   │   ├── form-submission.yaml
│   │   └── search-and-filter.yaml
│   ├── variable-workflows/
│   │   ├── data-driven-tests.yaml
│   │   ├── variable-chaining.yaml
│   │   └── conditional-execution.yaml
│   └── advanced-patterns/
│       ├── retry-patterns.yaml
│       ├── error-recovery.yaml
│       └── parallel-execution.yaml
│
├── yaml-features/                     # YAML-specific feature tests
│   ├── anchors-aliases/
│   │   ├── basic-anchors.yaml
│   │   ├── nested-references.yaml
│   │   └── merge-keys.yaml
│   ├── multi-document/
│   │   ├── stream-processing.yaml
│   │   └── document-separation.yaml
│   ├── formatting/
│   │   ├── block-styles.yaml
│   │   ├── flow-styles.yaml
│   │   ├── literal-blocks.yaml
│   │   └── folded-scalars.yaml
│   └── advanced/
│       ├── custom-tags.yaml
│       ├── complex-structures.yaml
│       └── large-documents.yaml
│
├── error-scenarios/                   # Comprehensive error testing
│   ├── malformed-yaml/
│   │   ├── syntax-errors.yaml
│   │   ├── indentation-errors.yaml
│   │   └── invalid-structures.yaml
│   ├── validation-errors/
│   │   ├── missing-required.yaml
│   │   ├── type-violations.yaml
│   │   └── constraint-failures.yaml
│   └── runtime-errors/
│       ├── api-failures.yaml
│       ├── timeout-errors.yaml
│       └── permission-errors.yaml
│
├── performance/                       # Performance and stress tests
│   ├── large-scale/
│   │   ├── bulk-operations.yaml
│   │   └── memory-stress.yaml
│   └── concurrency/
│       ├── parallel-tests.yaml
│       └── race-conditions.yaml
│
└── reports/                          # Test execution reports
    ├── coverage/
    ├── results/
    └── metrics/
```

## Test Categories

### 1. Command Coverage Tests

**Objective**: Validate all 69 CLI commands with standard positive cases

- **Coverage Target**: 100% of all command variations
- **Test Count**: ~69 base tests + ~200 variations
- **Validation**: Command execution, response format, side effects

### 2. Edge Case Tests

**Objective**: Test boundary conditions and special scenarios

- **Unicode & Internationalization**: 中文, العربية, emoji 🎉
- **Special Characters**: Quotes, backslashes, control characters
- **Boundary Values**: Max/min lengths, numeric limits, coordinate edges
- **XPath Complexity**: Nested predicates, axes, functions

### 3. Error Scenario Tests

**Objective**: Ensure graceful error handling

- **Invalid Arguments**: Wrong types, missing required, extra args
- **Malformed YAML**: Syntax errors, invalid indentation
- **API Failures**: Network errors, auth failures, rate limits
- **State Errors**: Missing elements, disabled states, timing issues

### 4. Complex Workflow Tests

**Objective**: Validate real-world usage patterns

- **Multi-Step Processes**: Login → Search → Purchase → Logout
- **Variable Management**: Store → Transform → Use → Validate
- **Conditional Logic**: If-then patterns, error recovery
- **Data-Driven**: Parameterized tests, loop constructs

### 5. YAML Feature Tests

**Objective**: Ensure full YAML 1.2 compatibility

- **Anchors & Aliases**: Reusable test components
- **Multi-Document**: Streaming test scenarios
- **Style Variations**: Block vs flow, literal vs folded
- **Advanced Features**: Custom tags, complex nesting

## Coverage Metrics

### Command Coverage Matrix

| Command Group | Commands | Positive | Edge Cases | Negative | Total Tests |
| ------------- | -------- | -------- | ---------- | -------- | ----------- |
| step-assert   | 12       | 12       | 24         | 12       | 48          |
| step-interact | 15       | 15       | 30         | 15       | 60          |
| step-navigate | 10       | 10       | 20         | 10       | 40          |
| step-window   | 5        | 5        | 10         | 5        | 20          |
| step-data     | 6        | 6        | 12         | 6        | 24          |
| step-dialog   | 5        | 5        | 10         | 5        | 20          |
| step-wait     | 2        | 2        | 4          | 2        | 8           |
| step-file     | 2        | 2        | 4          | 2        | 8           |
| step-misc     | 2        | 2        | 4          | 2        | 8           |
| library       | 6        | 6        | 12         | 6        | 24          |
| **Totals**    | **65**   | **65**   | **130**    | **65**   | **260**     |

### Additional Test Categories

| Category        | Test Count | Purpose                       |
| --------------- | ---------- | ----------------------------- |
| Workflows       | 20         | End-to-end scenarios          |
| YAML Features   | 15         | YAML specification compliance |
| Error Scenarios | 30         | Error handling validation     |
| Performance     | 10         | Stress and scale testing      |
| **Grand Total** | **335**    | **Comprehensive coverage**    |

## Execution Strategy

### 1. Test Runner Architecture

```bash
#!/bin/bash
# test-runner.sh - Main test orchestrator

# Features:
# - Parallel execution support
# - Progress tracking
# - Detailed reporting
# - Failure isolation
# - Retry mechanisms
```

### 2. Execution Phases

#### Phase 1: Environment Setup

- Validate CLI installation
- Check API connectivity
- Create test project/goal/journey
- Initialize test data

#### Phase 2: Sequential Tests

- Run smoke tests first
- Execute command groups in order
- Capture all outputs

#### Phase 3: Parallel Tests

- Execute independent test suites
- Monitor resource usage
- Aggregate results

#### Phase 4: Cleanup & Reporting

- Generate coverage reports
- Create failure summaries
- Archive test artifacts
- Cleanup test data

### 3. Test Execution Modes

```bash
# Quick smoke test (5 minutes)
./test-runner.sh --smoke

# Full regression (30 minutes)
./test-runner.sh --full

# Specific command group
./test-runner.sh --group step-assert

# Performance suite
./test-runner.sh --performance

# Debug mode with verbose output
./test-runner.sh --debug --verbose
```

### 4. Continuous Integration

```yaml
# .github/workflows/test-suite.yml
name: YAML Test Suite
on: [push, pull_request]
jobs:
  test:
    strategy:
      matrix:
        test-group: [commands, workflows, yaml-features, errors]
    steps:
      - run: ./test-runner.sh --group ${{ matrix.test-group }}
```

## Test Data Structure

### Standard Test Format

```yaml
# Example: commands/step-assert/positive/exists.yaml
metadata:
  name: "Assert Element Exists"
  description: "Validate element existence assertion"
  tags: [assert, positive, smoke]

setup:
  checkpoint_id: "${TEST_CHECKPOINT_ID}"
  base_url: "${TEST_BASE_URL}"

tests:
  - name: "Simple element exists"
    steps:
      - command: step-navigate
        action: to
        args:
          url: "${base_url}/test-page"
      - command: step-assert
        action: exists
        args:
          selector: "button.submit"
    expected:
      success: true
      step_count: 2

  - name: "Element with text exists"
    steps:
      - command: step-assert
        action: exists
        args:
          selector: "Submit Button"
    expected:
      success: true
```

### Workflow Test Format

```yaml
# Example: workflows/e2e-scenarios/login-flow.yaml
metadata:
  name: "Complete Login Flow"
  description: "End-to-end login scenario with validation"
  tags: [workflow, e2e, critical]

variables:
  username: "test@example.com"
  password: "TestPass123!"

steps:
  - name: "Navigate to login page"
    command: step-navigate
    action: to
    args:
      url: "${TEST_BASE_URL}/login"

  - name: "Enter credentials"
    command: step-interact
    action: write
    args:
      selector: "input#username"
      text: "${username}"

  - name: "Submit form"
    command: step-interact
    action: click
    args:
      selector: "button[type='submit']"

  - name: "Verify login success"
    command: step-assert
    action: exists
    args:
      selector: "Welcome, ${username}"

validation:
  - all_steps_succeed: true
  - final_url_contains: "/dashboard"
  - cookies_set: ["session_id", "user_token"]
```

### YAML Feature Test Format

```yaml
# Example: yaml-features/anchors-aliases/basic-anchors.yaml
common_elements: &common
  timeout: 5000
  retry: 3

test_cases:
  - name: "Reusable configuration"
    <<: *common
    steps:
      - &navigate_home
        command: step-navigate
        action: to
        args:
          url: "${TEST_BASE_URL}"

  - name: "Reuse navigation"
    <<: *common
    steps:
      - *navigate_home
      - command: step-assert
        action: exists
        args:
          selector: "h1"
```

## Success Metrics

### Coverage Goals

- **Command Coverage**: 100% of all 69 commands tested
- **Code Coverage**: >90% of CLI codebase exercised
- **Error Coverage**: All known error conditions tested
- **YAML Coverage**: All YAML 1.2 features validated

### Performance Targets

- **Execution Time**: <30 minutes for full suite
- **Parallel Efficiency**: >80% speedup with 4 workers
- **Memory Usage**: <500MB per test worker
- **API Rate**: <100 requests/minute

### Quality Metrics

- **Test Reliability**: <1% flaky test rate
- **Error Detection**: 100% of breaking changes caught
- **Documentation**: Every test self-documenting
- **Maintainability**: <5 minutes to add new test

## Implementation Priorities

### Phase 1: Foundation (Week 1)

1. Create directory structure
2. Implement test runner framework
3. Create configuration system
4. Build reporting infrastructure

### Phase 2: Core Commands (Week 2)

1. Implement all positive test cases
2. Add basic error scenarios
3. Create smoke test suite
4. Set up CI integration

### Phase 3: Advanced Testing (Week 3)

1. Add edge case tests
2. Implement workflow scenarios
3. Create YAML feature tests
4. Add performance tests

### Phase 4: Polish & Documentation (Week 4)

1. Complete error scenarios
2. Optimize test execution
3. Generate documentation
4. Create maintenance guides

## Maintenance Guidelines

### Adding New Tests

1. Follow directory structure
2. Use standard test format
3. Tag appropriately
4. Update coverage metrics
5. Run locally before commit

### Debugging Failures

1. Check test logs in `reports/`
2. Run individual test in debug mode
3. Verify API responses
4. Check for environment issues
5. Update test if API changed

### Regular Maintenance

- Weekly: Review flaky tests
- Monthly: Update test data
- Quarterly: Performance optimization
- Yearly: Major structure review
