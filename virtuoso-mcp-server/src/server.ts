import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
  ListResourcesRequestSchema,
  ReadResourceRequestSchema,
  Tool,
} from "@modelcontextprotocol/sdk/types.js";
import { VirtuosoCliWrapper } from "./cli-wrapper.js";
import { toolRegistry, toolHandlers, registerTool } from "./tools/all-tools.js";
import { registerAssertTools } from "./tools/assert.js";
import { registerInteractTools } from "./tools/interact.js";
import { registerNavigateTools } from "./tools/navigate.js";
import { registerDataTools } from "./tools/data.js";
import { registerWaitTools } from "./tools/wait.js";
import { registerDialogTools } from "./tools/dialog.js";
import { registerWindowTools } from "./tools/window.js";
import { registerMouseTools } from "./tools/mouse.js";
import { registerSelectTools } from "./tools/select.js";
import { registerFileTools } from "./tools/file.js";
import { registerMiscTools } from "./tools/misc.js";
import { registerLibraryTools } from "./tools/library.js";

export interface VirtuosoMcpServerOptions {
  cliPath: string;
  configPath?: string;
  debug?: boolean;
}

export class VirtuosoMcpServer {
  private server: Server;
  private cli: VirtuosoCliWrapper;
  private options: VirtuosoMcpServerOptions;

  constructor(options: VirtuosoMcpServerOptions) {
    this.options = options;
    this.cli = new VirtuosoCliWrapper(options.cliPath, options.configPath);

    // Initialize MCP server
    this.server = new Server(
      {
        name: "virtuoso-cli-mcp",
        version: "1.0.0",
      },
      {
        capabilities: {
          tools: {},
          resources: {},
        },
      },
    );

    // Set up error handling
    this.server.onerror = (error) => {
      console.error("[MCP Server Error]", error);
    };

    // Register all tools and handlers
    this.registerTools();
    this.registerResources();
    this.registerSystemHandlers();
  }

  /**
   * Register all tool groups
   */
  private registerTools() {
    // First, set up the centralized handlers
    this.server.setRequestHandler(ListToolsRequestSchema, async () => {
      return { tools: toolRegistry };
    });

    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      const handler = toolHandlers.get(request.params.name);
      if (!handler) {
        throw new Error(`Unknown tool: ${request.params.name}`);
      }
      return handler(request.params.arguments);
    });

    // Register management tools FIRST
    this.registerManagementTools();

    // Register each tool group (they should preserve existing tools)
    registerAssertTools(this.cli);
    registerInteractTools(this.cli);
    registerNavigateTools(this.cli);
    registerDataTools(this.cli);
    registerWaitTools(this.cli);
    registerDialogTools(this.cli);
    registerWindowTools(this.cli);
    registerMouseTools(this.cli);
    registerSelectTools(this.cli);
    registerFileTools(this.cli);
    registerMiscTools(this.cli);
    registerLibraryTools(this.cli);
  }

  /**
   * Register management tools (project, goal, journey, checkpoint operations)
   */
  private registerManagementTools() {
    // Add set context tool
    registerTool(
      {
        name: "virtuoso_set_context",
        description:
          "Set the session context for subsequent commands (checkpoint ID, position, etc.)",
        inputSchema: {
          type: "object",
          properties: {
            checkpointId: {
              type: "string",
              description: "Checkpoint ID to use for subsequent commands",
            },
            position: {
              type: "number",
              description: "Starting position for steps",
            },
            journeyId: {
              type: "string",
              description: "Journey ID for context",
            },
            goalId: { type: "string", description: "Goal ID for context" },
          },
          additionalProperties: false,
        },
      },
      async (args) => {
        this.cli.updateContext(args);
        return {
          content: [
            {
              type: "text",
              text: `✅ Context updated: ${JSON.stringify(
                this.cli.getContext(),
                null,
                2,
              )}`,
            },
          ],
        };
      },
    );

    // Add get context tool
    registerTool(
      {
        name: "virtuoso_get_context",
        description: "Get the current session context",
        inputSchema: {
          type: "object",
          properties: {},
          additionalProperties: false,
        },
      },
      async () => {
        return {
          content: [
            {
              type: "text",
              text: `Current context:\n${JSON.stringify(
                this.cli.getContext(),
                null,
                2,
              )}`,
            },
          ],
        };
      },
    );
  }

  /**
   * Register resource handlers (for future expansion)
   */
  private registerResources() {
    // List available resources
    this.server.setRequestHandler(ListResourcesRequestSchema, async () => {
      return {
        resources: [
          {
            uri: "virtuoso://session",
            name: "Current Session",
            description: "Current Virtuoso session context and state",
            mimeType: "application/json",
          },
        ],
      };
    });

    // Read resource content
    this.server.setRequestHandler(
      ReadResourceRequestSchema,
      async (request) => {
        if (request.params.uri === "virtuoso://session") {
          return {
            contents: [
              {
                uri: "virtuoso://session",
                mimeType: "application/json",
                text: JSON.stringify(
                  {
                    context: this.cli.getContext(),
                    server: {
                      version: "1.0.0",
                      cliPath: this.options.cliPath,
                      configPath: this.options.configPath,
                    },
                  },
                  null,
                  2,
                ),
              },
            ],
          };
        }

        throw new Error(`Unknown resource: ${request.params.uri}`);
      },
    );
  }

  /**
   * Register system-level handlers
   */
  private registerSystemHandlers() {
    // Log debug information if enabled
    if (this.options.debug) {
      console.error("[MCP Debug] Debug logging enabled");
    }
  }

  /**
   * Start the MCP server
   */
  async start() {
    const transport = new StdioServerTransport();
    await this.server.connect(transport);

    console.error("Virtuoso MCP Server started successfully");
    console.error(`CLI Path: ${this.options.cliPath}`);
    console.error(`Config Path: ${this.options.configPath || "default"}`);

    // Handle graceful shutdown
    process.on("SIGINT", async () => {
      console.error("Shutting down Virtuoso MCP Server...");
      await this.server.close();
      process.exit(0);
    });
  }
}
