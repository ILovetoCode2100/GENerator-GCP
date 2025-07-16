#!/usr/bin/env node

const { VirtuosoMcpServer } = require("../dist/server.js");
const { Server } = require("@modelcontextprotocol/sdk/server/index.js");

console.log("Testing tool registration...\n");

// Create a mock server to capture tool registrations
let registeredTools = [];
let handlers = {};

const mockServer = {
  setRequestHandler: (schema, handler) => {
    const schemaName = schema.parse ? "ListToolsRequestSchema" : "Unknown";
    console.log(`Registering handler for: ${schemaName}`);

    if (schemaName === "ListToolsRequestSchema") {
      // Call the handler to see what tools it returns
      const mockRequest = { result: { tools: [] } };
      handler(mockRequest)
        .then((result) => {
          console.log(
            `Tools registered by this handler: ${result.tools
              .map((t) => t.name)
              .join(", ")}`,
          );
          registeredTools.push(...result.tools);
        })
        .catch((err) => {
          console.error("Error calling handler:", err);
        });
    }

    handlers[schemaName] = handler;
  },
  onerror: null,
};

// Override the Server constructor
const OriginalServer = Server;
global.Server = function (...args) {
  console.log("Creating server with args:", args);
  return mockServer;
};

// Import server module which should trigger registrations
const serverModule = require("../dist/server.js");

// Create server instance
const server = new serverModule.VirtuosoMcpServer({
  cliPath: "../bin/api-cli",
  configPath: "~/.api-cli/virtuoso-config.yaml",
  debug: true,
});

// Wait a bit for async operations
setTimeout(() => {
  console.log("\n=== Final Results ===");
  console.log(`Total tools registered: ${registeredTools.length}`);
  console.log("\nAll registered tools:");
  registeredTools.forEach((tool, i) => {
    console.log(
      `${i + 1}. ${tool.name} - ${tool.description.substring(0, 50)}...`,
    );
  });
}, 1000);
