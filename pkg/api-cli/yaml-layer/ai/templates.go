package ai

import (
	"fmt"
	"strings"
)

// Template represents an AI generation template
type Template struct {
	Name        string
	Description string
	Prompt      string
	Example     string
	Tags        []string
}

// TemplateLibrary holds all available templates
type TemplateLibrary struct {
	templates map[string]*Template
}

// NewTemplateLibrary creates a new template library
func NewTemplateLibrary() *TemplateLibrary {
	lib := &TemplateLibrary{
		templates: make(map[string]*Template),
	}
	lib.initializeTemplates()
	return lib
}

// GetTemplate retrieves a template by name
func (l *TemplateLibrary) GetTemplate(name string) (*Template, bool) {
	template, ok := l.templates[name]
	return template, ok
}

// GetTemplatesByTag retrieves templates matching a tag
func (l *TemplateLibrary) GetTemplatesByTag(tag string) []*Template {
	var matches []*Template
	for _, template := range l.templates {
		for _, t := range template.Tags {
			if strings.EqualFold(t, tag) {
				matches = append(matches, template)
				break
			}
		}
	}
	return matches
}

// ListTemplates returns all template names
func (l *TemplateLibrary) ListTemplates() []string {
	names := make([]string, 0, len(l.templates))
	for name := range l.templates {
		names = append(names, name)
	}
	return names
}

