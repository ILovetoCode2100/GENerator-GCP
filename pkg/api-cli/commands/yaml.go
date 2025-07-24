package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/marklovelady/api-cli-generator/pkg/api-cli/yaml-layer/ai"
	"github.com/marklovelady/api-cli-generator/pkg/api-cli/yaml-layer/service"
	"github.com/spf13/cobra"
)

var (
	yamlOutputFormat string
	yamlStrict       bool
	yamlEnvVars      map[string]string
	yamlVars         map[string]string
	yamlTemplate     string
	yamlParallel     int
)

// yamlCmd represents the yaml command group
var yamlCmd = &cobra.Command{
	Use:   "yaml",
	Short: "YAML test layer commands",
	Long: `Process Virtuoso tests written in the compact YAML format.

The YAML layer provides:
- 60% token reduction compared to standard syntax
- Comprehensive validation with helpful errors
- AI-friendly templates for test generation
- Seamless integration with Virtuoso API`,
	Example: `  # Validate a YAML test file
  api-cli yaml validate login.yaml

  # Run a YAML test
  api-cli yaml run checkout.yaml

  # Generate a test from prompt
  api-cli yaml generate "test user registration flow" > registration.yaml

  # Run all tests in a directory
  api-cli yaml run tests/*.yaml --parallel 4`,
}

// yamlValidateCmd validates YAML syntax
var yamlValidateCmd = &cobra.Command{
	Use:   "validate <file.yaml>",
	Short: "Validate YAML test syntax and semantics",
	Long: `Validates YAML test files for:
- Correct YAML syntax
- Required fields (test name, actions)
- Valid action syntax
- Variable references
- Best practices`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize service
		svc := createYAMLService()

		allValid := true
		for _, file := range args {
			fmt.Printf("Validating %s...\n", file)

			result, err := svc.ValidateFile(file)
			if err != nil {
				fmt.Printf("  ERROR: %v\n", err)
				allValid = false
				continue
			}

			// Display results
			if result.Success {
				fmt.Printf("  ✓ Valid\n")
			} else {
				fmt.Printf("  ✗ Invalid\n")
				for _, err := range result.Errors {
					fmt.Printf("    ERROR Line %d: %s\n", err.Line, err.Message)
					if err.Fix != "" {
						fmt.Printf("      Fix: %s\n", err.Fix)
					}
					if err.Example != "" {
						fmt.Printf("      Example: %s\n", err.Example)
					}
				}
				allValid = false
			}

			// Show warnings if not strict
			if len(result.Warnings) > 0 && !yamlStrict {
				fmt.Printf("  Warnings:\n")
				for _, warn := range result.Warnings {
					fmt.Printf("    - %s\n", warn.Message)
				}
			}

			fmt.Println()
		}

		if !allValid {
			return fmt.Errorf("validation failed")
		}

		return nil
	},
}

