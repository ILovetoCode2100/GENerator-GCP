# Manual Testing Guide for Virtuoso MCP Server

## Overview

This guide provides step-by-step instructions for manually testing the Virtuoso MCP Server integration with Claude Desktop and validating all tool functionality.

## Prerequisites

1. **Virtuoso API Configuration**

   - Valid API key from Virtuoso
   - Configuration file at `~/.api-cli/virtuoso-config.yaml`:

   ```yaml
   api:
     auth_token: your-api-key-here
     base_url: https://api-app2.virtuoso.qa/api
   organization:
     id: "2242"
   headers:
     X-Virtuoso-Client-ID: "api-cli-generator"
     X-Virtuoso-Client-Name: "api-cli-generator"
   ```

2. **Claude Desktop** (version 0.7.0 or higher)

   - Download from: https://claude.ai/download
   - Ensure MCP support is enabled

3. **Node.js** (version 18 or higher)

   - Verify with: `node --version`

4. **Built MCP Server**
   - Run: `npm run build` or `./scripts/build.sh`

## Setup Instructions

### 1. Build the Server

```bash
cd /path/to/virtuoso-mcp-server
npm install
npm run build
```

### 2. Configure Claude Desktop

1. Open Claude Desktop settings
2. Navigate to Developer → Model Context Protocol
3. Add new MCP server configuration:

```json
{
  "mcpServers": {
    "virtuoso": {
      "command": "node",
      "args": ["/absolute/path/to/virtuoso-mcp-server/dist/index.js"],
      "env": {
        "NODE_ENV": "production"
      }
    }
  }
}
```

**macOS config location**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows config location**: `%APPDATA%\Claude\claude_desktop_config.json`
**Linux config location**: `~/.config/Claude/claude_desktop_config.json`

### 3. Restart Claude Desktop

After updating the configuration, completely quit and restart Claude Desktop.

### 4. Verify MCP Connection

In Claude Desktop, you should see:

- An MCP indicator in the chat interface
- "virtuoso" listed as a connected server
- Access to Virtuoso tools when typing "/"

## Test Scenarios

### Scenario 1: Basic Connectivity

**Test**: Verify MCP server is connected

1. Open Claude Desktop
2. Start a new conversation
3. Type: "What MCP tools are available?"
4. **Expected**: Claude lists Virtuoso tools grouped by category

### Scenario 2: Assert Commands

**Test**: Create assertion steps

1. Ask: "Create a Virtuoso test that asserts a login button exists"
2. **Expected**: Claude uses `virtuoso_assert_exists` tool
3. Verify the response includes step details and position

**Additional assert tests**:

- "Assert that the username field equals 'john@example.com'"
- "Assert that a checkbox is checked"
- "Assert that a value is greater than 100"

### Scenario 3: Interact Commands

**Test**: Create interaction steps

1. Ask: "Add a click on the Submit button to checkpoint 1680930"
2. **Expected**: Claude uses `virtuoso_interact_click` tool
3. Verify proper selector and position handling

**Additional interact tests**:

- "Type 'test@example.com' into the email field"
- "Double-click on the header element"
- "Press the Enter key"

### Scenario 4: Navigate Commands

**Test**: Create navigation steps

1. Ask: "Navigate to https://example.com in the test"
2. **Expected**: Claude uses `virtuoso_navigate_to` tool
3. Check URL formatting and position

**Additional navigate tests**:

- "Scroll to the bottom of the page"
- "Scroll to the contact form"
- "Scroll element #sidebar down by 200 pixels"

### Scenario 5: Data Management

**Test**: Store and manage data

1. Ask: "Store the username text in a variable called 'savedUser'"
2. **Expected**: Claude uses `virtuoso_data_store_text` tool
3. Verify variable name and selector

**Additional data tests**:

- "Create a session cookie with value 'abc123'"
- "Delete all cookies"
- "Store the input field value"

### Scenario 6: Wait Operations

**Test**: Add wait steps

1. Ask: "Wait for the loading spinner to appear"
2. **Expected**: Claude uses `virtuoso_wait_element` tool
3. Check timeout parameter handling

**Additional wait tests**:

- "Wait for 3 seconds"
- "Wait for #results with a 10 second timeout"

### Scenario 7: Window Management

**Test**: Window operations

1. Ask: "Resize the window to 1920x1080"
2. **Expected**: Claude uses `virtuoso_window_resize` tool
3. Verify dimensions are correct

**Additional window tests**:

- "Switch to the next tab"
- "Switch to iframe named 'payment'"

### Scenario 8: Complex Test Creation

