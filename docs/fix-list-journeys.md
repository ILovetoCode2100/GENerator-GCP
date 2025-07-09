# Fix for ListJourneys API Response

## The Problem

Our current implementation expects:
```json
{
  "success": true,
  "items": [...]
}
```

But the actual API returns:
```json
{
  "success": true,
  "map": {
    "608093": {
      "journey": { ... },
      "lastChange": { ... }
    },
    "608094": {
      "journey": { ... },
      "lastChange": { ... }
    }
  }
}
```

## Fix Required

Update the `ListJourneys` function in `/pkg/virtuoso/client.go` to:

```go
func (c *Client) ListJourneys(goalID, snapshotID int) ([]*Journey, error) {
    var response struct {
        Success bool                          `json:"success"`
        Map     map[string]JourneyMapEntry   `json:"map"`
        Error   string                        `json:"error,omitempty"`
    }
    
    type JourneyMapEntry struct {
        Journey struct {
            ID          int      `json:"id"`
            SnapshotID  int      `json:"snapshotId"`
            GoalID      int      `json:"goalId"`
            Name        string   `json:"name"`
            Title       string   `json:"title"`
            CanonicalID string   `json:"canonicalId"`
            Draft       bool     `json:"draft"`
            Tags        []string `json:"tags"`
        } `json:"journey"`
        LastChange interface{} `json:"lastChange"`
    }
    
    resp, err := c.httpClient.R().
        SetQueryParam("snapshotId", fmt.Sprintf("%d", snapshotID)).
        SetQueryParam("goalId", fmt.Sprintf("%d", goalID)).
        SetQueryParam("includeSequencesDetails", "true").
        SetResult(&response).
        Get("/testsuites/latest_status")
    
    if err != nil {
        return nil, fmt.Errorf("list journeys request failed: %w", err)
    }
    
    if resp.IsError() {
        if response.Error != "" {
            return nil, fmt.Errorf("list journeys failed: %s", response.Error)
        }
        return nil, fmt.Errorf("list journeys failed with status %d: %s", resp.StatusCode(), resp.String())
    }
    
    // Convert map to slice of Journey pointers
    journeys := make([]*Journey, 0, len(response.Map))
    for _, entry := range response.Map {
        journey := &Journey{
            ID:          entry.Journey.ID,
            SnapshotID:  entry.Journey.SnapshotID,
            GoalID:      entry.Journey.GoalID,
            Name:        entry.Journey.Name,
            Title:       entry.Journey.Title,
            CanonicalID: entry.Journey.CanonicalID,
            Draft:       entry.Journey.Draft,
            Tags:        entry.Journey.Tags,
        }
        journeys = append(journeys, journey)
    }
    
    // Sort by ID to ensure consistent ordering
    sort.Slice(journeys, func(i, j int) bool {
        return journeys[i].ID < journeys[j].ID
    })
    
    return journeys, nil
}
```

## Key Changes

1. **Response Structure**: Changed from expecting `items` array to `map` object
2. **Added Query Parameter**: Added `includeSequencesDetails=true` 
3. **Proper Type Mapping**: Created `JourneyMapEntry` to match the nested structure
4. **Sorting**: Added sorting by ID to ensure first journey is the auto-created one

## Impact

This fix will:
- ✅ Properly detect the auto-created journey (608093 in your example)
- ✅ Show both journeys when listing
- ✅ Allow the batch structure command to find and rename the auto-created journey
- ✅ Fix the "0 journeys found" issue
