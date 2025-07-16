import { jest } from "@jest/globals";
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { VirtuosoMcpServer } from "../server.js";
import { VirtuosoCliWrapper } from "../cli-wrapper.js";

// Mock dependencies
jest.mock("@modelcontextprotocol/sdk/server/index.js");
jest.mock("@modelcontextprotocol/sdk/server/stdio.js");
jest.mock("../cli-wrapper.js");

// Mock all tool registration modules
jest.mock("../tools/assert.js", () => ({
  registerAssertTools: jest.fn(),
}));
jest.mock("../tools/interact.js", () => ({
  registerInteractTools: jest.fn(),
}));
jest.mock("../tools/navigate.js", () => ({
  registerNavigateTools: jest.fn(),
}));
jest.mock("../tools/data.js", () => ({
  registerDataTools: jest.fn(),
}));
jest.mock("../tools/wait.js", () => ({
  registerWaitTools: jest.fn(),
}));
jest.mock("../tools/dialog.js", () => ({
  registerDialogTools: jest.fn(),
}));
jest.mock("../tools/window.js", () => ({
  registerWindowTools: jest.fn(),
}));
jest.mock("../tools/mouse.js", () => ({
  registerMouseTools: jest.fn(),
}));
jest.mock("../tools/select.js", () => ({
  registerSelectTools: jest.fn(),
}));
jest.mock("../tools/file.js", () => ({
  registerFileTools: jest.fn(),
}));
jest.mock("../tools/misc.js", () => ({
  registerMiscTools: jest.fn(),
}));
jest.mock("../tools/library.js", () => ({
  registerLibraryTools: jest.fn(),
}));

