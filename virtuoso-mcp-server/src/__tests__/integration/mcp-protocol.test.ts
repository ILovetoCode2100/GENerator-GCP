import {
  describe,
  it,
  expect,
  jest,
  beforeEach,
  afterEach,
} from "@jest/globals";
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
  ListResourcesRequestSchema,
  ReadResourceRequestSchema,
  InitializeRequestSchema,
  ListPromptsRequestSchema,
  ToolsCapability,
  ResourcesCapability,
} from "@modelcontextprotocol/sdk/types.js";
import { VirtuosoMcpServer } from "../../server.js";
import { VirtuosoCliWrapper } from "../../cli-wrapper.js";
import path from "path";
import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Mock the CLI wrapper
jest.mock("../../cli-wrapper.js");

describe("MCP Protocol Integration Tests", () => {
  let server: VirtuosoMcpServer;
  let mockCliWrapper: jest.Mocked<VirtuosoCliWrapper>;
  const mockCliPath = "/mock/path/to/cli";
  const mockConfigPath = "/mock/config.yaml";

  beforeEach(() => {
    // Reset mocks
    jest.clearAllMocks();

    // Create server instance
    server = new VirtuosoMcpServer({
      cliPath: mockCliPath,
      configPath: mockConfigPath,
      debug: false,
    });

    // Get the mocked CLI wrapper instance
    mockCliWrapper = (
      VirtuosoCliWrapper as jest.MockedClass<typeof VirtuosoCliWrapper>
    ).mock.instances[0] as jest.Mocked<VirtuosoCliWrapper>;
  });

  afterEach(() => {
    // Cleanup
    jest.restoreAllMocks();
  });

  describe("MCP Protocol Compliance", () => {
    it("should handle initialize request", async () => {
      const mockTransport = {
        send: jest.fn(),
        close: jest.fn(),
        onError: jest.fn(),
        onClose: jest.fn(),
      };

      // Access the private server instance for testing
      const serverInstance = (server as any).server as Server;

      // Mock the request handling
      const initializeHandler = jest.fn().mockResolvedValue({
        protocolVersion: "2024-11-05",
        capabilities: {
          tools: {},
          resources: {},
        },
        serverInfo: {
          name: "virtuoso-cli-mcp",
          version: "1.0.0",
        },
      });

      // Test that server responds to initialize correctly
      expect(serverInstance).toBeDefined();
      expect((serverInstance as any).serverInfo.name).toBe("virtuoso-cli-mcp");
      expect((serverInstance as any).serverInfo.version).toBe("1.0.0");
    });

    it("should support required capabilities", async () => {
      const serverInstance = (server as any).server as Server;
      const capabilities = (serverInstance as any).capabilities;

      expect(capabilities).toBeDefined();
      expect(capabilities.tools).toBeDefined();
      expect(capabilities.resources).toBeDefined();
    });
  });

  describe("Tool Listing", () => {
    it("should list all available tools", async () => {
      const serverInstance = (server as any).server as Server;

      // Get the tools list handler
      const handlers = (serverInstance as any)._requestHandlers;
      const listToolsHandler = handlers.get("tools/list");

      expect(listToolsHandler).toBeDefined();

      // Call the handler
      const result = await listToolsHandler({});

      expect(result.tools).toBeDefined();
      expect(Array.isArray(result.tools)).toBe(true);

      // Check for expected tool categories
      const toolNames = result.tools.map((tool: any) => tool.name);

      // Core tools
      expect(toolNames).toContain("virtuoso_assert");
      expect(toolNames).toContain("virtuoso_interact");
      expect(toolNames).toContain("virtuoso_navigate");
      expect(toolNames).toContain("virtuoso_data");
      expect(toolNames).toContain("virtuoso_wait");
      expect(toolNames).toContain("virtuoso_dialog");
      expect(toolNames).toContain("virtuoso_window");
      expect(toolNames).toContain("virtuoso_mouse");
      expect(toolNames).toContain("virtuoso_select");
      expect(toolNames).toContain("virtuoso_file");
      expect(toolNames).toContain("virtuoso_misc");
      expect(toolNames).toContain("virtuoso_library");

      // Management tools
      expect(toolNames).toContain("virtuoso_set_context");
      expect(toolNames).toContain("virtuoso_list_projects");
      expect(toolNames).toContain("virtuoso_create_checkpoint");
    });

    it("should provide proper tool schemas", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const listToolsHandler = handlers.get("tools/list");

      const result = await listToolsHandler({});
      const interactTool = result.tools.find(
        (tool: any) => tool.name === "virtuoso_interact",
      );

      expect(interactTool).toBeDefined();
      expect(interactTool.description).toBeDefined();
      expect(interactTool.inputSchema).toBeDefined();
      expect(interactTool.inputSchema.type).toBe("object");
      expect(interactTool.inputSchema.properties).toBeDefined();
      expect(interactTool.inputSchema.properties.action).toBeDefined();
      expect(interactTool.inputSchema.required).toContain("action");
    });
  });

  describe("Tool Calling", () => {
    it("should handle tool call requests", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");

      expect(callToolHandler).toBeDefined();

      // Mock CLI response
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "test-id", type: "UI_INTERACTION" },
        raw: JSON.stringify({ id: "test-id", type: "UI_INTERACTION" }),
      });

      // Call a tool
      const result = await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Submit button",
            checkpoint: "123456",
          },
        },
      });

      expect(result).toBeDefined();
      expect(result.content).toBeDefined();
      expect(Array.isArray(result.content)).toBe(true);
      expect(result.content[0].type).toBe("text");
    });

    it("should handle tool call errors gracefully", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");

      // Mock CLI error
      mockCliWrapper.execute.mockResolvedValue({
        success: false,
        error: "Command failed: checkpoint not found",
        raw: "",
      });

      // Call a tool with invalid params
      const result = await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Submit button",
            checkpoint: "invalid",
          },
        },
      });

      expect(result).toBeDefined();
      expect(result.content[0].type).toBe("text");
      expect(result.content[0].text).toContain("error");
    });

    it("should validate tool input schemas", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");

      // Call with invalid action
      await expect(
        callToolHandler({
          params: {
            name: "virtuoso_interact",
            arguments: {
              action: "invalid-action",
              selector: "Submit button",
            },
          },
        }),
      ).rejects.toThrow();
    });
  });

  describe("Resource Handling", () => {
    it("should list available resources", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const listResourcesHandler = handlers.get("resources/list");

      expect(listResourcesHandler).toBeDefined();

      const result = await listResourcesHandler({});

      expect(result.resources).toBeDefined();
      expect(Array.isArray(result.resources)).toBe(true);

      // Check for expected resources
      const resourceUris = result.resources.map(
        (resource: any) => resource.uri,
      );
      expect(resourceUris).toContain("virtuoso://session/context");
      expect(resourceUris).toContain("virtuoso://projects/list");
    });

    it("should read session context resource", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const readResourceHandler = handlers.get("resources/read");

      expect(readResourceHandler).toBeDefined();

      // Mock context
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "123456",
        position: 5,
        journeyId: "789",
        goalId: "456",
      });

      const result = await readResourceHandler({
        params: {
          uri: "virtuoso://session/context",
        },
      });

      expect(result.contents).toBeDefined();
      expect(Array.isArray(result.contents)).toBe(true);
      expect(result.contents[0].mimeType).toBe("application/json");

      const contextData = JSON.parse(result.contents[0].text);
      expect(contextData.checkpointId).toBe("123456");
      expect(contextData.position).toBe(5);
    });

    it("should handle resource read errors", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const readResourceHandler = handlers.get("resources/read");

      await expect(
        readResourceHandler({
          params: {
            uri: "virtuoso://invalid/resource",
          },
        }),
      ).rejects.toThrow("Resource not found");
    });
  });

  describe("Session Context Persistence", () => {
    it("should maintain context across multiple tool calls", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");

      // Set context
      await callToolHandler({
        params: {
          name: "virtuoso_set_context",
          arguments: {
            checkpointId: "123456",
            position: 1,
          },
        },
      });

      expect(mockCliWrapper.updateContext).toHaveBeenCalledWith({
        checkpointId: "123456",
        position: 1,
      });

      // Mock execute to return success
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-1" },
        raw: JSON.stringify({ id: "step-1" }),
      });

      // Call tool without explicit checkpoint
      await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Submit",
          },
        },
      });

      // Verify CLI was called with context
      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "interact",
        ["click", "Submit"],
        expect.objectContaining({
          checkpoint: "123456",
          position: 1,
        }),
      );
    });

    it("should auto-increment position after successful commands", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");

      // Mock context and execute
      mockCliWrapper.getContext.mockReturnValue({ position: 1 });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-1" },
        raw: JSON.stringify({ id: "step-1" }),
      });

      // First call
      await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Button 1",
            checkpoint: "123456",
          },
        },
      });

      // Verify position was incremented
      expect(mockCliWrapper.updateContext).toHaveBeenCalledWith({
        position: 2,
      });
    });
  });

  describe("Error Handling", () => {
    it("should handle unknown tool names", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");

      await expect(
        callToolHandler({
          params: {
            name: "virtuoso_unknown_tool",
            arguments: {},
          },
        }),
      ).rejects.toThrow("Unknown tool");
    });

    it("should handle CLI execution errors", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");

      // Mock CLI throwing error
      mockCliWrapper.execute.mockRejectedValue(
        new Error("CLI execution failed"),
      );

      const result = await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Submit",
          },
        },
      });

      expect(result.content[0].text).toContain("Error");
      expect(result.isError).toBe(true);
    });
  });

  describe("MCP Message Format", () => {
    it("should format successful responses correctly", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");

      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: {
          id: "test-step-123",
          type: "UI_INTERACTION",
          status: "created",
        },
        raw: JSON.stringify({
          id: "test-step-123",
          type: "UI_INTERACTION",
          status: "created",
        }),
      });

      const result = await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Submit",
            checkpoint: "123456",
          },
        },
      });

      expect(result).toMatchObject({
        content: [
          {
            type: "text",
            text: expect.stringContaining("Successfully executed"),
          },
        ],
      });
    });

    it("should include debug information when debug mode is enabled", async () => {
      // Create server with debug enabled
      const debugServer = new VirtuosoMcpServer({
        cliPath: mockCliPath,
        configPath: mockConfigPath,
        debug: true,
      });

      const serverInstance = (debugServer as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");

      // Get the debug CLI wrapper
      const debugCliWrapper = (
        VirtuosoCliWrapper as jest.MockedClass<typeof VirtuosoCliWrapper>
      ).mock.instances[1] as jest.Mocked<VirtuosoCliWrapper>;

      debugCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "test-123" },
        raw: JSON.stringify({ id: "test-123" }),
      });

      const result = await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Submit",
            checkpoint: "123456",
          },
        },
      });

      // In debug mode, should include raw response
      expect(
        result.content.some((c: any) => c.text?.includes("Raw response")),
      ).toBe(true);
    });
  });
});
