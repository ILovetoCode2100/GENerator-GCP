package ai

import (
	"fmt"
	"strings"
)

// InstructionSet provides comprehensive AI guidance
type InstructionSet struct {
	Purpose      string
	Guidelines   []string
	Examples     map[string]string
	Antipatterns map[string]string
}

// GetInstructions returns AI instructions for different scenarios
func GetInstructions(scenario string) *InstructionSet {
	switch scenario {
	case "generate":
		return getGenerationInstructions()
	case "optimize":
		return getOptimizationInstructions()
	case "fix":
		return getFixInstructions()
	case "convert":
		return getConversionInstructions()
	default:
		return getGeneralInstructions()
	}
}

// getGeneralInstructions provides general AI guidance
func getGeneralInstructions() *InstructionSet {
	return &InstructionSet{
		Purpose: "Generate efficient, maintainable Virtuoso tests using minimal tokens",
		Guidelines: []string{
			"Use the most compact syntax available",
			"Prefer text-based selectors over CSS/XPath",
			"Include assertions to verify test success",
			"Handle common edge cases (popups, slow loading)",
			"Use meaningful variable names",
			"Add comments only for complex logic",
			"Group related actions together",
			"Ensure tests are deterministic",
		},
		Examples: map[string]string{
			"minimal_click": `# Good - minimal tokens
- c: Login

# Avoid - verbose
- action: click
  target: "Login"`,

			"smart_selectors": `# Good - readable and stable
- c: "Add to Cart"
- c: @add-to-cart
- t: {#email: user@test.com}

# Avoid - brittle
- c: div.container > button:nth-child(3)
- c: //*[@id="btn-2847"]`,

			"efficient_typing": `# Good - combined selector and value
- t: {#email: test@example.com}
- t: {#password: secret123}

# Avoid - separate click and type
- c: #email
- t: test@example.com`,

			"smart_waits": `# Good - wait for specific element
- nav: /dashboard
- wait: .main-content

# Avoid - arbitrary waits
- nav: /dashboard
- wait: 3000`,
		},
		Antipatterns: map[string]string{
			"xpath_overuse":      "Avoid XPath unless absolutely necessary",
			"missing_assertions": "Always verify the action succeeded",
			"hardcoded_data":     "Use variables for test data",
			"no_error_handling":  "Consider what might go wrong",
		},
	}
}

// getGenerationInstructions for new test generation
func getGenerationInstructions() *InstructionSet {
	return &InstructionSet{
		Purpose: "Generate new Virtuoso tests from requirements",
		Guidelines: []string{
			"Start with clear test objective in the name",
			"Use descriptive but concise test names",
			"Include setup if needed (nav to start page)",
			"Follow logical user flow",
			"Add assertions after each major action",
			"Clean up if necessary (logout, clear data)",
			"Consider both happy path and edge cases",
			"Use data section for test inputs",
		},
		Examples: map[string]string{
			"complete_test": `test: User Registration Flow
nav: /register
data:
  email: test_{{timestamp}}@example.com
  password: SecurePass123!

do:
  # Fill registration form
  - t: {#email: $email}
  - t: {#password: $password}
  - t: {#confirm: $password}
  - c: @terms-checkbox
  - c: "Create Account"

  # Verify success
  - ch: "Welcome!"
  - ch: $email

  # Verify email sent
  - nav: /messages
  - ch: "Verify your email"`,

			"edge_cases": `test: Registration Edge Cases
nav: /register
do:
  # Test empty form
  - c: "Create Account"
  - ch: "Email is required"

  # Test invalid email
  - t: {#email: "not-an-email"}
  - c: "Create Account"
  - ch: "Invalid email"

  # Test password mismatch
  - t: {#email: valid@test.com}
  - t: {#password: Pass123!}
  - t: {#confirm: Different123!}
  - c: "Create Account"
  - ch: "Passwords don't match"`,
		},
		Antipatterns: map[string]string{
			"no_objective":    "Test name should indicate what's being tested",
			"no_verification": "Every test must verify its success",
			"assumes_state":   "Don't assume logged in/specific page",
			"ignores_errors":  "Handle predictable error cases",
		},
	}
}

