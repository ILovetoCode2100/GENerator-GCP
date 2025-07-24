package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/yaml-layer/core"
	"gopkg.in/yaml.v3"
)

// Validator performs multi-layer validation on YAML tests
type Validator struct {
	errors   []core.ValidationError
	warnings []core.ValidationError
	rules    []ValidationRule
}

// ValidationRule defines a validation check
type ValidationRule struct {
	Name  string
	Check func(*core.YAMLTest, *yaml.Node) []core.ValidationError
	Level string // "error" or "warning"
}

// NewValidator creates a validator with default rules
func NewValidator() *Validator {
	v := &Validator{
		errors:   []core.ValidationError{},
		warnings: []core.ValidationError{},
	}
	v.initializeRules()
	return v
}

// Validate performs all validation checks
func (v *Validator) Validate(yamlContent []byte) (bool, []core.ValidationError) {
	v.errors = []core.ValidationError{}
	v.warnings = []core.ValidationError{}

	// Parse YAML with node information for line numbers
	var node yaml.Node
	if err := yaml.Unmarshal(yamlContent, &node); err != nil {
		return false, []core.ValidationError{{
			Line:    1,
			Field:   "yaml",
			Message: fmt.Sprintf("Invalid YAML syntax: %v", err),
			Fix:     "Check YAML formatting, indentation, and special characters",
			Example: "test: My Test\nnav: /login\ndo:\n  - c: button",
		}}
	}

	// Parse into struct
	var test core.YAMLTest
	if err := node.Decode(&test); err != nil {
		return false, []core.ValidationError{{
			Line:    1,
			Field:   "structure",
			Message: fmt.Sprintf("Invalid test structure: %v", err),
			Fix:     "Ensure all fields match the schema",
		}}
	}

	// Run all validation rules
	for _, rule := range v.rules {
		errors := rule.Check(&test, &node)
		if rule.Level == "error" {
			v.errors = append(v.errors, errors...)
		} else {
			v.warnings = append(v.warnings, errors...)
		}
	}

	return len(v.errors) == 0, v.errors
}

// GetWarnings returns validation warnings
func (v *Validator) GetWarnings() []core.ValidationError {
	return v.warnings
}

// ValidateTest validates a parsed YAMLTest structure
func (v *Validator) ValidateTest(test *core.YAMLTest) (bool, []core.ValidationError) {
	v.errors = []core.ValidationError{}
	v.warnings = []core.ValidationError{}

	// Run all validation rules with nil node (structure-based validation only)
	for _, rule := range v.rules {
		errors := rule.Check(test, nil)
		if rule.Level == "error" {
			v.errors = append(v.errors, errors...)
		} else {
			v.warnings = append(v.warnings, errors...)
		}
	}

	return len(v.errors) == 0, v.errors
}

// initializeRules sets up all validation rules
func (v *Validator) initializeRules() {
	v.rules = []ValidationRule{
		// Required fields
		{
			Name:  "required-fields",
			Level: "error",
			Check: v.checkRequiredFields,
		},
		// Action validation
		{
			Name:  "action-syntax",
			Level: "error",
			Check: v.checkActionSyntax,
		},
		// Selector validation
		{
			Name:  "selector-format",
			Level: "warning",
			Check: v.checkSelectorFormat,
		},
		// Variable references
		{
			Name:  "variable-refs",
			Level: "error",
			Check: v.checkVariableReferences,
		},
		// Best practices
		{
			Name:  "best-practices",
			Level: "warning",
			Check: v.checkBestPractices,
		},
	}
}

// checkRequiredFields validates required fields
func (v *Validator) checkRequiredFields(test *core.YAMLTest, node *yaml.Node) []core.ValidationError {
	errors := []core.ValidationError{}

	if test.Test == "" {
		errors = append(errors, core.ValidationError{
			Line:    1,
			Field:   "test",
			Message: "Test name is required",
			Fix:     "Add a 'test:' field with a descriptive name",
			Example: "test: User Login Flow",
		})
	}

	if len(test.Do) == 0 {
		errors = append(errors, core.ValidationError{
			Line:    1,
			Field:   "do",
			Message: "Test must have at least one action in 'do' section",
			Fix:     "Add a 'do:' section with test steps",
			Example: "do:\n  - nav: /login\n  - c: Submit",
		})
	}

	return errors
}

