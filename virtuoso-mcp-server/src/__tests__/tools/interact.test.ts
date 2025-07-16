import {
  describe,
  it,
  expect,
  jest,
  beforeEach,
  afterEach,
} from "@jest/globals";
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { registerInteractTools } from "../../tools/interact.js";
import { VirtuosoCliWrapper } from "../../cli-wrapper.js";
import { z } from "zod";

// Mock the CLI wrapper
jest.mock("../../cli-wrapper.js");

describe("Interact Tools Tests", () => {
  let server: Server;
  let mockCliWrapper: jest.Mocked<VirtuosoCliWrapper>;
  let handlers: Map<string, Function>;

  beforeEach(() => {
    // Create mock server
    server = new Server(
      { name: "test-server", version: "1.0.0" },
      { capabilities: { tools: {} } },
    );

    // Create handlers map to capture registered handlers
    handlers = new Map();
    server.setRequestHandler = jest.fn((schema, handler) => {
      const schemaName = (schema as any).parse ? "tools/call" : "tools/list";
      handlers.set(schemaName, handler);
    }) as any;

    // Create mock CLI wrapper
    mockCliWrapper = new VirtuosoCliWrapper(
      "mock-path",
    ) as jest.Mocked<VirtuosoCliWrapper>;
    mockCliWrapper.execute = jest.fn();
    mockCliWrapper.getContext = jest.fn().mockReturnValue({ position: 1 });
    mockCliWrapper.updateContext = jest.fn();

    // Register the interact tools
    registerInteractTools(server, mockCliWrapper);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe("Tool Registration", () => {
    it("should register interact tool in tools list", async () => {
      const listHandler = handlers.get("tools/list");
      expect(listHandler).toBeDefined();

      const result = await listHandler({});
      const interactTool = result.tools.find(
        (tool: any) => tool.name === "virtuoso_interact",
      );

      expect(interactTool).toBeDefined();
      expect(interactTool.description).toContain("Perform user interactions");
      expect(interactTool.inputSchema.properties.action.enum).toEqual([
        "click",
        "double-click",
        "right-click",
        "hover",
        "write",
        "key",
      ]);
    });

    it("should have correct schema for all action types", async () => {
      const listHandler = handlers.get("tools/list");
      const result = await listHandler({});
      const interactTool = result.tools.find(
        (tool: any) => tool.name === "virtuoso_interact",
      );

      const schema = interactTool.inputSchema;

      // Required fields
      expect(schema.required).toContain("action");

      // Optional fields
      expect(schema.properties).toHaveProperty("selector");
      expect(schema.properties).toHaveProperty("text");
      expect(schema.properties).toHaveProperty("key");
      expect(schema.properties).toHaveProperty("checkpoint");
      expect(schema.properties).toHaveProperty("position");
      expect(schema.properties).toHaveProperty("variable");
      expect(schema.properties).toHaveProperty("elementType");
      expect(schema.properties).toHaveProperty("clickPosition");
      expect(schema.properties).toHaveProperty("clear");
      expect(schema.properties).toHaveProperty("delay");
      expect(schema.properties).toHaveProperty("target");
      expect(schema.properties).toHaveProperty("modifiers");
    });
  });

  describe("Click Actions", () => {
    it("should handle basic click action", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123", type: "UI_INTERACTION" },
        raw: JSON.stringify({ id: "step-123", type: "UI_INTERACTION" }),
      });

      const result = await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Submit button",
            checkpoint: "123456",
            position: 1,
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "interact",
        ["click", "Submit button"],
        expect.objectContaining({
          checkpoint: "123456",
          position: 1,
        }),
      );
      expect(result.content[0].text).toContain("Successfully executed");
    });

    it("should handle click with additional options", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Submit button",
            checkpoint: "123456",
            variable: "buttonText",
            elementType: "BUTTON",
            clickPosition: "TOP_LEFT",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "interact",
        ["click", "Submit button"],
        expect.objectContaining({
          checkpoint: "123456",
          variable: "buttonText",
          "element-type": "BUTTON",
          position: "TOP_LEFT",
        }),
      );
    });

    it("should handle double-click action", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "double-click",
            selector: "File item",
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "interact",
        ["double-click", "File item"],
        expect.objectContaining({
          checkpoint: "123456",
        }),
      );
    });

    it("should handle right-click action", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "right-click",
            selector: "Context menu target",
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "interact",
        ["right-click", "Context menu target"],
        expect.objectContaining({
          checkpoint: "123456",
        }),
      );
    });
  });

  describe("Write Action", () => {
    it("should handle write action with text", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "write",
            selector: "Email input",
            text: "test@example.com",
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "interact",
        ["write", "Email input", "test@example.com"],
        expect.objectContaining({
          checkpoint: "123456",
        }),
      );
    });

    it("should handle write action with clear option", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "write",
            selector: "Username",
            text: "newuser",
            clear: true,
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "interact",
        ["write", "Username", "newuser"],
        expect.objectContaining({
          checkpoint: "123456",
          clear: true,
        }),
      );
    });

    it("should handle write action with delay", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "write",
            selector: "Search field",
            text: "search query",
            delay: 500,
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "interact",
        ["write", "Search field", "search query"],
        expect.objectContaining({
          checkpoint: "123456",
          delay: 500,
        }),
      );
    });

    it("should validate write action requires text", async () => {
      const callHandler = handlers.get("tools/call");

      await expect(
        callHandler({
          params: {
            name: "virtuoso_interact",
            arguments: {
              action: "write",
              selector: "Email input",
              // Missing text
              checkpoint: "123456",
            },
          },
        }),
      ).rejects.toThrow();
    });
  });

  describe("Hover Action", () => {
    it("should handle hover action", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "hover",
            selector: "Dropdown menu",
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "interact",
        ["hover", "Dropdown menu"],
        expect.objectContaining({
          checkpoint: "123456",
        }),
      );
    });

    it("should handle hover with position", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "hover",
            selector: "Tooltip trigger",
            clickPosition: "CENTER",
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "interact",
        ["hover", "Tooltip trigger"],
        expect.objectContaining({
          checkpoint: "123456",
          position: "CENTER",
        }),
      );
    });
  });

  describe("Key Action", () => {
    it("should handle key action with single key", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "key",
            key: "Enter",
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "interact",
        ["key", "Enter"],
        expect.objectContaining({
          checkpoint: "123456",
        }),
      );
    });

    it("should handle key action with target", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "key",
            key: "Tab",
            target: 'input[type="text"]',
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "interact",
        ["key", "Tab"],
        expect.objectContaining({
          checkpoint: "123456",
          target: 'input[type="text"]',
        }),
      );
    });

    it("should handle key action with modifiers", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "key",
            key: "a",
            modifiers: ["ctrl"],
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "interact",
        ["key", "ctrl+a"],
        expect.objectContaining({
          checkpoint: "123456",
        }),
      );
    });

    it("should handle key action with multiple modifiers", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "key",
            key: "s",
            modifiers: ["ctrl", "shift"],
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "interact",
        ["key", "ctrl+shift+s"],
        expect.objectContaining({
          checkpoint: "123456",
        }),
      );
    });

    it("should validate key action requires key", async () => {
      const callHandler = handlers.get("tools/call");

      await expect(
        callHandler({
          params: {
            name: "virtuoso_interact",
            arguments: {
              action: "key",
              // Missing key
              checkpoint: "123456",
            },
          },
        }),
      ).rejects.toThrow();
    });
  });

  describe("Input Validation", () => {
    it("should reject invalid action types", async () => {
      const callHandler = handlers.get("tools/call");

      await expect(
        callHandler({
          params: {
            name: "virtuoso_interact",
            arguments: {
              action: "invalid-action",
              selector: "Something",
              checkpoint: "123456",
            },
          },
        }),
      ).rejects.toThrow();
    });

    it("should validate click position values", async () => {
      const callHandler = handlers.get("tools/call");

      await expect(
        callHandler({
          params: {
            name: "virtuoso_interact",
            arguments: {
              action: "click",
              selector: "Button",
              clickPosition: "INVALID_POSITION",
              checkpoint: "123456",
            },
          },
        }),
      ).rejects.toThrow();
    });

    it("should validate modifier values", async () => {
      const callHandler = handlers.get("tools/call");

      await expect(
        callHandler({
          params: {
            name: "virtuoso_interact",
            arguments: {
              action: "key",
              key: "a",
              modifiers: ["invalid-modifier"],
              checkpoint: "123456",
            },
          },
        }),
      ).rejects.toThrow();
    });

    it("should sanitize input strings", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "write",
            selector: "Input field",
            text: '<script>alert("xss")</script>',
            checkpoint: "123456",
          },
        },
      });

      // The sanitization happens in the CLI wrapper
      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "interact",
        ["write", "Input field", '<script>alert("xss")</script>'],
        expect.any(Object),
      );
    });
  });

  describe("Error Handling", () => {
    it("should handle CLI execution errors", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: false,
        error: "Element not found",
        raw: "",
      });

      const result = await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Non-existent button",
            checkpoint: "123456",
          },
        },
      });

      expect(result.content[0].text).toContain("Error");
      expect(result.content[0].text).toContain("Element not found");
      expect(result.isError).toBe(true);
    });

    it("should handle CLI exceptions", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockRejectedValue(new Error("CLI crashed"));

      const result = await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Button",
            checkpoint: "123456",
          },
        },
      });

      expect(result.content[0].text).toContain("Error");
      expect(result.content[0].text).toContain("CLI crashed");
      expect(result.isError).toBe(true);
    });
  });

  describe("Context Integration", () => {
    it("should use session context when checkpoint not provided", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "789",
        position: 3,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Button",
            // No checkpoint provided
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "interact",
        ["click", "Button"],
        expect.objectContaining({
          checkpoint: "789",
          position: 3,
        }),
      );
    });

    it("should auto-increment position after successful action", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.getContext.mockReturnValue({ position: 5 });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Button",
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.updateContext).toHaveBeenCalledWith({
        position: 6,
      });
    });

    it("should not increment position on error", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: false,
        error: "Failed",
        raw: "",
      });

      await callHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Button",
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.updateContext).not.toHaveBeenCalled();
    });
  });
});
