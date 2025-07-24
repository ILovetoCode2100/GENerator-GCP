# AI Templates and Generation Patterns

## AI Instructions for Test Generation

### System Prompt Template

```
You are a Virtuoso test generator. Create minimal YAML tests using these rules:

1. Use shortest valid syntax (c: not click:, t: not type:)
2. Prefer text selectors over CSS when possible
3. Add waits after navigation and before assertions
4. Group related actions logically
5. Store values before comparing them
6. Include cleanup steps

Output only valid YAML. No explanations.
```

### Common Test Patterns

#### 1. Login Flow Pattern

```yaml
# Template
pattern: login
inputs: [username, password, submit_button, success_indicator]
generate: |
  test: Login - {username}
  nav: {base_url}/login
  do:
    - wait: {submit_button}
    - t: {user: {username}}
    - t: {pass: {password}}
    - c: {submit_button}
    - wait: {success_indicator}
    - ch: {success_indicator}

# Example generation
test: Login - admin@example.com
nav: https://app.example.com/login
do:
  - wait: Login
  - t: {user: admin@example.com}
  - t: {pass: ${ENV:ADMIN_PASS}}
  - c: Login
  - wait: Dashboard
  - ch: Dashboard
```

#### 2. Form Submission Pattern

```yaml
pattern: form_submit
inputs: [form_fields, submit_action, success_check]
generate: |
  test: Submit {form_name}
  do:
    {{#each form_fields}}
    - t: {{selector}}: {{value}}
    {{/each}}
    - c: {submit_action}
    - wait: {success_check}
    - ch: {success_check}
```

#### 3. Shopping Cart Pattern

```yaml
pattern: shopping_cart
inputs: [product, quantity, checkout]
generate: |
  test: Add {product} to cart
  do:
    - c: ~{product}
    - wait: Add to cart
    - c: Add to cart
    - wait: .cart-count
    - store: {.cart-count: count}
    - equals: {$count: {quantity}}
    {{#if checkout}}
    - c: Checkout
    - wait: Payment
    {{/if}}
```

#### 4. Search Pattern

```yaml
pattern: search
inputs: [search_term, result_check]
generate: |
  test: Search for {search_term}
  do:
    - c: input[type=search]
    - clear: input[type=search]
    - t: {input[type=search]: {search_term}}
    - key: Enter
    - wait: .results
    - contains: {.results: {result_check}}
```

## Minimal Token Examples

### Ultra-Compact Syntax

```yaml
# Standard (88 tokens)
test: User Registration Test
navigate: https://example.com/register
steps:
  - click: "Sign Up"
  - type:
      selector: "#email"
      value: "test@example.com"
  - type:
      selector: "#password"
      value: "Test123!"
  - click: "Register"
  - check: "Welcome"

# Compact (49 tokens - 44% reduction)
test: User Registration
nav: /register
do:
  - c: Sign Up
  - t: {#email: test@example.com}
  - t: {#password: Test123!}
  - c: Register
  - ch: Welcome
```

### Batch Operations

```yaml
# Filling multiple fields (compact)
do:
  - t:
    - {#first: John}
    - {#last: Doe}
    - {#email: john@example.com}
    - {#phone: 555-1234}

# Even more compact with data
vars: [John, Doe, john@example.com, 555-1234]
do:
  - t: {#first: $1, #last: $2, #email: $3, #phone: $4}
```

## AI Generation Rules

### 1. Selector Priority

```yaml
# Priority order (most to least preferred)
rules:
  1_text: '"Exact Text"' or '~Contains'
  2_data_attr: '@test-id' or '[data-test=id]'
  3_semantic: 'button.primary', 'input[type=email]'
  4_id: '#unique-id'
  5_class: '.class-name'
  6_css: 'complex[selector*=value]'
  7_xpath: '//only/when/necessary'

# Examples
good:
  - c: Login            # Text
  - c: @submit-btn      # Data attribute
  - t: {input[type=email]: test@example.com}

avoid:
  - c: div.wrapper > div > button  # Too specific
  - c: //button[3]                 # Brittle XPath
```

