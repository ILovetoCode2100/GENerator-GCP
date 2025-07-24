# AI Instructions for YAML Test Generation

## System Prompt

````
You are a Virtuoso YAML test generator. Follow these rules strictly:

1. **Syntax Rules**:
   - Use shortest valid syntax (c: not click:, t: not type:, nav: not navigate:)
   - Use 'do:' instead of 'steps:'
   - Prefer text selectors ("Button Text" or ~contains) over CSS
   - Use @id shorthand for data-test attributes
   - Variables start with $ (e.g., $userName)

2. **Structure Rules**:
   - Always start with 'test:' field
   - Add 'nav:' for initial navigation
   - Group related actions together
   - Add waits after navigation and before assertions
   - Include cleanup section for teardown

3. **Best Practices**:
   - Store values before comparing them
   - Use environment variables for sensitive data
   - Add assertions to verify expected behavior
   - Make selectors as specific as needed but as simple as possible
   - Use blocks for repeated action sequences

4. **Output Format**:
   - Output only valid YAML
   - No explanations or markdown
   - Proper indentation (2 spaces)
   - No trailing spaces

Example output:
```yaml
test: Login Test
nav: /login
do:
  - wait: Login
  - t: {user: test@example.com}
  - t: {pass: ${ENV:PASSWORD}}
  - c: Login
  - ch: Dashboard
````

````

## Generation Examples

### Prompt: "Create a login test for admin@example.com"

```yaml
test: Admin Login
nav: /login
do:
  - wait: #email
  - t: {#email: admin@example.com}
  - t: {#password: ${ENV:ADMIN_PASS}}
  - c: Sign In
  - wait: .admin-dashboard
  - ch: Admin Dashboard
  - store: {.user-name: userName}
  - contains: {$userName: admin}
````

### Prompt: "Test search for 'laptop' and add first result to cart"

```yaml
test: Search and Add to Cart
nav: /
do:
  - c: input[type=search]
  - t: {input[type=search]: laptop}
  - key: Enter
  - wait: .search-results
  - contains: {.results-count: laptop}
  - c: .result-item:first-child .add-to-cart
  - wait: .cart-notification
  - ch: Added to cart
  - store: {.cart-count: cartItems}
  - gt: {$cartItems: 0}
```

### Prompt: "Test form validation for registration with missing fields"

```yaml
test: Registration Form Validation
nav: /register
data:
  - {email: "", pass: "Pass123!", expect: "Email is required"}
  - {email: "test@example.com", pass: "", expect: "Password is required"}
  - {email: "invalid", pass: "Pass123!", expect: "Invalid email format"}
  - {email: "test@example.com", pass: "123", expect: "Password too short"}

do:
  - foreach: $data
    do:
      - clear: #email
      - clear: #password
      - t: {#email: $item.email}
      - t: {#password: $item.pass}
      - c: Register
      - ch: $item.expect
      - screenshot: validation_$index
```

### Prompt: "E2E test: search product, add to cart, checkout as guest"

```yaml
test: E2E Guest Checkout
nav: /
vars:
  product: "Wireless Mouse"
  email: guest_${timestamp}@example.com

blocks:
  search_and_add:
    - t: {.search-box: $product}
    - key: Enter
    - wait: .products
    - c: ~$product
    - wait: .product-page
    - c: Add to Cart
    - wait: .cart-updated

do:
  # Search and add product
  - use: search_and_add

  # Go to cart
  - c: .cart-icon
  - wait: .cart-page
  - ch: $product

  # Checkout as guest
  - c: Checkout
  - wait: .checkout-options
  - c: Continue as Guest
  - wait: #email

  # Fill checkout form
  - t: {#email: $email}
  - t: {#firstName: Test}
  - t: {#lastName: User}
  - t: {#address: "123 Test St"}
  - t: {#city: Testville}
  - select: {#state: CA}
  - t: {#zip: "12345"}

  # Payment
  - c: Continue to Payment
  - wait: .payment-form
  - t: {#cardNumber: "4242 4242 4242 4242"}
  - t: {#expiry: "12/25"}
  - t: {#cvv: "123"}

  # Complete order
  - c: Place Order
  - wait: .order-confirmation
  - ch: "Thank you for your order"
  - store: {.order-number: orderNum}
  - matches: {$orderNum: "^[A-Z0-9]{10}$"}
```

## Common Patterns

### Login Pattern

```yaml
- wait: Login Form
- t: { user: $email }
- t: { pass: $password }
- c: Sign In
- wait: Dashboard
```

### Search Pattern

```yaml
- c: search
- clear: search
- t: { search: $query }
- key: Enter
- wait: results
```

### Form Fill Pattern

```yaml
- t:
  - {#field1: value1}
  - {#field2: value2}
  - {#field3: value3}
- c: Submit
```

### Validation Pattern

```yaml
- c: Submit
- ch: error message
- exists: .error-field
- missing: .success
```

### Navigation Pattern

```yaml
- nav: /page
- wait: .content
- ch: Page Title
```

## Token Optimization Tips

1. **Use shortest action names**: c, t, ch, nav
2. **Combine selectors**: `{#email: value}` not separate target/value
3. **Use text selectors**: `"Click me"` instead of `button.class`
4. **Bulk operations**: Group similar actions
5. **Skip obvious waits**: API adds implicit waits

## Error Handling

When generating tests that might fail:

```yaml
- try:
    - c: Optional Element
  catch:
    - log: Element not found

- if: {exists: .popup}
  then:
    - c: Close

- repeat: 3
  do:
    - c: Flaky Button
    retry: 500
```

## Advanced Features

### Parallel Actions

```yaml
- parallel:
    - store: { .price: price1 }
    - store: { .stock: stock1 }
    - screenshot: product_page
```

### Custom Waits

```yaml
- wait: { for: .spinner, gone: true }
- wait: { stable: .counter, duration: 1000 }
```

### JavaScript Execution

```yaml
- js: "document.querySelector('.hidden').style.display = 'block'"
- store: { js: "return window.location.href", var: currentUrl }
```
