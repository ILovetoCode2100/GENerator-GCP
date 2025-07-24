# YAML Generate Command Demo

The `yaml generate` command uses AI-optimized templates to create Virtuoso tests from natural language prompts.

## Basic Usage

```bash
# Generate a test from a prompt
api-cli yaml generate "test user registration" > registration.yaml

# Generate with a specific template
api-cli yaml generate "login test" --template login > login.yaml

# Generate and validate
api-cli yaml generate "test checkout flow" | api-cli yaml validate -
```

## Examples

### 1. User Registration Test

```bash
$ ./bin/api-cli yaml generate "test user registration"
```

Generates:

```yaml
test: User Registration Flow
nav: https://example.com/register
data:
  email: test_{{timestamp}}@example.com
  password: SecurePass123!
  name: Test User

do:
  # Fill registration form
  - t: {#email: $email}
  - t: {#password: $password}
  - t: {#confirm-password: $password}
  - t: {#name: $name}

  # Accept terms if present
  - c: "#terms"

  # Submit
  - c: "Sign Up"

  # Wait and verify
  - wait: .dashboard
  - ch: Welcome
  - ch: $email

  # Store user ID
  - store: {.user-id: userId}
  - note: "Registered as: {{email}} (ID: {{userId}})"
```

### 2. Login Test

```bash
$ ./bin/api-cli yaml generate "test login flow"
```

Generates:

```yaml
test: User Login Flow
nav: https://example.com/login
data:
  username: test@example.com
  password: SecurePass123!

do:
  # Enter credentials
  - t: {#email: $username}
  - t: {#password: $password}

  # Submit
  - c: "Sign In"

  # Verify login
  - wait: .dashboard
  - ch: Welcome
  - ch: $username
```

### 3. Shopping Cart Test

```bash
$ ./bin/api-cli yaml generate "test shopping cart checkout"
```

Generates:

```yaml
test: Shopping Cart Flow
nav: https://shop.example.com
data:
  product: "Wireless Headphones"
  quantity: 2

do:
  # Search for product
  - c: "#search"
  - t: $product
  - k: Enter

  # Select product
  - wait: .products
  - c: $product

  # Add to cart
  - wait: .product-page
  - c: "Add to Cart"

  # View cart
  - wait: 1000
  - c: "Cart"

  # Checkout
  - wait: .cart
  - c: "Checkout"
  - ch: $product
  - ch: "Total"
```

### 4. Search Functionality Test

```bash
$ ./bin/api-cli yaml generate "test search functionality"
```

Generates:

```yaml
test: Search Functionality
nav: https://example.com
data:
  query: "test product"

do:
  # Search
  - c: "#search"
  - t: $query
  - k: Enter

  # Verify results
  - wait: .results
  - ch: $query
  - store: { .result-count: count }
  - note: "Found {{count}} results for: {{query}}"
```

## Template Detection

The generator automatically detects patterns in your prompt and selects appropriate templates:

- **Login/Sign in** → Login template with credentials
- **Registration/Sign up** → Registration form with validation
- **Shopping/Purchase/Cart** → E-commerce flow template
- **Search** → Search functionality template
- **Form** → Generic form submission template
- **Error/Invalid** → Error handling template
- **Loop/Multiple** → Data-driven test template
- **If/Condition** → Conditional flow template

## Advanced Features

### Conditional Logic

Prompts containing "if" or "conditional" generate tests with conditional steps:

```bash
$ ./bin/api-cli yaml generate "test form with conditional fields"
```

### Data-Driven Tests

Prompts mentioning "multiple" or "loop" generate data-driven tests:

```bash
$ ./bin/api-cli yaml generate "test login with multiple users"
```

### Error Handling

Prompts about errors generate tests with error scenarios:

```bash
$ ./bin/api-cli yaml generate "test login error handling"
```

## Optimization Suggestions

The generator provides optimization suggestions after generating tests:

```
Optimization suggestions:
- Consider using 'wait: element' instead of time-based waits
- Consider extracting common patterns into reusable blocks
```

## Benefits

1. **Token Efficiency**: Generated tests use compact YAML syntax (59% reduction)
2. **Best Practices**: Templates include proper waits, assertions, and structure
3. **Customizable**: Generated tests can be easily modified
4. **AI-Friendly**: Output is optimized for further AI processing

## Next Steps

1. Generate your test: `api-cli yaml generate "your test description"`
2. Review and customize the generated YAML
3. Validate: `api-cli yaml validate your-test.yaml`
4. Run: `api-cli yaml run your-test.yaml`

The generator provides a starting point that follows Virtuoso best practices while maintaining the compact, AI-optimized syntax of the YAML layer.