// getOptimizationInstructions for improving existing tests
func getOptimizationInstructions() *InstructionSet {
	return &InstructionSet{
		Purpose: "Optimize existing tests for efficiency and reliability",
		Guidelines: []string{
			"Reduce token count without losing clarity",
			"Replace brittle selectors with stable ones",
			"Remove unnecessary waits",
			"Combine related actions",
			"Extract repeated patterns to blocks",
			"Use variables for repeated values",
			"Simplify complex selectors",
			"Remove redundant assertions",
		},
		Examples: map[string]string{
			"token_reduction": `# Before (45 tokens)
steps:
  - action: navigate
    url: https://example.com/login
  - action: click
    selector: "#username"
  - action: type
    text: "testuser"
  - action: click
    selector: "#password"
  - action: type
    text: "password123"
  - action: click
    selector: "button[type='submit']"

# After (19 tokens)
nav: /login
do:
  - t: {#username: testuser}
  - t: {#password: password123}
  - c: Submit`,

			"selector_improvement": `# Before - brittle
- c: body > div:nth-child(3) > form > button
- ch: div.success-message.green-text

# After - stable
- c: "Submit Order"
- ch: "Order placed successfully"`,

			"action_combining": `# Before - separate actions
- c: #email
- wait: 500
- t: user@test.com
- c: #password
- wait: 500
- t: pass123

# After - combined
- t: {#email: user@test.com}
- t: {#password: pass123}`,
		},
		Antipatterns: map[string]string{
			"over_optimization":  "Don't sacrifice readability for tokens",
			"removing_necessary": "Keep waits that prevent flakiness",
			"generic_selectors":  "Balance specificity and stability",
		},
	}
}

// getFixInstructions for fixing broken tests
func getFixInstructions() *InstructionSet {
	return &InstructionSet{
		Purpose: "Fix failing or flaky tests",
		Guidelines: []string{
			"Identify the root cause of failure",
			"Update changed selectors",
			"Add waits for async operations",
			"Handle dynamic content",
			"Make assertions more specific",
			"Add error handling for edge cases",
			"Ensure proper test isolation",
			"Fix timing issues",
		},
		Examples: map[string]string{
			"selector_fix": `# Element not found - selector changed
# Before
- c: .btn-primary

# After - use text or data attribute
- c: "Continue"
- c: @continue-button`,

			"timing_fix": `# Timing issue - element not ready
# Before
- nav: /dashboard
- c: .chart-container

# After - wait for element
- nav: /dashboard
- wait: .chart-container
- c: .chart-container`,

			"dynamic_content": `# Dynamic content - handle variations
# Before
- ch: "3 items"

# After - flexible assertion
- if:
    cond: exists("0 items")
    then:
      - note: "Cart is empty"
    else:
      - ch: "item"  # Just verify items exist`,

			"isolation_fix": `# Test pollution - ensure clean state
# Before
do:
  - nav: /cart
  - c: "Checkout"

# After - ensure cart has items
setup:
  - nav: /cart
  - if:
      cond: exists("Empty cart")
      then:
        - nav: /products
        - c: "Add to Cart"
        - nav: /cart

do:
  - c: "Checkout"`,
		},
		Antipatterns: map[string]string{
			"bandaid_fixes":    "Fix root cause, not symptoms",
			"excessive_waits":  "Don't just add waits everywhere",
			"ignoring_changes": "Update for UI changes properly",
		},
	}
}

// getConversionInstructions for converting other formats
func getConversionInstructions() *InstructionSet {
	return &InstructionSet{
		Purpose: "Convert tests from other formats to optimal YAML",
		Guidelines: []string{
			"Preserve test intent and coverage",
			"Use most compact equivalent syntax",
			"Improve selectors during conversion",
			"Add missing assertions",
			"Group related actions",
			"Extract variables for repeated values",
			"Remove unnecessary complexity",
			"Maintain test reliability",
		},
		Examples: map[string]string{
			"selenium_conversion": `# Selenium/WebDriver
driver.get("https://example.com/login")
driver.find_element(By.ID, "username").send_keys("testuser")
driver.find_element(By.ID, "password").send_keys("pass123")
driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
assert "Dashboard" in driver.title

# Converted to YAML
test: Login Test
nav: /login
do:
  - t: {#username: testuser}
  - t: {#password: pass123}
  - c: Submit
  - ch: Dashboard`,

			"playwright_conversion": `# Playwright
await page.goto('/search');
await page.fill('[placeholder="Search"]', 'laptop');
await page.press('[placeholder="Search"]', 'Enter');
await page.waitForSelector('.results');
const count = await page.locator('.result-item').count();
expect(count).toBeGreaterThan(0);

# Converted to YAML
test: Product Search
nav: /search
do:
  - t: {[placeholder="Search"]: laptop}
  - k: Enter
  - wait: .results
  - ch: .result-item`,

			"verbose_yaml": `# Verbose YAML
name: "Test checkout process"
steps:
  - step:
      action: "navigate"
      parameters:
        url: "https://example.com/cart"
  - step:
      action: "click"
      parameters:
        selector: "button.checkout-button"

# Optimized YAML
test: Checkout Process
nav: /cart
do:
  - c: Checkout`,
		},
		Antipatterns: map[string]string{
			"literal_translation": "Optimize during conversion",
			"keeping_waits":       "Remove unnecessary sync code",
			"complex_assertions":  "Simplify where possible",
		},
	}
}