**Test**: Multi-step test creation

1. Ask: "Create a login test that:
   - Navigates to https://example.com/login
   - Enters 'user@example.com' in the email field
   - Enters password in the password field
   - Clicks the login button
   - Asserts 'Welcome' text exists"
2. **Expected**: Claude uses multiple tools in sequence
3. Verify correct step ordering and positions

### Scenario 9: Library Commands

**Test**: Library checkpoint operations

1. Ask: "Get details for library checkpoint 7023"
2. **Expected**: Claude uses `virtuoso_library_get` tool
3. Verify response includes steps

**Additional library tests**:

- "Add checkpoint 1680930 to the library"
- "Attach library checkpoint 7023 to journey 608926 at position 4"
- "Move step 19660498 to position 2 in library checkpoint 7023"

### Scenario 10: Error Handling

**Test**: Invalid inputs

1. Ask: "Create a test step with invalid checkpoint ID 'abc'"
2. **Expected**: Claude handles the error gracefully
3. Verify error message is helpful

**Additional error tests**:

- Missing required parameters
- Invalid selectors
- Malformed URLs

## Validation Checklist

### Pre-Test Validation

- [ ] Virtuoso config file exists and is valid
- [ ] MCP server builds without errors
- [ ] Claude Desktop is version 0.7.0+
- [ ] Node.js version is 18+

### Connection Validation

- [ ] MCP server appears in Claude Desktop
- [ ] No connection errors in Claude logs
- [ ] Tools are accessible via "/" command
- [ ] Server responds to tool calls

### Tool Group Validation

- [ ] **Assert** (12 tools) - All assertion types work
- [ ] **Interact** (6 tools) - Click, type, and key actions work
- [ ] **Navigate** (5 tools) - URL navigation and scrolling work
- [ ] **Data** (5 tools) - Variable storage and cookies work
- [ ] **Dialog** (3 tools) - Alert handling works
- [ ] **Wait** (2 tools) - Element and time waits work
- [ ] **Window** (5 tools) - Resize and switching work
- [ ] **Mouse** (6 tools) - Mouse movements work
- [ ] **Select** (3 tools) - Dropdown selection works
- [ ] **File** (1 tool) - File upload works
- [ ] **Misc** (3 tools) - Comments work
- [ ] **Library** (6 tools) - Checkpoint operations work

### Response Validation

- [ ] Responses are properly formatted
- [ ] Step IDs are returned
- [ ] Positions increment correctly
- [ ] Error messages are clear
- [ ] Output format options work (human/json/yaml/ai)

## Troubleshooting

### MCP Server Not Appearing

1. Check Claude Desktop logs:

   - macOS: `~/Library/Logs/Claude/`
   - Windows: `%APPDATA%\Claude\logs\`
   - Linux: `~/.config/Claude/logs/`

2. Verify config file syntax (valid JSON)
3. Ensure absolute paths are used
4. Check Node.js is in PATH

### Tool Calls Failing

1. Verify Virtuoso config file exists
2. Check API key is valid
3. Test CLI directly: `node dist/index.js`
4. Look for error messages in responses

### Common Issues

**Issue**: "Cannot find module" errors
**Solution**: Rebuild with `npm run build`

**Issue**: "Authentication failed"
**Solution**: Check `~/.api-cli/virtuoso-config.yaml` has valid API key

**Issue**: Tools not showing in Claude
**Solution**: Restart Claude Desktop completely (not just reload)

**Issue**: Position conflicts
**Solution**: Let Claude manage positions automatically

## Advanced Testing

### Performance Testing

1. Create a test with 20+ steps rapidly
2. Monitor response times
3. Check for memory leaks with long sessions

### Concurrent Operations

1. Open multiple Claude conversations
2. Use Virtuoso tools in each simultaneously
3. Verify no cross-contamination

### Edge Cases

1. Very long selectors (>200 chars)
2. Special characters in text inputs
3. Maximum position numbers
4. Unicode in comments
5. Extremely long URLs

## Reporting Issues

When reporting issues, include:

1. **Claude Desktop version**
2. **Node.js version**
3. **Error messages** (full text)
4. **Steps to reproduce**
5. **Expected vs actual behavior**
6. **MCP server logs** if available

## Success Criteria

The MCP server integration is considered successful when:

- [x] All 12 tool groups are accessible
- [x] 90%+ of test scenarios pass
- [x] Error handling is graceful
- [x] Performance is acceptable (<2s response time)
- [x] No crashes during extended use
- [x] Clear error messages for invalid inputs

---

**Last Updated**: January 2025
**Version**: 1.0.0
