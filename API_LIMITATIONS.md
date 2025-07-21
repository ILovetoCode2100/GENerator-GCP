# Virtuoso API Limitations - Removed Commands

Based on HAR file analysis and testing, the following commands have been removed from the CLI due to API limitations:

## Removed Navigation Commands

- **navigate back** - API requires URL parameter, doesn't support browser back button (REMOVED)
- **navigate forward** - API requires URL parameter, doesn't support browser forward button (REMOVED)
- **navigate refresh** - API requires URL parameter, doesn't support browser refresh (REMOVED)

## Removed Window/Tab Operations

- **window close** - Not supported by API (REMOVED)
- **tab switch by index** - API actually supports TAB type with index value (WORKING)

## Removed Frame Operations

- **window switch frame-index** - Not supported (only FRAME_BY_ELEMENT works) (REMOVED)
- **window switch frame-name** - Not supported (only FRAME_BY_ELEMENT works) (REMOVED)
- **window switch main-content** - Not supported (no MAIN_CONTENT or DEFAULT_CONTENT type) (REMOVED)

## Updated File Operations

- **file upload** - Now only accepts URLs, not local file paths (UPDATED)
  - Previous error: "Invalid file URL /tmp/test.txt"
  - Both `file upload` and `file upload-url` now work with URLs

## Working Operations Confirmed

Based on HAR file evidence:

- NAVIGATE with URL (including useNewTab option)
- SWITCH with types: NEXT_TAB, PREV_TAB, PARENT_FRAME, FRAME_BY_ELEMENT
- WINDOW with type: RESIZE
- UPLOAD with URL values (not local files)

## API Request Structure Examples

### Working SWITCH operation:

```json
{
  "checkpointId": 1682010,
  "stepIndex": 5,
  "parsedStep": {
    "action": "SWITCH",
    "value": "",
    "meta": {
      "type": "NEXT_TAB"
    }
  }
}
```

### Working UPLOAD operation:

```json
{
  "checkpointId": 1682010,
  "stepIndex": 3,
  "parsedStep": {
    "action": "UPLOAD",
    "target": {
      "selectors": [
        {
          "type": "GUESS",
          "value": "{\"clue\":\"Upload Area\"}"
        }
      ]
    },
    "value": "https://example.com/file.pdf",
    "meta": {}
  }
}
```

### Working WINDOW RESIZE operation:

```json
{
  "checkpointId": 1682010,
  "stepIndex": 7,
  "parsedStep": {
    "action": "WINDOW",
    "value": "",
    "meta": {
      "type": "RESIZE",
      "dimension": {
        "width": 1024,
        "height": 768
      }
    }
  }
}
```

## Correct API Endpoints

### Library Operations

The following endpoints are correctly implemented in the CLI:

1. **Add checkpoint to library**:

   - Endpoint: `POST /api/testcases/{checkpointId}/add-to-library`
   - CLI: `api-cli library add {checkpointId}`

2. **Attach library checkpoint to journey**:
   - Endpoint: `POST /api/testsuites/{journeyId}/checkpoints/attach`
   - Body: `{"libraryCheckpointId": 7051, "position": 2}`
   - CLI: `api-cli library attach {journeyId} {libraryCheckpointId} {position}`

## Recommendations

1. Remove unsupported commands from CLI or mark them as deprecated
2. Update documentation to reflect these limitations
3. For file uploads, consider implementing a file hosting service or require users to provide URLs
4. For navigation commands, consider requiring URL parameters instead of trying browser navigation