// GenerateAIPrompt creates a comprehensive prompt for test generation
func GenerateAIPrompt(requirement string, context map[string]interface{}) string {
	prompt := strings.Builder{}

	// Base instructions
	prompt.WriteString(SystemPrompt)
	prompt.WriteString("\n\n")

	// Specific requirement
	prompt.WriteString("REQUIREMENT:\n")
	prompt.WriteString(requirement)
	prompt.WriteString("\n\n")

	// Context if provided
	if len(context) > 0 {
		prompt.WriteString("CONTEXT:\n")
		for key, value := range context {
			prompt.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
		}
		prompt.WriteString("\n")
	}

	// Generation guidelines
	instructions := getGenerationInstructions()
	prompt.WriteString("GUIDELINES:\n")
	for _, guideline := range instructions.Guidelines {
		prompt.WriteString(fmt.Sprintf("- %s\n", guideline))
	}
	prompt.WriteString("\n")

	// Output format
	prompt.WriteString("Generate a complete Virtuoso test in the compact YAML format.\n")
	prompt.WriteString("Ensure the test is self-contained, reliable, and uses minimal tokens.\n")
	prompt.WriteString("Include appropriate assertions and error handling.\n")

	return prompt.String()
}

// ValidateAIOutput checks if AI-generated YAML follows guidelines
func ValidateAIOutput(yaml string) (bool, []string) {
	issues := []string{}
	lines := strings.Split(yaml, "\n")

	// Check for required fields
	hasTest := false
	hasDo := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "test:") {
			hasTest = true
		}
		if strings.HasPrefix(trimmed, "do:") {
			hasDo = true
		}

		// Check for verbose syntax
		if strings.Contains(line, "action:") || strings.Contains(line, "selector:") {
			issues = append(issues, "Use compact syntax instead of verbose format")
		}

		// Check for inefficient patterns
		if strings.Contains(line, "wait: 1000") || strings.Contains(line, "wait: 2000") {
			if !strings.Contains(getPreviousLine(lines, line), "nav:") {
				issues = append(issues, "Avoid hardcoded waits; use element waits")
			}
		}

		// Check for poor selectors
		if strings.Contains(line, "> div >") || strings.Contains(line, "nth-child(") {
			issues = append(issues, "Use simpler, more stable selectors")
		}
	}

	if !hasTest {
		issues = append(issues, "Missing 'test:' field")
	}
	if !hasDo {
		issues = append(issues, "Missing 'do:' section")
	}

	// Check for assertions
	hasAssertion := false
	for _, line := range lines {
		if strings.Contains(line, "ch:") || strings.Contains(line, "eq:") {
			hasAssertion = true
			break
		}
	}
	if !hasAssertion {
		issues = append(issues, "Test should include assertions to verify success")
	}

	return len(issues) == 0, issues
}

// getPreviousLine helper to check context
func getPreviousLine(lines []string, currentLine string) string {
	for i, line := range lines {
		if line == currentLine && i > 0 {
			return lines[i-1]
		}
	}
	return ""
}

// OptimizationSuggestions provides specific optimization tips
func OptimizationSuggestions(yaml string) []string {
	suggestions := []string{}

	// Token counting approximation
	tokens := len(strings.Fields(yaml))
	if tokens > 100 {
		suggestions = append(suggestions, "Consider breaking into smaller tests or extracting common patterns")
	}

	// Check for repeated selectors
	selectorCount := make(map[string]int)
	lines := strings.Split(yaml, "\n")
	for _, line := range lines {
		if match := extractSelector(line); match != "" {
			selectorCount[match]++
		}
	}

	for selector, count := range selectorCount {
		if count > 3 {
			suggestions = append(suggestions, fmt.Sprintf("Selector '%s' used %d times - consider extracting to a variable", selector, count))
		}
	}

	// Check for optimization opportunities
	if strings.Count(yaml, "- c:") > 5 {
		suggestions = append(suggestions, "Multiple clicks might be combined into a more direct flow")
	}

	if strings.Contains(yaml, "wait:") && strings.Contains(yaml, "- ch:") {
		suggestions = append(suggestions, "Consider using 'wait: element' instead of time-based waits")
	}

	return suggestions
}

// extractSelector helper to find selectors in YAML
func extractSelector(line string) string {
	// Simple extraction - would be more sophisticated in production
	parts := strings.Split(strings.TrimSpace(line), ":")
	if len(parts) >= 2 {
		selector := strings.TrimSpace(parts[1])
		if strings.HasPrefix(selector, "#") || strings.HasPrefix(selector, ".") {
			return selector
		}
	}
	return ""
}