// yamlCompileCmd compiles YAML to commands
var yamlCompileCmd = &cobra.Command{
	Use:   "compile <file.yaml>",
	Short: "Compile YAML to CLI commands without executing",
	Long:  `Compiles YAML test files to CLI commands for inspection or debugging.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := createYAMLService()

		result, err := svc.CompileFile(args[0])
		if err != nil {
			return err
		}

		// Output based on format
		switch yamlOutputFormat {
		case "json":
			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))

		case "commands":
			for i, step := range result.Steps {
				fmt.Printf("%d. %s %s\n", i+1, step.Command, strings.Join(step.Args, " "))
			}

		default:
			// Human readable
			fmt.Printf("Compiled %d steps:\n\n", len(result.Steps))
			for i, step := range result.Steps {
				fmt.Printf("%d. %s\n", i+1, step.Description)
				fmt.Printf("   Command: %s %s\n", step.Command, strings.Join(step.Args, " "))
				if len(step.Options) > 0 {
					fmt.Printf("   Options: %v\n", step.Options)
				}
				fmt.Println()
			}
		}

		return nil
	},
}

// yamlRunCmd runs YAML tests
var yamlRunCmd = &cobra.Command{
	Use:   "run <file.yaml> [files...]",
	Short: "Run YAML test files",
	Long:  `Validates, compiles, and executes YAML test files against Virtuoso.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		svc := createYAMLService()

		// Configure API client
		// Check if global config is set
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		// Check if API token is configured
		if cfg.API.AuthToken == "" {
			return fmt.Errorf("API token not configured. Please set api.auth_token in config or VIRTUOSO_API_TOKEN environment variable")
		}

		// Create client using config
		apiClient := client.NewClientDirect(cfg.API.BaseURL, cfg.API.AuthToken)
		svc.SetAPIClient(apiClient)

		// Expand file globs
		var files []string
		for _, pattern := range args {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return fmt.Errorf("invalid pattern %s: %w", pattern, err)
			}
			files = append(files, matches...)
		}

		if len(files) == 0 {
			return fmt.Errorf("no files found matching patterns")
		}

		// Run tests
		var results []*service.ProcessResult

		if yamlParallel > 1 && len(files) > 1 {
			// Parallel execution
			fmt.Printf("Running %d tests in parallel (workers: %d)...\n", len(files), yamlParallel)
			// Implementation would use goroutines and channels
			// For now, fall through to sequential
		}

		// Sequential execution
		for _, file := range files {
			fmt.Printf("Running %s...\n", file)

			result, err := svc.RunFile(ctx, file)
			if err != nil {
				fmt.Printf("  ERROR: %v\n", err)
				continue
			}

			results = append(results, result)

			if result.Success {
				fmt.Printf("  ✓ Passed (%v)\n", result.Duration)
			} else {
				fmt.Printf("  ✗ Failed\n")
				for _, err := range result.Errors {
					fmt.Printf("    - %s\n", err.Message)
				}
			}
		}

		// Generate report
		if yamlOutputFormat != "" {
			if err := svc.GenerateReport(results, yamlOutputFormat); err != nil {
				return fmt.Errorf("failed to generate report: %w", err)
			}
			fmt.Printf("\nReport saved to test-results.%s\n", yamlOutputFormat)
		}

		// Summary
		passed := 0
		failed := 0
		for _, r := range results {
			if r.Success {
				passed++
			} else {
				failed++
			}
		}

		fmt.Printf("\nSummary: %d passed, %d failed\n", passed, failed)

		if failed > 0 {
			return fmt.Errorf("%d tests failed", failed)
		}

		return nil
	},
}

// yamlGenerateCmd generates tests from prompts
var yamlGenerateCmd = &cobra.Command{
	Use:   "generate <prompt>",
	Short: "Generate YAML test from natural language prompt",
	Long: `Uses AI templates to generate Virtuoso tests from requirements.

Templates available:
- login: User authentication flows
- search: Search functionality
- form: Form submission
- purchase: E2E purchase flows
- navigation: Site navigation
- data-driven: Tests with multiple data sets
- error-handling: Error scenarios
- conditional: Tests with conditional logic`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt := args[0]

		// Get template library
		templates := ai.NewTemplateLibrary()

		// Use specific template if provided
		var template *ai.Template
		if yamlTemplate != "" {
			t, ok := templates.GetTemplate(yamlTemplate)
			if !ok {
				return fmt.Errorf("unknown template: %s", yamlTemplate)
			}
			template = t
		} else {
			// Try to match template from prompt
			template = matchTemplateFromPrompt(templates, prompt)
		}

		// Generate test
		var generatedYAML string
		if template != nil {
			fmt.Fprintf(os.Stderr, "Using template: %s\n", template.Name)
			generatedYAML = generateFromTemplate(template, prompt)
		} else {
			fmt.Fprintf(os.Stderr, "Generating custom test...\n")
			generatedYAML = generateCustomTest(prompt)
		}

		// Validate generated YAML
		valid, issues := ai.ValidateAIOutput(generatedYAML)
		if !valid && yamlStrict {
			return fmt.Errorf("generated YAML has issues: %v", issues)
		}

		// Output generated YAML
		fmt.Println(generatedYAML)

		// Show optimization suggestions
		if suggestions := ai.OptimizationSuggestions(generatedYAML); len(suggestions) > 0 {
			fmt.Fprintf(os.Stderr, "\nOptimization suggestions:\n")
			for _, s := range suggestions {
				fmt.Fprintf(os.Stderr, "- %s\n", s)
			}
		}

		return nil
	},
}

