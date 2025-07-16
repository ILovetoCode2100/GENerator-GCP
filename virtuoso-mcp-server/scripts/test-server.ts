#!/usr/bin/env tsx
/**
 * Test Server Script
 *
 * This script simulates Claude Desktop interaction with the MCP server
 * and tests all tool groups with sample calls.
 */

import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";
import { spawn } from "child_process";
import path from "path";
import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Color codes for terminal output
const colors = {
  reset: "\x1b[0m",
  green: "\x1b[32m",
  red: "\x1b[31m",
  yellow: "\x1b[33m",
  blue: "\x1b[34m",
  dim: "\x1b[2m",
};

// Test scenarios for each tool group
const testScenarios = [
  // Assert Tools
  {
    tool: "virtuoso_assert_exists",
    args: { checkpointId: "1680930", selector: "Login button", position: 1 },
    description: "Test assert exists command",
  },
  {
    tool: "virtuoso_assert_equals",
    args: {
      checkpointId: "1680930",
      target: "Username",
      expectedValue: "john@example.com",
      position: 2,
    },
    description: "Test assert equals command",
  },

  // Interact Tools
  {
    tool: "virtuoso_interact_click",
    args: { checkpointId: "1680930", selector: "Submit", position: 3 },
    description: "Test interact click command",
  },
  {
    tool: "virtuoso_interact_write",
    args: {
      checkpointId: "1680930",
      selector: "Email field",
      text: "test@example.com",
      position: 4,
    },
    description: "Test interact write command",
  },

  // Navigate Tools
  {
    tool: "virtuoso_navigate_to",
    args: { checkpointId: "1680930", url: "https://example.com", position: 5 },
    description: "Test navigate to command",
  },
  {
    tool: "virtuoso_navigate_scroll_top",
    args: { checkpointId: "1680930", position: 6 },
    description: "Test navigate scroll top command",
  },

  // Data Tools
  {
    tool: "virtuoso_data_store_text",
    args: {
      checkpointId: "1680930",
      selector: "Username",
      variableName: "userVar",
      position: 7,
    },
    description: "Test data store text command",
  },
  {
    tool: "virtuoso_data_cookie_create",
    args: {
      checkpointId: "1680930",
      name: "session",
      value: "abc123",
      position: 8,
    },
    description: "Test data cookie create command",
  },

  // Wait Tools
  {
    tool: "virtuoso_wait_element",
    args: {
      checkpointId: "1680930",
      selector: "#loader",
      position: 9,
      timeout: 5000,
    },
    description: "Test wait element command",
  },
  {
    tool: "virtuoso_wait_time",
    args: { checkpointId: "1680930", milliseconds: 2000, position: 10 },
    description: "Test wait time command",
  },

  // Window Tools
  {
    tool: "virtuoso_window_resize",
    args: { checkpointId: "1680930", width: 1024, height: 768, position: 11 },
    description: "Test window resize command",
  },

  // Mouse Tools
  {
    tool: "virtuoso_mouse_move_to",
    args: { checkpointId: "1680930", x: 100, y: 200, position: 12 },
    description: "Test mouse move to command",
  },

  // Select Tools
  {
    tool: "virtuoso_select_option",
    args: {
      checkpointId: "1680930",
      selector: "#country",
      option: "USA",
      position: 13,
    },
    description: "Test select option command",
  },

  // File Tools
  {
    tool: "virtuoso_file_upload",
    args: {
      checkpointId: "1680930",
      url: "https://example.com/file.pdf",
      selector: "#file-input",
      position: 14,
    },
    description: "Test file upload command",
  },

  // Misc Tools
  {
    tool: "virtuoso_misc_comment",
    args: { checkpointId: "1680930", text: "Test login flow", position: 15 },
    description: "Test misc comment command",
  },

  // Library Tools
  {
    tool: "virtuoso_library_get",
    args: { checkpointId: "7023" },
    description: "Test library get command",
  },
];