// checkActionSyntax validates action syntax
func (v *Validator) checkActionSyntax(test *core.YAMLTest, node *yaml.Node) []core.ValidationError {
	errors := []core.ValidationError{}

	validateActions := func(actions []core.Action, section string) {
		for i, action := range actions {
			actionCount := 0
			actionType := ""

			// Count action types in this step
			if action.Nav != "" {
				actionCount++
				actionType = "nav"
			}
			if action.Scroll != "" {
				actionCount++
				actionType = "scroll"
			}
			if action.C != nil {
				actionCount++
				actionType = "c"
			}
			if action.T != nil {
				actionCount++
				actionType = "t"
			}
			if action.K != nil {
				actionCount++
				actionType = "k"
			}
			if action.Ch != nil {
				actionCount++
				actionType = "ch"
			}
			if action.Eq != nil {
				actionCount++
				actionType = "eq"
			}
			if action.Store != nil {
				actionCount++
				actionType = "store"
			}
			if action.Wait != nil {
				actionCount++
				actionType = "wait"
			}
			if action.If != nil {
				actionCount++
				actionType = "if"
			}
			if action.Loop != nil {
				actionCount++
				actionType = "loop"
			}

			if actionCount == 0 {
				errors = append(errors, core.ValidationError{
					Field:   fmt.Sprintf("%s[%d]", section, i),
					Message: "Action has no recognized operation",
					Fix:     "Add a valid action like 'c:', 't:', 'nav:', etc.",
					Example: "- c: button.submit",
				})
			} else if actionCount > 1 {
				errors = append(errors, core.ValidationError{
					Field:   fmt.Sprintf("%s[%d]", section, i),
					Message: "Action has multiple operations",
					Fix:     "Use only one action type per step",
					Example: fmt.Sprintf("Split into:\n  - %s: ...\n  - next_action: ...", actionType),
				})
			}

			// Validate specific action formats
			v.validateActionFormat(&action, fmt.Sprintf("%s[%d]", section, i), &errors)
		}
	}

	if len(test.Setup) > 0 {
		validateActions(test.Setup, "setup")
	}
	validateActions(test.Do, "do")
	if len(test.Teardown) > 0 {
		validateActions(test.Teardown, "teardown")
	}

	return errors
}

// validateActionFormat checks specific action format requirements
func (v *Validator) validateActionFormat(action *core.Action, path string, errors *[]core.ValidationError) {
	// Type action validation
	if action.T != nil {
		switch t := action.T.(type) {
		case string:
			// Simple string format is valid
		case map[interface{}]interface{}:
			if len(t) != 1 {
				*errors = append(*errors, core.ValidationError{
					Field:   path + ".t",
					Message: "Type action map must have exactly one selector:value pair",
					Fix:     "Use format: t: {selector: value}",
					Example: "t: {'#email': 'user@example.com'}",
				})
			}
		case map[string]interface{}:
			if len(t) != 1 {
				*errors = append(*errors, core.ValidationError{
					Field:   path + ".t",
					Message: "Type action map must have exactly one selector:value pair",
					Fix:     "Use format: t: {selector: value}",
					Example: "t: {'#email': 'user@example.com'}",
				})
			}
		default:
			*errors = append(*errors, core.ValidationError{
				Field:   path + ".t",
				Message: "Invalid type action format",
				Fix:     "Use string or {selector: value} format",
				Example: "t: 'text' or t: {'#input': 'value'}",
			})
		}
	}

	// Click action validation
	if action.C != nil {
		switch c := action.C.(type) {
		case string:
			// Valid
		case map[interface{}]interface{}:
			// Check for valid options
			for k := range c {
				key := fmt.Sprintf("%v", k)
				if key != "pos" && key != "var" {
					*errors = append(*errors, core.ValidationError{
						Field:   path + ".c",
						Message: fmt.Sprintf("Unknown click option: %s", key),
						Fix:     "Valid options are 'pos' and 'var'",
						Example: "c: {button: {pos: center, var: btnText}}",
					})
				}
			}
		default:
			*errors = append(*errors, core.ValidationError{
				Field:   path + ".c",
				Message: "Invalid click format",
				Fix:     "Use string or map with options",
				Example: "c: 'button' or c: {button: {pos: center}}",
			})
		}
	}

	// Wait validation
	if action.Wait != nil {
		switch w := action.Wait.(type) {
		case int, float64:
			// Valid numeric wait
		case string:
			// Should be a selector
		case map[interface{}]interface{}:
			if _, hasFor := w["for"]; !hasFor {
				*errors = append(*errors, core.ValidationError{
					Field:   path + ".wait",
					Message: "Wait with options must include 'for' field",
					Fix:     "Add 'for' field with selector",
					Example: "wait: {for: '.loaded', max: 5000}",
				})
			}
		default:
			*errors = append(*errors, core.ValidationError{
				Field:   path + ".wait",
				Message: "Invalid wait format",
				Fix:     "Use number (ms), selector, or {for: selector, max: ms}",
				Example: "wait: 1000 or wait: '.loaded'",
			})
		}
	}
}