// initializeTemplates sets up the template library
func (l *TemplateLibrary) initializeTemplates() {
	// Login template
	l.templates["login"] = &Template{
		Name:        "login",
		Description: "User authentication flow",
		Prompt: `Generate a login test with:
- Navigate to login page
- Enter credentials
- Submit form
- Verify successful login`,
		Example: `test: User Login
nav: /login
do:
  - t: {#username: $user}
  - t: {#password: $pass}
  - c: Login
  - ch: Dashboard`,
		Tags: []string{"auth", "basic", "common"},
	}

	// Search template
	l.templates["search"] = &Template{
		Name:        "search",
		Description: "Search functionality test",
		Prompt: `Generate a search test with:
- Navigate to search
- Enter search term
- Submit search
- Verify results`,
		Example: `test: Product Search
do:
  - c: search
  - t: {#search: $query}
  - k: Enter
  - ch: .results
  - ch: $query`,
		Tags: []string{"search", "basic", "common"},
	}

	// Form submission template
	l.templates["form"] = &Template{
		Name:        "form",
		Description: "Form filling and submission",
		Prompt: `Generate a form test with:
- Fill all required fields
- Handle optional fields
- Submit form
- Verify submission`,
		Example: `test: Contact Form
nav: /contact
do:
  - t: {#name: John Doe}
  - t: {#email: john@example.com}
  - select: {#subject: Support}
  - t: {#message: Test message}
  - c: Submit
  - ch: "Thank you"`,
		Tags: []string{"form", "input", "common"},
	}

	// E2E purchase template
	l.templates["purchase"] = &Template{
		Name:        "purchase",
		Description: "End-to-end purchase flow",
		Prompt: `Generate a purchase test with:
- Product selection
- Add to cart
- Checkout process
- Payment
- Order confirmation`,
		Example: `test: E2E Purchase
nav: /products
do:
  # Select product
  - c: "Product Name"
  - c: Add to Cart
  - ch: "Added to cart"

  # Checkout
  - c: Cart
  - c: Checkout

  # Shipping
  - t: {#shipping-name: $name}
  - t: {#shipping-address: $address}
  - t: {#shipping-city: $city}
  - select: {#shipping-state: $state}
  - t: {#shipping-zip: $zip}
  - c: Continue

  # Payment
  - t: {#card-number: $cardNumber}
  - t: {#card-expiry: $expiry}
  - t: {#card-cvv: $cvv}
  - c: Place Order

  # Confirmation
  - ch: "Order Confirmed"
  - store: {.order-number: orderNum}`,
		Tags: []string{"e2e", "purchase", "complex"},
	}

	// Navigation template
	l.templates["navigation"] = &Template{
		Name:        "navigation",
		Description: "Site navigation test",
		Prompt: `Generate a navigation test with:
- Test main menu links
- Verify page loads
- Check breadcrumbs
- Test footer links`,
		Example: `test: Site Navigation
nav: /
do:
  # Main menu
  - c: Products
  - ch: .products-page
  - c: About
  - ch: .about-page
  - c: Contact
  - ch: .contact-page

  # Footer
  - scroll: bottom
  - c: Privacy Policy
  - ch: "Privacy"
  - nav: /`,
		Tags: []string{"navigation", "basic"},
	}

	// Data-driven template
	l.templates["data-driven"] = &Template{
		Name:        "data-driven",
		Description: "Data-driven test with multiple inputs",
		Prompt: `Generate a data-driven test with:
- Define test data
- Loop through data
- Execute same steps with different inputs
- Collect results`,
		Example: `test: Login Scenarios
data:
  users:
    - {email: valid@test.com, pass: valid123, expected: Dashboard}
    - {email: invalid@test.com, pass: wrong, expected: "Invalid credentials"}
    - {email: "", pass: "", expected: "Required field"}

nav: /login
do:
  - loop:
      over: $users
      as: user
      do:
        - t: {#email: $user.email}
        - t: {#password: $user.pass}
        - c: Login
        - ch: $user.expected
        - nav: /login  # Reset for next iteration`,
		Tags: []string{"data-driven", "advanced", "loop"},
	}

	// Error handling template
	l.templates["error-handling"] = &Template{
		Name:        "error-handling",
		Description: "Test with error handling",
		Prompt: `Generate a test with error handling:
- Try risky operations
- Handle potential failures
- Verify error messages
- Recovery actions`,
		Example: `test: Error Handling
nav: /form
do:
  # Submit empty form
  - c: Submit
  - ch: "Please fill required fields"

  # Invalid email
  - t: {#email: "not-an-email"}
  - c: Submit
  - ch: "Invalid email format"

  # Valid submission
  - t: {#email: valid@email.com}
  - t: {#name: Test User}
  - c: Submit
  - ch: Success`,
		Tags: []string{"error", "validation", "negative"},
	}

	// Conditional flow template
	l.templates["conditional"] = &Template{
		Name:        "conditional",
		Description: "Test with conditional logic",
		Prompt: `Generate a test with conditions:
- Check for element existence
- Branch based on conditions
- Different paths for different states`,
		Example: `test: Conditional Flow
nav: /dashboard
do:
  # Check if logged in
  - if:
      cond: exists(.logout-btn)
      then:
        - note: Already logged in
      else:
        - c: Login
        - t: {#user: $username}
        - t: {#pass: $password}
        - c: Submit

  # Dismiss optional popup
  - if:
      cond: exists(.popup)
      then:
        - c: .popup-close
        - wait: 500`,
		Tags: []string{"conditional", "advanced", "logic"},
	}

	// API integration template
	l.templates["api-setup"] = &Template{
		Name:        "api-setup",
		Description: "Test with API setup/teardown",
		Prompt: `Generate a test with API integration:
- Setup test data via API
- Perform UI test
- Verify via API
- Cleanup via API`,
		Example: `test: API Integration Test
setup:
  - js: |
      // Create test user via API
      fetch('/api/users', {
        method: 'POST',
        body: JSON.stringify({
          email: 'test@example.com',
          name: 'Test User'
        })
      }).then(r => r.json()).then(data => {
        window.testUserId = data.id;
      });
  - wait: 1000

do:
  - nav: /users
  - ch: test@example.com
  - c: test@example.com
  - ch: "Test User"

teardown:
  - js: |
      // Delete test user
      fetch('/api/users/' + window.testUserId, {
        method: 'DELETE'
      });`,
		Tags: []string{"api", "advanced", "setup"},
	}

	// Mobile responsive template
	l.templates["responsive"] = &Template{
		Name:        "responsive",
		Description: "Responsive design test",
		Prompt: `Generate a responsive test:
- Test desktop view
- Test tablet view
- Test mobile view
- Verify layout changes`,
		Example: `test: Responsive Design
nav: /
do:
  # Desktop
  - window: 1920x1080
  - ch: .desktop-menu
  - nch: .mobile-menu

  # Tablet
  - window: 768x1024
  - ch: .tablet-layout

  # Mobile
  - window: 375x667
  - nch: .desktop-menu
  - ch: .mobile-menu
  - c: ☰  # Hamburger menu
  - ch: .mobile-nav`,
		Tags: []string{"responsive", "mobile", "layout"},
	}

	// Accessibility template
	l.templates["accessibility"] = &Template{
		Name:        "accessibility",
		Description: "Accessibility testing",
		Prompt: `Generate an accessibility test:
- Test keyboard navigation
- Verify ARIA labels
- Check color contrast
- Test screen reader compatibility`,
		Example: `test: Accessibility Check
nav: /
do:
  # Keyboard navigation
  - k: Tab
  - ch: ":focus"
  - k: Tab
  - ch: 'a:focus'
  - k: Enter

  # ARIA labels
  - ch: '[aria-label="Main navigation"]'
  - ch: '[role="button"]'

  # Form labels
  - ch: 'label[for="email"]'
  - ch: '#email[aria-required="true"]'`,
		Tags: []string{"accessibility", "a11y", "compliance"},
	}

	// Performance template
	l.templates["performance"] = &Template{
		Name:        "performance",
		Description: "Performance monitoring test",
		Prompt: `Generate a performance test:
- Measure page load times
- Check resource loading
- Monitor user interactions
- Verify performance thresholds`,
		Example: `test: Performance Check
config:
  timeout: 30000

do:
  - js: "window.perfStart = Date.now()"
  - nav: /
  - wait: .main-content
  - js: |
      const loadTime = Date.now() - window.perfStart;
      if (loadTime > 3000) {
        throw new Error('Page load too slow: ' + loadTime + 'ms');
      }

  # Interaction performance
  - js: "window.clickStart = Date.now()"
  - c: Load More
  - wait: .new-content
  - js: |
      const responseTime = Date.now() - window.clickStart;
      console.log('Response time:', responseTime, 'ms');`,
		Tags: []string{"performance", "monitoring", "advanced"},
	}
}