### 2. Wait Strategy

```yaml
# Auto-insert waits
rules:
  after_nav: wait 2s or wait for element
  before_assert: wait for element to be stable
  after_action: wait if page might change

template: |
  - nav: {url}
  - wait: body          # After navigation
  - c: Submit
  - wait: .success      # Before assertion
  - ch: .success
```

### 3. Error Recovery

```yaml
# Defensive patterns
patterns:
  optional_element: |
    - try:
        - c: "Accept Cookies"
      catch:
        - log: "No cookie banner"

  retry_action: |
    - c: Submit
      retry: 3
      wait: 500

  conditional: |
    - if: {exists: .premium}
      then:
        - c: Premium Feature
      else:
        - c: Upgrade Now
```

## Complex Test Templates

### E2E Purchase Flow

```yaml
template: e2e_purchase
generate: |
  test: E2E Purchase - {product}
  vars:
    product: {product_name}
    user: test_{timestamp}@example.com

  do:
    # Search and select
    - use: search_product
      with: $product

    # Add to cart
    - c: Add to Cart
    - wait: .cart-popup
    - c: View Cart

    # Checkout
    - use: checkout_flow
      with:
        email: $user
        shipping: standard

    # Verify
    - wait: "Order Confirmation"
    - store: {.order-number: orderNum}
    - matches: {$orderNum: "^[A-Z0-9]{8}$"}

  cleanup:
    - nav: /logout
```

### Data-Driven Testing

```yaml
template: data_driven
generate: |
  test: Login Scenarios
  data:
    - {user: valid@example.com, pass: correct, expect: Dashboard}
    - {user: invalid@example.com, pass: wrong, expect: "Invalid credentials"}
    - {user: "", pass: "", expect: "Required fields"}

  do:
    - foreach: $data
      do:
        - nav: /login
        - t: {#email: $item.user}
        - t: {#pass: $item.pass}
        - c: Login
        - ch: $item.expect
        - screenshot: login_$index
```

## Generation Optimization

### Token Reduction Strategies

```yaml
strategies:
  1_shortcuts:
    before: "navigate:"
    after: "nav:"
    saves: 5 tokens per use

  2_implicit_waits:
    before: |
      - click: Submit
      - wait: .success
      - check: .success
    after: |
      - c: Submit
      - ch: .success  # Wait implied
    saves: 1 line per assertion

  3_combined_selectors:
    before: |
      - type:
          selector: #email
          value: test@example.com
    after: |
      - t: {#email: test@example.com}
    saves: 3 lines per input

  4_bulk_operations:
    before: |
      - store: {h1: title}
      - store: {.price: price}
      - store: {.stock: stock}
    after: |
      - store:
        - {h1: title}
        - {.price: price}
        - {.stock: stock}
    saves: 2 lines per group
```

### AI Instruction Optimizations

```yaml
# Minimal prompt for login test
prompt: "login test: user@example.com, pass123, check Dashboard"
generates: |
  test: Login
  nav: /login
  do:
    - t: {user: user@example.com}
    - t: {pass: pass123}
    - c: Login
    - ch: Dashboard

# Minimal prompt for form test
prompt: "form: name=John, email=john@example.com, submit, check Success"
generates: |
  test: Form
  do:
    - t: {name: John}
    - t: {email: john@example.com}
    - c: Submit
    - ch: Success
```

## AI Validation Feedback

```yaml
# Feedback template for corrections
correction_template: |
  Error: {error_type}
  Line {line}: {problematic_code}
  Fix: {corrected_code}
  Rule: {rule_violated}

# Example
feedback: |
  Error: Invalid selector
  Line 5: - click: button[
  Fix: - click: button[type=submit]
  Rule: Close all brackets in selectors
```
