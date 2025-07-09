# Batch Structure Implementation Plan

## Key Virtuoso API Behaviors to Handle

### 1. Auto-Created Journey Management
- When creating a goal, Virtuoso automatically creates an initial journey
- This journey should be **renamed and reused** as the first journey in our structure
- Don't create a duplicate - work with what Virtuoso provides

### 2. First Checkpoint Navigation Rules
- The first checkpoint in a goal is **always** a navigation step
- This navigation step is **shared/synced across the goal**
- If users specify a navigation step for the first checkpoint, **update the existing one**
- Don't create a new navigation step - modify the auto-created one

### 3. New API Endpoints Required

#### Update Journey Name
```bash
PUT /api/testsuites/{journey_id}
Body: {
  "name": "Updated Journey Name"
}
```

#### Get Step Details
```bash
GET /api/teststeps/{step_id}
# Returns step details including canonicalId needed for updates
```

#### Update Navigation Step
```bash
PUT /api/teststeps/{step_id}/properties
Body: {
  "id": {step_id},
  "canonicalId": "{canonical_id}",
  "action": "NAVIGATE",
  "value": "https://www.example.com",
  "meta": {
    "kind": "NAVIGATE",
    "type": "URL",
    "value": "https://www.example.com",
    "url": "https://www.example.com",
    "useNewTab": false
  },
  "optional": false,
  "ignoreOutcome": false,
  "skip": false
}
```

## Implementation Flow

### Phase 1: Create Project & Goal
```go
// 1. Create project (if needed)
projectID := createProject(structure.Project.Name)

// 2. Create goal - this auto-creates initial journey
goalResult := createGoal(projectID, structure.Goals[0].Name, structure.Goals[0].URL)
goalID := goalResult.GoalID
snapshotID := goalResult.SnapshotID
autoJourneyID := goalResult.JourneyID // The auto-created journey
```

### Phase 2: Handle Auto-Created Journey
```go
// 3. Rename the auto-created journey to match our first journey
if len(structure.Goals[0].Journeys) > 0 {
    updateJourneyName(autoJourneyID, structure.Goals[0].Journeys[0].Name)
    
    // Use this journey ID for the first journey's checkpoints
    firstJourneyID = autoJourneyID
}
```

### Phase 3: Handle First Checkpoint Navigation
```go
// 4. Get the auto-created first checkpoint
checkpoints := listCheckpoints(firstJourneyID)
firstCheckpointID := checkpoints[0].ID

// 5. Get the navigation step details
steps := getCheckpointSteps(firstCheckpointID)
navStep := steps[0] // First step is always navigation
navStepDetails := getStepDetails(navStep.ID)

// 6. Update navigation if user specified different URL
if userSpecifiedNavigation {
    updateNavigationStep(navStep.ID, navStepDetails.CanonicalID, newURL)
}
```

### Phase 4: Create Additional Resources
```go
// 7. Create additional journeys (skip first since we reused auto-created)
for i := 1; i < len(structure.Goals[0].Journeys); i++ {
    journeyID := createJourney(goalID, snapshotID, journey.Name)
    // Process checkpoints for this journey
}

// 8. Create checkpoints (skip first for first journey)
// 9. Add steps (skip navigation for first checkpoint of first journey)
```

## Structure Format Enhancement

```yaml
project:
  name: "E-Commerce Tests"
  # Optional: specify existing project ID to skip creation
  id: 9056
  
goals:
  - name: "Checkout Flow"
    url: "https://shop.example.com"
    journeys:
      - name: "Guest Checkout"  # This will rename the auto-created journey
        checkpoints:
          - name: "Browse Products"
            # Navigation URL will update the existing nav step
            navigation_url: "https://shop.example.com/products"
            steps:
              # Non-navigation steps only
              - type: click
                selector: ".product-card"
              - type: wait
                selector: ".product-details"
                timeout: 5000
          - name: "Add to Cart"
            # No navigation_url = keep existing navigation
            steps:
              - type: click
                selector: ".add-to-cart"
              - type: wait
                selector: ".cart-updated"
```

## Error Handling

1. **Journey Rename Failure**: Log warning but continue
2. **Navigation Update Failure**: Critical - stop execution
3. **Missing Auto-Resources**: Verify goal creation included journey
4. **Duplicate Navigation**: Warn user, use update instead of create

## Validation Rules

1. First journey in structure → reuses auto-created journey
2. First checkpoint navigation → updates existing, never creates
3. Additional journeys → created normally
4. Additional checkpoints → created with standard attachment

## Command Usage

```bash
# Basic usage
./bin/api-cli create-structure --file test-suite.yaml

# Dry run mode
./bin/api-cli create-structure --file test-suite.yaml --dry-run

# Use existing project
./bin/api-cli create-structure --file test-suite.yaml --project-id 9056

# Verbose output
./bin/api-cli create-structure --file test-suite.yaml --verbose
```

## Success Output Example

```
Creating test structure from: test-suite.yaml

✓ Created project: E-Commerce Tests (ID: 9061)
✓ Created goal: Checkout Flow (ID: 13782)
  ✓ Retrieved snapshot ID: 43808
  ✓ Renamed auto-created journey to: Guest Checkout (ID: 608048)
  
Processing Guest Checkout journey:
  ✓ Updated navigation URL for first checkpoint
  ✓ Added 2 steps to checkpoint: Browse Products
  ✓ Created checkpoint: Add to Cart (ID: 1678325)
  ✓ Added 2 steps to checkpoint: Add to Cart

Summary:
- Project ID: 9061
- Goals created: 1
- Journeys processed: 1 (1 renamed, 0 created)
- Checkpoints: 2 (1 existing, 1 created)
- Steps added: 4
- Navigation updates: 1

Total time: 4.2s
```