describe("VirtuosoMcpServer", () => {
  let mockServer: jest.Mocked<Server>;
  let mockTransport: jest.Mocked<StdioServerTransport>;
  let mockCliWrapper: jest.Mocked<VirtuosoCliWrapper>;
  let server: VirtuosoMcpServer;

  beforeEach(() => {
    jest.clearAllMocks();

    // Setup mock Server
    mockServer = {
      setRequestHandler: jest.fn(),
      connect: jest.fn(),
      close: jest.fn(),
      onerror: undefined,
      onrequest: undefined,
    } as any;

    (Server as jest.MockedClass<typeof Server>).mockImplementation(
      () => mockServer,
    );

    // Setup mock transport
    mockTransport = {} as any;
    (
      StdioServerTransport as jest.MockedClass<typeof StdioServerTransport>
    ).mockImplementation(() => mockTransport);

    // Setup mock CLI wrapper
    mockCliWrapper = {
      updateContext: jest.fn(),
      getContext: jest.fn().mockReturnValue({
        checkpointId: "12345",
        position: 1,
        journeyId: undefined,
        goalId: undefined,
      }),
      execute: jest.fn(),
    } as any;

    (
      VirtuosoCliWrapper as jest.MockedClass<typeof VirtuosoCliWrapper>
    ).mockImplementation(() => mockCliWrapper);
  });

  describe("Constructor", () => {
    test("should initialize with provided options", () => {
      server = new VirtuosoMcpServer({
        cliPath: "/path/to/cli",
        configPath: "/path/to/config.yaml",
        debug: true,
      });

      expect(VirtuosoCliWrapper).toHaveBeenCalledWith(
        "/path/to/cli",
        "/path/to/config.yaml",
      );
      expect(Server).toHaveBeenCalledWith(
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
    });

    test("should register all tool groups", () => {
      const {
        registerAssertTools,
        registerInteractTools,
        registerNavigateTools,
        registerDataTools,
        registerWaitTools,
        registerDialogTools,
        registerWindowTools,
        registerMouseTools,
        registerSelectTools,
        registerFileTools,
        registerMiscTools,
        registerLibraryTools,
      } = jest.requireMock("../tools/assert.js");

      server = new VirtuosoMcpServer({
        cliPath: "/path/to/cli",
      });

      // Verify all tool groups are registered
      expect(registerAssertTools).toHaveBeenCalledWith(
        mockServer,
        mockCliWrapper,
      );
      expect(
        jest.requireMock("../tools/interact.js").registerInteractTools,
      ).toHaveBeenCalledWith(mockServer, mockCliWrapper);
      expect(
        jest.requireMock("../tools/navigate.js").registerNavigateTools,
      ).toHaveBeenCalledWith(mockServer, mockCliWrapper);
      expect(
        jest.requireMock("../tools/data.js").registerDataTools,
      ).toHaveBeenCalledWith(mockServer, mockCliWrapper);
      expect(
        jest.requireMock("../tools/wait.js").registerWaitTools,
      ).toHaveBeenCalledWith(mockServer, mockCliWrapper);
      expect(
        jest.requireMock("../tools/dialog.js").registerDialogTools,
      ).toHaveBeenCalledWith(mockServer, mockCliWrapper);
      expect(
        jest.requireMock("../tools/window.js").registerWindowTools,
      ).toHaveBeenCalledWith(mockServer, mockCliWrapper);
      expect(
        jest.requireMock("../tools/mouse.js").registerMouseTools,
      ).toHaveBeenCalledWith(mockServer, mockCliWrapper);
      expect(
        jest.requireMock("../tools/select.js").registerSelectTools,
      ).toHaveBeenCalledWith(mockServer, mockCliWrapper);
      expect(
        jest.requireMock("../tools/file.js").registerFileTools,
      ).toHaveBeenCalledWith(mockServer, mockCliWrapper);
      expect(
        jest.requireMock("../tools/misc.js").registerMiscTools,
      ).toHaveBeenCalledWith(mockServer, mockCliWrapper);
      expect(
        jest.requireMock("../tools/library.js").registerLibraryTools,
      ).toHaveBeenCalledWith(mockServer, mockCliWrapper);
    });

    test("should set up error handler", () => {
      server = new VirtuosoMcpServer({
        cliPath: "/path/to/cli",
      });

      expect(mockServer.onerror).toBeDefined();

      // Test error handler
      const consoleErrorSpy = jest.spyOn(console, "error").mockImplementation();
      const testError = new Error("Test error");
      mockServer.onerror!(testError);

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "[MCP Server Error]",
        testError,
      );
      consoleErrorSpy.mockRestore();
    });

    test("should enable debug mode when specified", () => {
      server = new VirtuosoMcpServer({
        cliPath: "/path/to/cli",
        debug: true,
      });

      expect(mockServer.onrequest).toBeDefined();

      // Test request logger
      const consoleErrorSpy = jest.spyOn(console, "error").mockImplementation();
      const testRequest = { method: "test", params: {} };
      mockServer.onrequest!(testRequest);

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "[MCP Request]",
        JSON.stringify(testRequest, null, 2),
      );
      consoleErrorSpy.mockRestore();
    });
  });

  describe("Context Management Tools", () => {
    beforeEach(() => {
      server = new VirtuosoMcpServer({
        cliPath: "/path/to/cli",
      });
    });

    test("should handle virtuoso_set_context tool", async () => {
      // Find the CallToolRequestSchema handler
      const callToolHandler = mockServer.setRequestHandler.mock.calls.find(
        (call) => call[0] === CallToolRequestSchema,
      )?.[1];

      expect(callToolHandler).toBeDefined();

      const result = await callToolHandler!({
        params: {
          name: "virtuoso_set_context",
          arguments: {
            checkpointId: "67890",
            position: 5,
          },
        },
      } as any);

      expect(mockCliWrapper.updateContext).toHaveBeenCalledWith({
        checkpointId: "67890",
        position: 5,
      });

      expect(result).toEqual({
        content: [
          {
            type: "text",
            text: expect.stringContaining("✅ Context updated:"),
          },
        ],
      });
    });

    test("should handle virtuoso_get_context tool", async () => {
      // Find the CallToolRequestSchema handler
      const callToolHandler = mockServer.setRequestHandler.mock.calls.find(
        (call) => call[0] === CallToolRequestSchema,
      )?.[1];

      const result = await callToolHandler!({
        params: {
          name: "virtuoso_get_context",
          arguments: {},
        },
      } as any);

      expect(result).toEqual({
        content: [
          {
            type: "text",
            text: expect.stringContaining("Current context:"),
          },
        ],
      });
    });
  });

  describe("Resource Handlers", () => {
    beforeEach(() => {
      server = new VirtuosoMcpServer({
        cliPath: "/path/to/cli",
        configPath: "/path/to/config.yaml",
      });
    });

    test("should list available resources", async () => {
      const listResourcesHandler = mockServer.setRequestHandler.mock.calls.find(
        (call) => call[0] === ListResourcesRequestSchema,
      )?.[1];

      const result = await listResourcesHandler!({} as any);

      expect(result).toEqual({
        resources: [
          {
            uri: "virtuoso://session",
            name: "Current Session",
            description: "Current Virtuoso session context and state",
            mimeType: "application/json",
          },
        ],
      });
    });

    test("should read session resource", async () => {
      const readResourceHandler = mockServer.setRequestHandler.mock.calls.find(
        (call) => call[0] === ReadResourceRequestSchema,
      )?.[1];

      const result = await readResourceHandler!({
        params: { uri: "virtuoso://session" },
      } as any);

      expect(result).toEqual({
        contents: [
          {
            uri: "virtuoso://session",
            mimeType: "application/json",
            text: expect.stringContaining('"context":'),
          },
        ],
      });

      // Verify the content includes expected data
      const content = JSON.parse(result.contents[0].text);
      expect(content).toHaveProperty("context");
      expect(content).toHaveProperty("server");
      expect(content.server).toEqual({
        version: "1.0.0",
        cliPath: "/path/to/cli",
        configPath: "/path/to/config.yaml",
      });
    });

    test("should throw error for unknown resource", async () => {
      const readResourceHandler = mockServer.setRequestHandler.mock.calls.find(
        (call) => call[0] === ReadResourceRequestSchema,
      )?.[1];

      await expect(
        readResourceHandler!({
          params: { uri: "virtuoso://unknown" },
        } as any),
      ).rejects.toThrow("Unknown resource: virtuoso://unknown");
    });
  });

  describe("Server Lifecycle", () => {
    beforeEach(() => {
      server = new VirtuosoMcpServer({
        cliPath: "/path/to/cli",
      });
    });

    test("should start server and connect transport", async () => {
      const consoleErrorSpy = jest.spyOn(console, "error").mockImplementation();

      await server.start();

      expect(StdioServerTransport).toHaveBeenCalled();
      expect(mockServer.connect).toHaveBeenCalledWith(mockTransport);

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Virtuoso MCP Server started successfully",
      );
      expect(consoleErrorSpy).toHaveBeenCalledWith("CLI Path: /path/to/cli");
      expect(consoleErrorSpy).toHaveBeenCalledWith("Config Path: default");

      consoleErrorSpy.mockRestore();
    });

    test("should handle SIGINT gracefully", async () => {
      const consoleErrorSpy = jest.spyOn(console, "error").mockImplementation();
      const processExitSpy = jest
        .spyOn(process, "exit")
        .mockImplementation(() => {
          throw new Error("process.exit called");
        });

      await server.start();

      // Get the SIGINT handler
      const sigintHandler = process.listeners("SIGINT").pop() as Function;
      expect(sigintHandler).toBeDefined();

      // Trigger SIGINT
      try {
        await sigintHandler();
      } catch (e) {
        // Expected due to process.exit mock
      }

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Shutting down Virtuoso MCP Server...",
      );
      expect(mockServer.close).toHaveBeenCalled();
      expect(processExitSpy).toHaveBeenCalledWith(0);

      consoleErrorSpy.mockRestore();
      processExitSpy.mockRestore();

      // Clean up listener
      process.removeListener("SIGINT", sigintHandler);
    });
  });
});