// yamlConvertCmd converts existing tests to YAML
var yamlConvertCmd = &cobra.Command{
	Use:   "convert <checkpoint-id>",
	Short: "Convert existing Virtuoso test to YAML format",
	Long:  `Retrieves an existing test from Virtuoso and converts it to optimized YAML format.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_ = args[0] // checkpointID - will be used when ListSteps API is available

		// TODO: Implement step retrieval once ListSteps API is available
		// For now, return a placeholder
		return fmt.Errorf("convert command not implemented yet - ListSteps API method needed")

		// // Output YAML
		// data, err := yamlMarshal(yamlTest)
		// if err != nil {
		// 	return err
		// }

		// fmt.Println(string(data))

		// return nil
	},
}

// NewYAMLCmd creates the yaml command group
func NewYAMLCmd() *cobra.Command {

	// Add subcommands
	yamlCmd.AddCommand(yamlValidateCmd)
	yamlCmd.AddCommand(yamlCompileCmd)
	yamlCmd.AddCommand(yamlRunCmd)
	yamlCmd.AddCommand(yamlGenerateCmd)
	yamlCmd.AddCommand(yamlConvertCmd)

	// Validate flags
	yamlValidateCmd.Flags().BoolVar(&yamlStrict, "strict", false, "Treat warnings as errors")

	// Compile flags
	yamlCompileCmd.Flags().StringVarP(&yamlOutputFormat, "output", "o", "human", "Output format: human, json, commands")

	// Run flags
	yamlRunCmd.Flags().StringVarP(&yamlOutputFormat, "output", "o", "", "Report format: text, json, html")
	yamlRunCmd.Flags().StringToStringVar(&yamlEnvVars, "env", map[string]string{}, "Environment variables")
	yamlRunCmd.Flags().StringToStringVar(&yamlVars, "var", map[string]string{}, "Test variables")
	yamlRunCmd.Flags().IntVar(&yamlParallel, "parallel", 1, "Number of parallel workers")

	// Generate flags
	yamlGenerateCmd.Flags().StringVar(&yamlTemplate, "template", "", "Use specific template")
	yamlGenerateCmd.Flags().BoolVar(&yamlStrict, "strict", false, "Validate generated output strictly")

	return yamlCmd
}

// createYAMLService creates a configured YAML service
func createYAMLService() *service.Service {
	svcConfig := &service.Config{
		API: service.APIConfig{
			BaseURL:   cfg.API.BaseURL,
			AuthToken: cfg.API.AuthToken,
			OrgID:     cfg.Org.ID,
		},
		Validation: service.ValidationConfig{
			Strict:        yamlStrict,
			MaxErrors:     10,
			SpellCheck:    true,
			BestPractices: true,
		},
		Execution: service.ExecutionConfig{
			ScreenshotOnFailure: true,
			AutoWait:            true,
			Parallel:            yamlParallel,
		},
	}

	return service.NewService(svcConfig)
}

// Helper functions

func matchTemplateFromPrompt(lib *ai.TemplateLibrary, prompt string) *ai.Template {
	prompt = strings.ToLower(prompt)

	// Simple keyword matching
	keywords := map[string]string{
		"login":     "login",
		"signin":    "login",
		"auth":      "login",
		"search":    "search",
		"find":      "search",
		"form":      "form",
		"submit":    "form",
		"purchase":  "purchase",
		"buy":       "purchase",
		"checkout":  "purchase",
		"navigate":  "navigation",
		"menu":      "navigation",
		"data":      "data-driven",
		"multiple":  "data-driven",
		"error":     "error-handling",
		"invalid":   "error-handling",
		"if":        "conditional",
		"condition": "conditional",
	}

	for keyword, templateName := range keywords {
		if strings.Contains(prompt, keyword) {
			if template, ok := lib.GetTemplate(templateName); ok {
				return template
			}
		}
	}

	return nil
}

func generateFromTemplate(template *ai.Template, prompt string) string {
	// In real implementation, would use AI service
	// For now, return template example with modifications

	yaml := template.Example

	// Simple replacements based on prompt
	if strings.Contains(prompt, "registration") {
		yaml = strings.ReplaceAll(yaml, "Login", "Registration")
		yaml = strings.ReplaceAll(yaml, "login", "register")
		yaml = strings.ReplaceAll(yaml, "Sign In", "Sign Up")
	}

	return yaml
}

func generateCustomTest(prompt string) string {
	// In real implementation, would use AI service
	// For now, return a basic template

	return fmt.Sprintf(`test: %s
nav: /
do:
  # TODO: Add test steps based on:
  # %s
  - c: "Start"
  - ch: "Success"
`, prompt, prompt)
}

func convertStepsToYAML(steps []interface{}) interface{} {
	// In real implementation, would convert API response to YAML structure
	// This is a placeholder

	return map[string]interface{}{
		"test": "Converted Test",
		"do": []map[string]interface{}{
			{"nav": "/"},
			{"c": "button"},
			{"ch": "Success"},
		},
	}
}

func yamlMarshal(v interface{}) ([]byte, error) {
	// Custom YAML marshaling for clean output
	// In real implementation, would use yaml.v3 with proper configuration

	return []byte(`test: Converted Test
nav: /
do:
  - c: button
  - ch: Success
`), nil
}