async function testServer() {
  console.log(
    `${colors.blue}Starting Virtuoso MCP Server Test...${colors.reset}\n`,
  );

  // Check if config exists
  const configPath = path.join(
    process.env.HOME || "",
    ".api-cli",
    "virtuoso-config.yaml",
  );
  console.log(
    `${colors.dim}Checking for config at: ${configPath}${colors.reset}`,
  );

  // Start the server
  const serverPath = path.join(__dirname, "..", "dist", "index.js");
  console.log(
    `${colors.dim}Starting server from: ${serverPath}${colors.reset}\n`,
  );

  const serverProcess = spawn("node", [serverPath], {
    stdio: ["pipe", "pipe", "pipe"],
  });

  // Create client
  const transport = new StdioClientTransport({
    command: "node",
    args: [serverPath],
  });

  const client = new Client(
    {
      name: "virtuoso-test-client",
      version: "1.0.0",
    },
    {
      capabilities: {},
    },
  );

  try {
    // Connect to server
    console.log(`${colors.yellow}Connecting to server...${colors.reset}`);
    await client.connect(transport);
    console.log(`${colors.green}✓ Connected successfully${colors.reset}\n`);

    // List available tools
    console.log(`${colors.yellow}Listing available tools...${colors.reset}`);
    const tools = await client.listTools();
    console.log(
      `${colors.green}✓ Found ${tools.tools.length} tools${colors.reset}`,
    );
    console.log(
      `${colors.dim}Tool groups: ${Array.from(
        new Set(tools.tools.map((t) => t.name.split("_")[1])),
      ).join(", ")}${colors.reset}\n`,
    );

    // Run test scenarios
    console.log(`${colors.yellow}Running test scenarios...${colors.reset}\n`);

    let passed = 0;
    let failed = 0;

    for (const scenario of testScenarios) {
      console.log(`${colors.blue}Test: ${scenario.description}${colors.reset}`);
      console.log(`${colors.dim}Tool: ${scenario.tool}${colors.reset}`);
      console.log(
        `${colors.dim}Args: ${JSON.stringify(scenario.args, null, 2)}${
          colors.reset
        }`,
      );

      try {
        const startTime = Date.now();
        const result = await client.callTool({
          name: scenario.tool,
          arguments: scenario.args,
        });
        const duration = Date.now() - startTime;

        if (result.content && result.content.length > 0) {
          console.log(
            `${colors.green}✓ Success (${duration}ms)${colors.reset}`,
          );
          console.log(
            `${colors.dim}Response: ${JSON.stringify(
              result.content[0],
              null,
              2,
            )}${colors.reset}`,
          );
          passed++;
        } else {
          console.log(
            `${colors.red}✗ Failed: No content in response${colors.reset}`,
          );
          failed++;
        }
      } catch (error) {
        console.log(
          `${colors.red}✗ Failed: ${
            error instanceof Error ? error.message : "Unknown error"
          }${colors.reset}`,
        );
        failed++;
      }

      console.log("");
    }

    // Summary
    console.log(`${colors.yellow}Test Summary:${colors.reset}`);
    console.log(`${colors.green}Passed: ${passed}${colors.reset}`);
    console.log(`${colors.red}Failed: ${failed}${colors.reset}`);
    console.log(`Total: ${passed + failed}\n`);

    // Test error handling
    console.log(`${colors.yellow}Testing error handling...${colors.reset}`);
    try {
      await client.callTool({
        name: "virtuoso_invalid_tool",
        arguments: {},
      });
      console.log(
        `${colors.red}✗ Error handling failed - invalid tool was accepted${colors.reset}`,
      );
    } catch (error) {
      console.log(
        `${colors.green}✓ Error handling works correctly${colors.reset}`,
      );
      console.log(
        `${colors.dim}Error: ${
          error instanceof Error ? error.message : "Unknown error"
        }${colors.reset}\n`,
      );
    }
  } catch (error) {
    console.error(`${colors.red}Test failed:${colors.reset}`, error);
    process.exit(1);
  } finally {
    // Cleanup
    console.log(`${colors.yellow}Cleaning up...${colors.reset}`);
    await client.close();
    serverProcess.kill();
    console.log(`${colors.green}✓ Test completed${colors.reset}`);
  }
}

// Run the test
testServer().catch(console.error);
