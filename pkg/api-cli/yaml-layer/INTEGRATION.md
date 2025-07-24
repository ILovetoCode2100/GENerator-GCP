# YAML Layer Integration Guide

## Integration with Existing CLI

### 1. CLI Command Structure

```bash
# New YAML commands
api-cli yaml validate <file.yaml>           # Validate YAML syntax and semantics
api-cli yaml compile <file.yaml>            # Compile to commands without executing
api-cli yaml run <file.yaml>                # Validate, compile, and execute
api-cli yaml generate <prompt>              # AI-generate test from prompt
api-cli yaml convert <existing-test-id>     # Convert existing test to YAML

# With options
api-cli yaml run test.yaml --env staging --var user=admin@test.com
api-cli yaml validate *.yaml --strict --max-errors 5
api-cli yaml generate "login test" --template login --output login.yaml
```

### 2. Integration Points

```go
// pkg/api-cli/commands/yaml.go
package commands

import (
    "github.com/spf13/cobra"
    "github.com/virtuoso/api-cli/pkg/api-cli/yaml-layer"
)

func init() {
    rootCmd.AddCommand(yamlCmd)
    yamlCmd.AddCommand(yamlValidateCmd)
    yamlCmd.AddCommand(yamlCompileCmd)
    yamlCmd.AddCommand(yamlRunCmd)
    yamlCmd.AddCommand(yamlGenerateCmd)
}

var yamlCmd = &cobra.Command{
    Use:   "yaml",
    Short: "YAML test layer commands",
    Long:  "Validate, compile, and run YAML test definitions",
}

var yamlRunCmd = &cobra.Command{
    Use:   "run <file.yaml>",
    Short: "Run a YAML test file",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        // Initialize service
        cfg := &yamllayer.Config{
            API: yamllayer.APIConfig{
                BaseURL:   config.BaseURL,
                AuthToken: config.AuthToken,
                OrgID:     config.OrgID,
            },
            Validation: yamllayer.ValidationConfig{
                Strict:    viper.GetBool("yaml.strict"),
                MaxErrors: viper.GetInt("yaml.max-errors"),
            },
        }

        service := yamllayer.NewService(cfg)

        // Open file
        file, err := os.Open(args[0])
        if err != nil {
            return err
        }
        defer file.Close()

        // Process YAML
        result, err := service.ProcessYAML(cmd.Context(), file)
        if err != nil {
            return err
        }

        // Display results
        return displayResults(result)
    },
}
```

### 3. API Client Integration

```go
// Extend existing client to support YAML commands
type YAMLClient struct {
    *Client
}

func (c *YAMLClient) ExecuteCommands(ctx context.Context, commands []yamllayer.Command) error {
    for _, cmd := range commands {
        switch cmd.Type {
        case "click":
            err := c.StepInteractClick(ctx, cmd.Target, cmd.Position)
        case "type":
            err := c.StepInteractWrite(ctx, cmd.Target, cmd.Value.(string), cmd.Position)
        case "navigate":
            err := c.StepNavigateTo(ctx, cmd.Target, cmd.Position)
        // ... map all command types
        }

        if err != nil {
            if cmd.Optional {
                log.Printf("Optional command failed: %v", err)
                continue
            }
            return err
        }

        // Apply wait if specified
        if cmd.Wait > 0 {
            time.Sleep(time.Duration(cmd.Wait) * time.Millisecond)
        }
    }
    return nil
}
```

### 4. Configuration

```yaml
# ~/.api-cli/virtuoso-config.yaml
api:
  auth_token: your-token
  base_url: https://api.virtuoso.qa

yaml:
  strict: false # Treat warnings as errors
  max_errors: 10 # Stop after N errors
  spell_check: true # Check selector spelling
  auto_wait: true # Add implicit waits
  screenshot_on_failure: true

  shortcuts: # Custom shortcuts
    login: "Sign In"
    logout: "Sign Out"

  templates_dir: ~/.api-cli/yaml-templates/

  validation:
    require_assertions: true
    require_cleanup: false
    max_step_count: 100
```

## Usage Workflows

### 1. Local Development Workflow

```bash
# 1. Create test
api-cli yaml generate "test user registration" > registration.yaml

# 2. Edit and validate
vim registration.yaml
api-cli yaml validate registration.yaml

# 3. Dry run (compile only)
api-cli yaml compile registration.yaml --output commands.json

# 4. Run test
api-cli yaml run registration.yaml --env dev

# 5. Run with overrides
api-cli yaml run registration.yaml \
  --var email=test@example.com \
  --var password='${ENV:TEST_PASS}' \
  --tag smoke
```

### 2. CI/CD Integration

```yaml
# .github/workflows/test.yml
name: Run Virtuoso Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Install Virtuoso CLI
        run: |
          curl -L https://virtuoso.qa/cli/install.sh | bash

      - name: Validate all tests
        run: |
          api-cli yaml validate tests/**/*.yaml --strict

      - name: Run smoke tests
        run: |
          api-cli yaml run tests/smoke/*.yaml \
            --env staging \
            --tag ci \
            --parallel 4

      - name: Upload results
        if: always()
        uses: actions/upload-artifact@v2
        with:
          name: test-results
          path: virtuoso-results/
```

### 3. Test Organization

```
tests/
├── smoke/
│   ├── login.yaml
│   ├── search.yaml
│   └── checkout.yaml
├── regression/
│   ├── user-management.yaml
│   ├── payment-flows.yaml
│   └── admin-features.yaml
├── data/
│   ├── users.yaml
│   └── products.yaml
└── common/
    ├── auth.yaml
    ├── navigation.yaml
    └── assertions.yaml
```

