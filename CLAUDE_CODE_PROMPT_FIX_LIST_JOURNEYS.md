# Fix ListJourneys Implementation

Please fix the ListJourneys function in `/pkg/virtuoso/client.go` to handle the actual API response format.

**Current Issue**: The function expects an `items` array but the API returns a `map` object.

**Actual API Response Structure**:
```json
{
  "success": true,
  "map": {
    "608093": {
      "journey": {
        "id": 608093,
        "snapshotId": 43830,
        "goalId": 13807,
        "name": "Suite 1",
        "title": "First journey",
        "canonicalId": "b073b528-1dd3-480b-a34d-6933473b185c",
        "draft": false,
        "tags": []
      },
      "lastChange": { ... }
    }
  }
}
```

**Changes needed**:

1. Update the response struct to handle the `map` field instead of `items`
2. Add the missing query parameter `includeSequencesDetails=true`
3. Create proper type for the map entries
4. Convert the map to a slice of Journey pointers
5. Sort by ID to ensure consistent ordering (auto-created journey first)

**Also update the Journey struct** to include all fields:
- Title (string)
- CanonicalID (string) 
- Draft (bool)
- Tags ([]string)

Make sure to:
- Import the `sort` package if not already imported
- Handle the nested structure properly
- Keep error handling intact

Location: /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/pkg/virtuoso/client.go