// checkSelectorFormat validates CSS selector format
func (v *Validator) checkSelectorFormat(test *core.YAMLTest, node *yaml.Node) []core.ValidationError {
	warnings := []core.ValidationError{}
	selectorRegex := regexp.MustCompile(`^[#\.\[\]a-zA-Z0-9\-_:\s\*>+~="'\(\)]+$`)

	checkSelector := func(selector string, path string) {
		if selector == "" {
			return
		}

		// Skip variable references
		if strings.Contains(selector, "{{") {
			return
		}

		if !selectorRegex.MatchString(selector) {
			warnings = append(warnings, core.ValidationError{
				Field:   path,
				Message: fmt.Sprintf("Unusual selector format: %s", selector),
				Fix:     "Verify selector is valid CSS",
				Example: "#id, .class, button[type='submit']",
			})
		}

		// Check for common issues
		if strings.Count(selector, "'") == 1 || strings.Count(selector, "\"") == 1 {
			warnings = append(warnings, core.ValidationError{
				Field:   path,
				Message: "Unclosed quote in selector",
				Fix:     "Ensure quotes are properly paired",
				Example: "button[data-test='value']",
			})
		}
	}

	// Check selectors in actions
	checkActionsForSelectors := func(actions []core.Action, section string) {
		for i, action := range actions {
			basePath := fmt.Sprintf("%s[%d]", section, i)

			// Check string selectors
			if s, ok := action.C.(string); ok {
				checkSelector(s, basePath+".c")
			}
			if s, ok := action.Ch.(string); ok {
				checkSelector(s, basePath+".ch")
			}
			if action.H != "" {
				checkSelector(action.H, basePath+".h")
			}

			// Check map selectors
			if m, ok := action.T.(map[interface{}]interface{}); ok {
				for k := range m {
					checkSelector(fmt.Sprintf("%v", k), basePath+".t")
				}
			} else if m, ok := action.T.(map[string]interface{}); ok {
				for k := range m {
					checkSelector(k, basePath+".t")
				}
			}
		}
	}

	checkActionsForSelectors(test.Do, "do")

	return warnings
}

// checkVariableReferences validates variable usage
func (v *Validator) checkVariableReferences(test *core.YAMLTest, node *yaml.Node) []core.ValidationError {
	errors := []core.ValidationError{}
	definedVars := make(map[string]bool)

	// Pre-populate with data variables
	for k := range test.Data {
		definedVars[k] = true
	}

	// Track variable definitions and usage
	checkActions := func(actions []core.Action, section string) {
		for i, action := range actions {
			basePath := fmt.Sprintf("%s[%d]", section, i)

			// Check store actions for definitions
			if action.Store != nil {
				if m, ok := action.Store.(map[interface{}]interface{}); ok {
					for _, v := range m {
						if varName, ok := v.(string); ok {
							definedVars[varName] = true
						}
					}
				}
			}

			// Check for variable usage in strings
			checkVarUsage := func(s string, path string) {
				varPattern := regexp.MustCompile(`\{\{(\w+)\}\}`)
				matches := varPattern.FindAllStringSubmatch(s, -1)
				for _, match := range matches {
					varName := match[1]
					if !definedVars[varName] {
						errors = append(errors, core.ValidationError{
							Field:   path,
							Message: fmt.Sprintf("Undefined variable: %s", varName),
							Fix:     "Define variable with 'store' action or in 'data' section",
							Example: fmt.Sprintf("store: {'#element': '%s'} or data: {%s: value}", varName, varName),
						})
					}
				}
			}

			// Check all string fields for variable references
			if action.Nav != "" {
				checkVarUsage(action.Nav, basePath+".nav")
			}
			if s, ok := action.T.(string); ok {
				checkVarUsage(s, basePath+".t")
			}
		}
	}

	if len(test.Setup) > 0 {
		checkActions(test.Setup, "setup")
	}
	checkActions(test.Do, "do")

	return errors
}

// checkBestPractices validates against best practices
func (v *Validator) checkBestPractices(test *core.YAMLTest, node *yaml.Node) []core.ValidationError {
	warnings := []core.ValidationError{}

	// Check test name
	if len(test.Test) > 80 {
		warnings = append(warnings, core.ValidationError{
			Field:   "test",
			Message: "Test name is very long",
			Fix:     "Consider a shorter, more concise name",
		})
	}

	// Check for explicit waits after navigation
	hasNavWithoutWait := false
	for i, action := range test.Do {
		if action.Nav != "" && i < len(test.Do)-1 {
			nextAction := test.Do[i+1]
			if nextAction.Wait == nil && nextAction.Ch == nil {
				hasNavWithoutWait = true
				break
			}
		}
	}

	if hasNavWithoutWait {
		warnings = append(warnings, core.ValidationError{
			Field:   "do",
			Message: "Navigation without wait may cause timing issues",
			Fix:     "Consider adding 'wait' or 'ch' after navigation",
			Example: "- nav: /page\n- wait: '.loaded'",
		})
	}

	// Check for hardcoded credentials
	checkForCredentials := func(s string) bool {
		lowerStr := strings.ToLower(s)
		return strings.Contains(lowerStr, "password") ||
			strings.Contains(lowerStr, "secret") ||
			strings.Contains(lowerStr, "token") ||
			strings.Contains(lowerStr, "api_key")
	}

	// Check data section
	for k, v := range test.Data {
		if checkForCredentials(k) && v != "{{ENV_VAR}}" {
			warnings = append(warnings, core.ValidationError{
				Field:   "data." + k,
				Message: "Possible hardcoded credential",
				Fix:     "Use environment variables or external config",
				Example: fmt.Sprintf("data: {%s: '{{%s}}'}", k, strings.ToUpper(k)),
			})
		}
	}

	return warnings
}