// GeneratePrompt creates a system prompt for AI test generation
func GeneratePrompt(template *Template, inputs map[string]string) string {
	prompt := fmt.Sprintf(`You are generating a Virtuoso test using the compact YAML syntax.

RULES:
1. Use the shortest possible syntax (c: for click, t: for type, nav: for navigate)
2. Prefer text selectors over CSS when possible
3. Add waits only when necessary
4. Include meaningful assertions
5. Use variables for reusable values
6. Keep descriptions concise

TEMPLATE: %s
%s

EXAMPLE:
%s

INPUTS:
`, template.Name, template.Prompt, template.Example)

	for k, v := range inputs {
		prompt += fmt.Sprintf("- %s: %s\n", k, v)
	}

	prompt += `
Generate a complete test following this template and using the provided inputs.
Use the compact syntax shown in the example.`

	return prompt
}

// SystemPrompt provides the base system prompt for AI
const SystemPrompt = `You are an expert at creating Virtuoso tests using the compact YAML syntax.

SYNTAX REFERENCE:
- Navigation: nav: /path or nav: https://example.com
- Click: c: "Button Text" or c: #id or c: .class
- Type: t: "text" or t: {#id: value} or t: {selector: value}
- Key: k: Enter or k: Tab or k: Escape
- Hover: h: selector
- Select: select: {select#id: "Option Text"} or select: {selector: 3}
- Check exists: ch: selector or ch: "Text"
- Check not exists: nch: selector
- Equals: eq: {selector: "expected value"}
- Not equals: neq: {selector: "not this value"}
- Wait: wait: 1000 or wait: selector or wait: {for: selector, max: 5000}
- Store: store: {selector: varName}
- Scroll: scroll: top/bottom or scroll: selector or scroll: 100,200
- Window: window: maximize or window: 1024x768 or window: next
- Dialog: dialog: accept/dismiss or dialog: "text for prompt"
- JavaScript: js: "code"
- Note: note: "comment"
- If/then: if: {cond: "exists(selector)", then: [...], else: [...]}
- Loop: loop: {over: $array, as: item, do: [...]}

BEST PRACTICES:
1. Always use the shortest syntax possible
2. Prefer readable text selectors: c: "Login" over c: button.btn-primary
3. Use meaningful test names
4. Add assertions to verify success
5. Handle common scenarios (popups, slow loading)
6. Use variables for repeated values
7. Group related actions
8. Add notes for complex logic

COMMON PATTERNS:
- Login: t: {#user: $username}, t: {#pass: $password}, c: Login
- Search: c: search, t: $query, k: Enter
- Form: t: {#field1: value1}, select: {#dropdown: option}, c: Submit
- Navigation: c: "Menu Item", ch: .page-indicator
- Wait for load: nav: /page, wait: .main-content

Remember: Optimize for minimal tokens while maintaining clarity and completeness.`

// ValidateTemplate checks if generated YAML follows best practices
func ValidateTemplate(yaml string) []string {
	var issues []string

	// Check for verbose syntax
	if strings.Contains(yaml, "action:") || strings.Contains(yaml, "target:") {
		issues = append(issues, "Use compact syntax (c:, t:, nav:) instead of verbose format")
	}

	// Check for missing assertions
	if !strings.Contains(yaml, "ch:") && !strings.Contains(yaml, "eq:") {
		issues = append(issues, "Add assertions to verify test success")
	}

	// Check for hardcoded waits without reason
	lines := strings.Split(yaml, "\n")
	for i, line := range lines {
		if strings.Contains(line, "wait:") && strings.TrimSpace(line) == "- wait: 1000" {
			if i == 0 || !strings.Contains(lines[i-1], "nav:") {
				issues = append(issues, "Avoid hardcoded waits; use element waits instead")
			}
		}
	}

	// Check for overly specific selectors
	if strings.Contains(yaml, "div > div > div") || strings.Contains(yaml, "tbody > tr > td") {
		issues = append(issues, "Simplify overly specific selectors")
	}

	return issues
}
