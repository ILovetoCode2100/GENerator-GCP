# Requirements to Virtuoso NLP Test Mapping

## Project Details

- **Project**: Rocketshop Requirements Test (ID: 9411)
- **Goal**: Validate E-commerce Requirements (ID: 14583)
- **Journey**: End-to-End Purchase Requirements (ID: 611015)

## Requirements Traceability Matrix

### Checkpoint 1: Access and Product Selection (ID: 1683919)

| Requirement                                                                                                               | NLP Test Steps                                                                            | Status    |
| ------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- | --------- |
| **REQ-001**: The system shall make the Rocketshop storefront accessible via the public URL https://rocketshop.virtuoso.qa | 1. Navigate to "https://rocketshop.virtuoso.qa"<br>2. Verify that "Rocketshop" is visible | ✓ Created |
| **REQ-002**: The system shall allow users to add at least one product to the shopping bag directly from the landing page  | 3. Click on "Add to Bag"                                                                  | ✓ Created |

### Checkpoint 2: Shopping Bag Navigation (ID: 1683920)

| Requirement                                                                                                               | NLP Test Steps                                                             | Status    |
| ------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------- | --------- |
| **REQ-003**: The system shall allow users to navigate to the shopping bag from the main navigation after adding a product | 1. Click on "Shopping Bag"                                                 | ✓ Created |
| **REQ-004**: The system shall display the correct Shopping Bag page header when accessed                                  | 2. Verify that the heading "Shopping Bag" is visible                       | ✓ Created |
| **REQ-005**: The system shall provide a button to proceed to checkout from the Shopping Bag page                          | 3. Verify that "Go to Checkout" is visible<br>4. Click on "Go to Checkout" | ✓ Created |

### Checkpoint 3: Checkout Form Validation (ID: 1683921)

| Requirement                                                                                                                                                   | NLP Test Steps                                                                                                                                                                                                                                               | Status    |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------- |
| **REQ-006**: The checkout process shall require the following fields:<br>- Full name<br>- Email address<br>- Shipping address<br>- Phone number<br>- ZIP code | 1. Enter "John Smith" in the "Full name" field<br>2. Enter "john.smith@example.com" in the "Email" field<br>3. Enter "456 Oak Avenue" in the "Address" field<br>4. Enter "555-9876" in the "Phone numbers" field<br>5. Enter "94105" in the "ZIP code" field | ✓ Created |

### Checkpoint 4: Payment and Confirmation (ID: 1683922)

| Requirement                                                                                                                               | NLP Test Steps                                                                                                   | Status    |
| ----------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- | --------- |
| **REQ-007**: The payment process shall require the following fields:<br>- Card number<br>- Card security code (CVV)                       | 1. Enter "4242 4242 4242 4242" in the "Card number" field<br>2. Enter "123" in the CVV field (placeholder "xxx") | ✓ Created |
| **REQ-008**: The system shall accept a payment and proceed with the checkout when valid card details are entered                          | 3. Click on "Confirm and Pay"                                                                                    | ✓ Created |
| **REQ-009**: Upon successful payment, the system shall display a Purchase Confirmed message to the user                                   | 4. Wait for processing<br>5. Verify that "Purchase Confirmed!" is visible                                        | ✓ Created |
| **REQ-010**: The system shall provide the option for the user to download a purchase confirmation or receipt from the confirmation screen | 6. Verify that "Download Confirmation" is visible                                                                | ✓ Created |

## Test Coverage Summary

- **Total Requirements**: 10
- **Requirements Covered**: 10 (100%)
- **Total Test Steps Created**: 20
- **Checkpoints Created**: 4

## NLP Test Advantages

1. **Natural Language**: All test steps use plain English commands that match what users see on screen
2. **Requirements Traceability**: Each requirement is directly mapped to specific test steps
3. **Maintainability**: Tests can be updated by modifying the text that appears on screen, not technical selectors
4. **Readability**: Non-technical stakeholders can understand and validate test coverage

## Execution

The complete test suite can be executed via:

- Virtuoso Web Platform: https://app.virtuoso.qa/#/project/9411
- Individual checkpoint execution through the platform
- Full journey execution to validate all requirements in sequence

## Notes

- All test steps use NLP patterns focusing on visible text and labels
- Test data uses realistic values (e.g., "John Smith", valid credit card test number)
- Wait steps are included where necessary for page processing
- Each checkpoint groups related requirements for logical test organization