## Advanced Features

### 1. Parallel Execution

```go
func (s *Service) RunParallel(ctx context.Context, files []string, workers int) error {
    sem := make(chan struct{}, workers)
    errCh := make(chan error, len(files))

    for _, file := range files {
        sem <- struct{}{}
        go func(f string) {
            defer func() { <-sem }()

            result, err := s.RunFile(ctx, f)
            if err != nil {
                errCh <- fmt.Errorf("%s: %w", f, err)
            }
        }(file)
    }

    // Wait for completion
    for i := 0; i < cap(sem); i++ {
        sem <- struct{}{}
    }

    close(errCh)

    // Collect errors
    var errs []error
    for err := range errCh {
        errs = append(errs, err)
    }

    if len(errs) > 0 {
        return fmt.Errorf("parallel execution failed: %v", errs)
    }

    return nil
}
```

### 2. Test Discovery

```go
func DiscoverTests(root string) ([]TestInfo, error) {
    var tests []TestInfo

    err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
            // Parse test metadata
            test, err := ParseTestInfo(path)
            if err != nil {
                log.Printf("Skipping %s: %v", path, err)
                return nil
            }

            tests = append(tests, test)
        }

        return nil
    })

    return tests, err
}
```

### 3. Reporting Integration

```go
type Reporter interface {
    OnTestStart(test *Test)
    OnStepComplete(step Step, result StepResult)
    OnTestComplete(test *Test, result *TestResult)
    GenerateReport() error
}

type HTMLReporter struct {
    results []TestResult
    outputPath string
}

func (r *HTMLReporter) GenerateReport() error {
    // Generate HTML report with:
    // - Test summary
    // - Step-by-step results
    // - Screenshots
    // - Performance metrics
    // - Error details
}
```

### 4. IDE Integration

```json
// VS Code extension settings
{
  "virtuoso.yaml.validation": true,
  "virtuoso.yaml.autocomplete": true,
  "virtuoso.yaml.hover": true,
  "virtuoso.yaml.snippets": [
    {
      "prefix": "vtest",
      "body": [
        "test: ${1:Test Name}",
        "nav: ${2:/}",
        "do:",
        "  - ${3:c}: ${4:element}",
        "  - ch: ${5:assertion}"
      ]
    }
  ]
}
```

## Performance Optimizations

### 1. Command Batching

```go
// Batch similar commands for efficiency
func (o *Optimizer) batchCommands(commands []Command) []Command {
    batched := []Command{}
    current := []Command{}

    for _, cmd := range commands {
        if len(current) == 0 || canBatch(current[0], cmd) {
            current = append(current, cmd)
        } else {
            if len(current) > 1 {
                batched = append(batched, createBatch(current))
            } else {
                batched = append(batched, current[0])
            }
            current = []Command{cmd}
        }
    }

    return batched
}
```

### 2. Caching

```go
type Cache struct {
    compiled map[string][]Command
    mu       sync.RWMutex
}

func (c *Cache) GetOrCompile(file string, compiler *Compiler) ([]Command, error) {
    c.mu.RLock()
    if commands, ok := c.compiled[file]; ok {
        c.mu.RUnlock()
        return commands, nil
    }
    c.mu.RUnlock()

    // Compile and cache
    commands, err := compiler.CompileFile(file)
    if err != nil {
        return nil, err
    }

    c.mu.Lock()
    c.compiled[file] = commands
    c.mu.Unlock()

    return commands, nil
}
```

## Monitoring and Debugging

### 1. Debug Mode

```yaml
# Enable debug output
api-cli yaml run test.yaml --debug
# Debug output includes:
# - Parsed YAML structure
# - Validation results
# - Compiled commands
# - API requests/responses
# - Timing information
```

### 2. Metrics Collection

```go
type Metrics struct {
    ParseTime    time.Duration
    ValidateTime time.Duration
    CompileTime  time.Duration
    ExecuteTime  time.Duration
    StepTimes    map[string]time.Duration
}

func (m *Metrics) Report() {
    log.Printf("Performance Report:")
    log.Printf("  Parse: %v", m.ParseTime)
    log.Printf("  Validate: %v", m.ValidateTime)
    log.Printf("  Compile: %v", m.CompileTime)
    log.Printf("  Execute: %v", m.ExecuteTime)
    log.Printf("  Slowest steps:")
    // ... report slowest steps
}
```

## Migration Guide

### Converting Existing Tests

```bash
# Convert existing Virtuoso test to YAML
api-cli yaml convert <checkpoint-id> > converted.yaml

# Bulk conversion
for id in $(api-cli list-checkpoints <journey-id> -o json | jq -r '.[].id'); do
  api-cli yaml convert $id > "tests/$id.yaml"
done
```

### Gradual Adoption

1. Start with new tests in YAML
2. Convert critical tests first
3. Build common block library
4. Train team on YAML syntax
5. Integrate into CI/CD
6. Monitor and optimize

## Best Practices

1. **File Organization**

   - One test per file for simple tests
   - Group related tests in directories
   - Share common blocks via includes

2. **Naming Conventions**

   - Files: `feature-action.yaml` (e.g., `user-login.yaml`)
   - Variables: `camelCase` (e.g., `userName`)
   - Blocks: `snake_case` (e.g., `fill_shipping_form`)

3. **Version Control**

   - Commit YAML files
   - Review changes in PRs
   - Tag stable test suites

4. **Maintenance**
   - Regular validation runs
   - Update selectors proactively
   - Monitor flaky tests
   - Refactor common patterns
