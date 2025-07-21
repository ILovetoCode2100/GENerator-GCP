# Virtuoso API Limitations Discovered

Based on HAR file analysis and testing, the following API limitations have been identified:

## Navigation Commands

- **navigate back** - API requires URL parameter, doesn't support browser back button
- **navigate forward** - API requires URL parameter, doesn't support browser forward button
- **navigate refresh** - API requires URL parameter, doesn't support browser refresh

## Window/Tab Operations

- **window close** - Not supported by API
- **tab switch by index** - API only supports NEXT_TAB and PREV_TAB, not TAB_BY_INDEX

## Frame Operations

- **frame switch by index** - Not supported (only FRAME_BY_ELEMENT works)
- **frame switch by name** - Not supported (only FRAME_BY_ELEMENT works)
- **switch to main content** - Not supported (no MAIN_CONTENT or DEFAULT_CONTENT type)

## File Operations

- **file upload (local files)** - API only accepts URLs, not local file paths
  - Error: "Invalid file URL /tmp/test.txt"
  - Only `upload-url` with remote URLs works

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
