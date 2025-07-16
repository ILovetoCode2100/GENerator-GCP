# Virtuoso MCP Server Tools Documentation

Generated on: 2025-07-16T03:58:48.973Z

## Overview

The Virtuoso MCP Server provides 12 tools across 12 command groups.

## Tool Groups

### Assert Commands (1 tools)

#### `virtuoso_assert`

Create assertion steps in Virtuoso tests. Supports various assertion types including element existence, value comparisons, and pattern matching.

**Subcommand:** `N/A`

### Data Commands (1 tools)

#### `virtuoso_data`

Manage data storage and cookies in Virtuoso tests. Store element text/values in variables and manage browser cookies.

**Subcommand:** `N/A`

### Dialog Commands (1 tools)

#### `virtuoso_dialog`

Handle browser dialogs in Virtuoso tests including alerts, confirms, and prompts.

**Subcommand:** `N/A`

### File Commands (1 tools)

#### `virtuoso_file_upload`

Upload a file in Virtuoso tests. Uploads a file from a URL to a file input element.

**Subcommand:** `N/A`

### Interact Commands (1 tools)

#### `virtuoso_interact`

Perform user interactions in Virtuoso tests including clicks, typing, hovering, and keyboard actions.

**Subcommand:** `N/A`

### Library Commands (1 tools)

#### `virtuoso_library`

Manage Virtuoso library checkpoints. Add checkpoints to library, get details, attach to journeys, and manage test steps within library checkpoints.

**Subcommand:** `N/A`

### Misc Commands (1 tools)

#### `virtuoso_misc`

Miscellaneous Virtuoso test actions including comments, script execution, and keyboard shortcuts.

**Subcommand:** `N/A`

### Mouse Commands (1 tools)

#### `virtuoso_mouse`

Perform mouse operations including movement, clicks, and viewport entry.

**Subcommand:** `N/A`

### Navigate Commands (1 tools)

#### `virtuoso_navigate`

Navigate to URLs and control page scrolling in Virtuoso tests.

**Subcommand:** `N/A`

### Select Commands (1 tools)

#### `virtuoso_select`

Select dropdown options in Virtuoso tests by text, value, index, or select the last option.

**Subcommand:** `N/A`

### Wait Commands (1 tools)

#### `virtuoso_wait`

Add wait conditions in Virtuoso tests. Wait for elements to appear/disappear or wait for a specific duration.

**Subcommand:** `N/A`

### Window Commands (1 tools)

#### `virtuoso_window`

Manage browser windows, tabs, and frames in Virtuoso tests.

**Subcommand:** `N/A`

## Usage Examples

### Assert Commands

```json
{
  "tool": "virtuoso_assert_exists",
  "arguments": {
    "checkpointId": "1680930",
    "selector": "Login button",
    "position": 1
  }
}
```

### Interact Commands

```json
{
  "tool": "virtuoso_interact_click",
  "arguments": {
    "checkpointId": "1680930",
    "selector": "Submit",
    "position": 1
  }
}
```

### Navigate Commands

```json
{
  "tool": "virtuoso_navigate_to",
  "arguments": {
    "checkpointId": "1680930",
    "url": "https://example.com",
    "position": 1
  }
}
```

## Integration with Claude Desktop

Add to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "virtuoso": {
      "command": "node",
      "args": ["/path/to/virtuoso-mcp-server/dist/index.js"]
    }
  }
}
```
