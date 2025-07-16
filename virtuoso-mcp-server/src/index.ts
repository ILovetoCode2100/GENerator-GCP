import { VirtuosoMcpServer } from "./server.js";
import path from "path";
import { fileURLToPath } from "url";
import dotenv from "dotenv";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Load environment variables
dotenv.config();

async function main() {
  try {
    // Get configuration from environment variables or defaults
    const cliPath =
      process.env.VIRTUOSO_CLI_PATH ||
      path.resolve(__dirname, "../../bin/api-cli");

    const configPath =
      process.env.VIRTUOSO_CONFIG_PATH ||
      path.join(process.env.HOME || "", ".api-cli/virtuoso-config.yaml");

    const debug =
      process.env.DEBUG === "true" || process.env.NODE_ENV === "development";

    // Validate CLI path exists
    const fs = await import("fs/promises");
    try {
      await fs.access(cliPath);
    } catch (error) {
      console.error(`Error: CLI not found at ${cliPath}`);
      console.error("Please set VIRTUOSO_CLI_PATH environment variable");
      process.exit(1);
    }

    // Create and start server
    const server = new VirtuosoMcpServer({
      cliPath,
      configPath,
      debug,
    });

    await server.start();

    // Keep the process alive
    process.stdin.resume();
  } catch (error) {
    console.error("Failed to start Virtuoso MCP Server:", error);
    process.exit(1);
  }
}

// Handle uncaught errors
process.on("uncaughtException", (error) => {
  console.error("Uncaught exception:", error);
  process.exit(1);
});

process.on("unhandledRejection", (reason, promise) => {
  console.error("Unhandled rejection at:", promise, "reason:", reason);
  process.exit(1);
});

// Start the server
main();
