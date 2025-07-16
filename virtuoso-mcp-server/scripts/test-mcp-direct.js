#!/usr/bin/env node

// Test MCP server directly with proper protocol
import { spawn } from "child_process";
import readline from "readline";

const server = spawn("node", ["dist/index.js"], {
  stdio: ["pipe", "pipe", "pipe"],
  env: {
    ...process.env,
    VIRTUOSO_CLI_PATH: "../bin/api-cli",
    VIRTUOSO_CONFIG_PATH: "~/.api-cli/virtuoso-config.yaml",
    VIRTUOSO_DEBUG: "true",
  },
});

const rl = readline.createInterface({
  input: server.stdout,
  crlfDelay: Infinity,
});

let messageId = 1;

// Handle server output
rl.on("line", (line) => {
  console.log("Server:", line);
});

// Handle server errors
server.stderr.on("data", (data) => {
  console.error("Server Error:", data.toString());
});

// Send initialization
const init = {
  jsonrpc: "2.0",
  method: "initialize",
  params: {
    protocolVersion: "2024.11",
    capabilities: {},
    clientInfo: {
      name: "test-client",
      version: "1.0.0",
    },
  },
  id: messageId++,
};

console.log("Sending:", JSON.stringify(init));
server.stdin.write(JSON.stringify(init) + "\n");

// After initialization, list tools
setTimeout(() => {
  const listTools = {
    jsonrpc: "2.0",
    method: "tools/list",
    params: {},
    id: messageId++,
  };

  console.log("Sending:", JSON.stringify(listTools));
  server.stdin.write(JSON.stringify(listTools) + "\n");

  // Give it time to respond then exit
  setTimeout(() => {
    server.kill();
    process.exit(0);
  }, 2000);
}, 1000);
