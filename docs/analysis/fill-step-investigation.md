# Possible FILL Step Formats to Try

## Current Implementation (returns 400):
```json
{
  "checkpointId": 1678326,
  "stepIndex": 999,
  "parsedStep": {
    "action": "FILL",
    "target": {
      "selectors": [
        {
          "type": "GUESS",
          "value": "{\"clue\":\"email\"}"
        }
      ]
    },
    "value": "test@example.com",
    "meta": {}
  }
}
```

## Alternative Formats to Try:

### Option 1: WRITE action instead of FILL
```json
{
  "action": "WRITE",
  "target": {
    "selectors": [
      {
        "type": "GUESS",
        "value": "{\"clue\":\"email\"}"
      }
    ]
  },
  "value": "test@example.com",
  "meta": {}
}
```

### Option 2: Different value structure
```json
{
  "action": "FILL",
  "target": {
    "selectors": [
      {
        "type": "GUESS",
        "value": "{\"clue\":\"email\"}"
      }
    ]
  },
  "value": "{\"text\":\"test@example.com\"}",
  "meta": {}
}
```

### Option 3: TYPE action
```json
{
  "action": "TYPE",
  "target": {
    "selectors": [
      {
        "type": "GUESS",
        "value": "{\"clue\":\"email\"}"
      }
    ]
  },
  "value": "test@example.com",
  "meta": {}
}
```

## To investigate:
1. Check if Postman collection has examples of successful FILL/WRITE/TYPE steps
2. Look for Virtuoso API documentation on step types
3. Try capturing a working FILL step from the Virtuoso UI
