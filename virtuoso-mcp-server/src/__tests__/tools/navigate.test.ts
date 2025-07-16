import {
  describe,
  it,
  expect,
  jest,
  beforeEach,
  afterEach,
} from "@jest/globals";
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { registerNavigateTools } from "../../tools/navigate.js";
import { VirtuosoCliWrapper } from "../../cli-wrapper.js";

// Mock the CLI wrapper
jest.mock("../../cli-wrapper.js");

describe("Navigate Tools Tests", () => {
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

    // Register the navigate tools
    registerNavigateTools(server, mockCliWrapper);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe("Tool Registration", () => {
    it("should register navigate tool in tools list", async () => {
      const listHandler = handlers.get("tools/list");
      expect(listHandler).toBeDefined();

      const result = await listHandler({});
      const navigateTool = result.tools.find(
        (tool: any) => tool.name === "virtuoso_navigate",
      );

      expect(navigateTool).toBeDefined();
      expect(navigateTool.description).toContain("Navigate and scroll");
      expect(navigateTool.inputSchema.properties.action.enum).toEqual([
        "to",
        "scroll-to",
        "scroll-top",
        "scroll-bottom",
        "scroll-element",
      ]);
    });

    it("should have correct schema for navigate actions", async () => {
      const listHandler = handlers.get("tools/list");
      const result = await listHandler({});
      const navigateTool = result.tools.find(
        (tool: any) => tool.name === "virtuoso_navigate",
      );

      const schema = navigateTool.inputSchema;

      // Required fields
      expect(schema.required).toContain("action");

      // Properties
      expect(schema.properties).toHaveProperty("url");
      expect(schema.properties).toHaveProperty("selector");
      expect(schema.properties).toHaveProperty("checkpoint");
      expect(schema.properties).toHaveProperty("position");
      expect(schema.properties).toHaveProperty("newTab");
      expect(schema.properties).toHaveProperty("scrollAmount");
      expect(schema.properties).toHaveProperty("direction");
    });
  });

  describe("Navigate To URL", () => {
    it("should handle navigate to URL", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123", type: "NAVIGATION" },
        raw: JSON.stringify({ id: "step-123", type: "NAVIGATION" }),
      });

      const result = await callHandler({
        params: {
          name: "virtuoso_navigate",
          arguments: {
            action: "to",
            url: "https://example.com",
            checkpoint: "123456",
            position: 1,
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "navigate",
        ["to", "https://example.com"],
        expect.objectContaining({
          checkpoint: "123456",
          position: 1,
        }),
      );
      expect(result.content[0].text).toContain("Successfully executed");
    });

    it("should handle navigate with new tab option", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_navigate",
          arguments: {
            action: "to",
            url: "https://example.com",
            newTab: true,
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "navigate",
        ["to", "https://example.com"],
        expect.objectContaining({
          checkpoint: "123456",
          "new-tab": true,
        }),
      );
    });

    it("should validate URL format", async () => {
      const callHandler = handlers.get("tools/call");

      // Valid URLs should pass
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      // Test various URL formats
      const validUrls = [
        "https://example.com",
        "http://example.com",
        "https://example.com/path",
        "https://example.com/path?query=value",
        "https://example.com:8080",
        "https://subdomain.example.com",
      ];

      for (const url of validUrls) {
        await expect(
          callHandler?.({
            params: {
              name: "virtuoso_navigate",
              arguments: {
                action: "to",
                url,
                checkpoint: "123456",
              },
            },
          }),
        ).resolves.toBeDefined();
      }
    });

    it("should reject invalid URL formats", async () => {
      const callHandler = handlers.get("tools/call");

      const invalidUrls = [
        "not a url",
        "ftp://example.com", // Only http/https supported
        "javascript:alert(1)", // XSS attempt
        "file:///etc/passwd", // File protocol
        "example.com", // Missing protocol
        "", // Empty URL
      ];

      for (const url of invalidUrls) {
        await expect(
          callHandler?.({
            params: {
              name: "virtuoso_navigate",
              arguments: {
                action: "to",
                url,
                checkpoint: "123456",
              },
            },
          }),
        ).rejects.toThrow();
      }
    });

    it("should require URL for navigate to action", async () => {
      const callHandler = handlers.get("tools/call");

      await expect(
        callHandler?.({
          params: {
            name: "virtuoso_navigate",
            arguments: {
              action: "to",
              // Missing URL
              checkpoint: "123456",
            },
          },
        }),
      ).rejects.toThrow();
    });
  });

  describe("Scroll To Element", () => {
    it("should handle scroll to element", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_navigate",
          arguments: {
            action: "scroll-to",
            selector: "#target-element",
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "navigate",
        ["scroll-to", "#target-element"],
        expect.objectContaining({
          checkpoint: "123456",
        }),
      );
    });

    it("should require selector for scroll-to action", async () => {
      const callHandler = handlers.get("tools/call");

      await expect(
        callHandler?.({
          params: {
            name: "virtuoso_navigate",
            arguments: {
              action: "scroll-to",
              // Missing selector
              checkpoint: "123456",
            },
          },
        }),
      ).rejects.toThrow();
    });
  });

  describe("Scroll Top/Bottom", () => {
    it("should handle scroll to top", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_navigate",
          arguments: {
            action: "scroll-top",
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "navigate",
        ["scroll-top"],
        expect.objectContaining({
          checkpoint: "123456",
        }),
      );
    });

    it("should handle scroll to bottom", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_navigate",
          arguments: {
            action: "scroll-bottom",
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "navigate",
        ["scroll-bottom"],
        expect.objectContaining({
          checkpoint: "123456",
        }),
      );
    });
  });

  describe("Scroll Element", () => {
    it("should handle scroll element up", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_navigate",
          arguments: {
            action: "scroll-element",
            selector: ".scrollable-container",
            direction: "up",
            scrollAmount: 100,
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "navigate",
        ["scroll-element", ".scrollable-container", "up", "100"],
        expect.objectContaining({
          checkpoint: "123456",
        }),
      );
    });

    it("should handle scroll element down", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_navigate",
          arguments: {
            action: "scroll-element",
            selector: ".scrollable-container",
            direction: "down",
            scrollAmount: 200,
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "navigate",
        ["scroll-element", ".scrollable-container", "down", "200"],
        expect.objectContaining({
          checkpoint: "123456",
        }),
      );
    });

    it("should use default scroll amount if not provided", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_navigate",
          arguments: {
            action: "scroll-element",
            selector: ".scrollable-container",
            direction: "down",
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "navigate",
        ["scroll-element", ".scrollable-container", "down", "100"], // Default 100
        expect.objectContaining({
          checkpoint: "123456",
        }),
      );
    });

    it("should require selector for scroll-element", async () => {
      const callHandler = handlers.get("tools/call");

      await expect(
        callHandler?.({
          params: {
            name: "virtuoso_navigate",
            arguments: {
              action: "scroll-element",
              // Missing selector
              direction: "up",
              checkpoint: "123456",
            },
          },
        }),
      ).rejects.toThrow();
    });

    it("should require direction for scroll-element", async () => {
      const callHandler = handlers.get("tools/call");

      await expect(
        callHandler?.({
          params: {
            name: "virtuoso_navigate",
            arguments: {
              action: "scroll-element",
              selector: ".container",
              // Missing direction
              checkpoint: "123456",
            },
          },
        }),
      ).rejects.toThrow();
    });

    it("should validate direction values", async () => {
      const callHandler = handlers.get("tools/call");

      await expect(
        callHandler?.({
          params: {
            name: "virtuoso_navigate",
            arguments: {
              action: "scroll-element",
              selector: ".container",
              direction: "sideways", // Invalid direction
              checkpoint: "123456",
            },
          },
        }),
      ).rejects.toThrow();
    });
  });

  describe("Error Handling", () => {
    it("should handle CLI execution errors", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: false,
        error: "Navigation failed: timeout",
        raw: "",
      });

      const result = await callHandler({
        params: {
          name: "virtuoso_navigate",
          arguments: {
            action: "to",
            url: "https://unreachable.com",
            checkpoint: "123456",
          },
        },
      });

      expect(result.content[0].text).toContain("Error");
      expect(result.content[0].text).toContain("Navigation failed: timeout");
      expect(result.isError).toBe(true);
    });

    it("should handle CLI exceptions", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockRejectedValue(
        new Error("CLI process crashed"),
      );

      const result = await callHandler({
        params: {
          name: "virtuoso_navigate",
          arguments: {
            action: "to",
            url: "https://example.com",
            checkpoint: "123456",
          },
        },
      });

      expect(result.content[0].text).toContain("Error");
      expect(result.content[0].text).toContain("CLI process crashed");
      expect(result.isError).toBe(true);
    });

    it("should handle element not found errors", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.execute.mockResolvedValue({
        success: false,
        error: "Element not found: #non-existent",
        raw: "",
      });

      const result = await callHandler({
        params: {
          name: "virtuoso_navigate",
          arguments: {
            action: "scroll-to",
            selector: "#non-existent",
            checkpoint: "123456",
          },
        },
      });

      expect(result.content[0].text).toContain("Error");
      expect(result.content[0].text).toContain("Element not found");
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
          name: "virtuoso_navigate",
          arguments: {
            action: "to",
            url: "https://example.com",
            // No checkpoint provided
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "navigate",
        ["to", "https://example.com"],
        expect.objectContaining({
          checkpoint: "789",
          position: 3,
        }),
      );
    });

    it("should auto-increment position after successful navigation", async () => {
      const callHandler = handlers.get("tools/call");
      mockCliWrapper.getContext.mockReturnValue({ position: 5 });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-123" },
        raw: JSON.stringify({ id: "step-123" }),
      });

      await callHandler({
        params: {
          name: "virtuoso_navigate",
          arguments: {
            action: "to",
            url: "https://example.com",
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
          name: "virtuoso_navigate",
          arguments: {
            action: "to",
            url: "https://example.com",
            checkpoint: "123456",
          },
        },
      });

      expect(mockCliWrapper.updateContext).not.toHaveBeenCalled();
    });
  });

  describe("Input Validation", () => {
    it("should reject invalid action types", async () => {
      const callHandler = handlers.get("tools/call");

      await expect(
        callHandler?.({
          params: {
            name: "virtuoso_navigate",
            arguments: {
              action: "invalid-action",
              checkpoint: "123456",
            },
          },
        }),
      ).rejects.toThrow();
    });

    it("should validate required parameters for each action", async () => {
      const callHandler = handlers.get("tools/call");

      // Missing URL for 'to' action
      await expect(
        callHandler?.({
          params: {
            name: "virtuoso_navigate",
            arguments: {
              action: "to",
              checkpoint: "123456",
            },
          },
        }),
      ).rejects.toThrow();

      // Missing selector for 'scroll-to' action
      await expect(
        callHandler?.({
          params: {
            name: "virtuoso_navigate",
            arguments: {
              action: "scroll-to",
              checkpoint: "123456",
            },
          },
        }),
      ).rejects.toThrow();

      // Missing selector for 'scroll-element' action
      await expect(
        callHandler?.({
          params: {
            name: "virtuoso_navigate",
            arguments: {
              action: "scroll-element",
              direction: "up",
              checkpoint: "123456",
            },
          },
        }),
      ).rejects.toThrow();
    });

    it("should validate numeric values", async () => {
      const callHandler = handlers.get("tools/call");

      await expect(
        callHandler?.({
          params: {
            name: "virtuoso_navigate",
            arguments: {
              action: "scroll-element",
              selector: ".container",
              direction: "up",
              scrollAmount: -100, // Negative value
              checkpoint: "123456",
            },
          },
        }),
      ).rejects.toThrow();

      await expect(
        callHandler?.({
          params: {
            name: "virtuoso_navigate",
            arguments: {
              action: "scroll-element",
              selector: ".container",
              direction: "up",
              scrollAmount: 0, // Zero value
              checkpoint: "123456",
            },
          },
        }),
      ).rejects.toThrow();
    });
  });
});
