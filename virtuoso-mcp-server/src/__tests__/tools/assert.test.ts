import { jest } from "@jest/globals";
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";
import { registerAssertTools } from "../../tools/assert.js";
import { VirtuosoCliWrapper } from "../../cli-wrapper.js";

// Mock dependencies
jest.mock("../../cli-wrapper.js");
jest.mock("../../utils/formatting.js", () => ({
  formatToolResponse: jest.fn((result, context) => ({
    success: result.success,
    error: result.error,
    data: result.data,
  })),
}));
jest.mock("../../utils/validation.js", () => ({
  sanitizeInput: jest.fn((input) => input),
}));

describe("Assert Tools", () => {
  let mockServer: jest.Mocked<Server>;
  let mockCli: jest.Mocked<VirtuosoCliWrapper>;
  let listToolsHandler: Function;
  let callToolHandler: Function;

  beforeEach(() => {
    jest.clearAllMocks();

    // Setup mock server
    mockServer = {
      setRequestHandler: jest.fn((schema, handler) => {
        if (schema === ListToolsRequestSchema) {
          listToolsHandler = handler;
        } else if (schema === CallToolRequestSchema) {
          callToolHandler = handler;
        }
      }),
    } as any;

    // Setup mock CLI wrapper
    mockCli = {
      execute: jest.fn(),
      getContext: jest.fn().mockReturnValue({
        checkpointId: "12345",
        position: 1,
      }),
    } as any;

    // Register the tools
    registerAssertTools(mockServer, mockCli);
  });

  describe("Tool Registration", () => {
    test("should register list tools handler", () => {
      expect(mockServer.setRequestHandler).toHaveBeenCalledWith(
        ListToolsRequestSchema,
        expect.any(Function),
      );
    });

    test("should register call tool handler", () => {
      expect(mockServer.setRequestHandler).toHaveBeenCalledWith(
        CallToolRequestSchema,
        expect.any(Function),
      );
    });

    test("should return virtuoso_assert tool in list", async () => {
      const result = await listToolsHandler({});

      expect(result.tools).toHaveLength(1);
      expect(result.tools[0]).toEqual({
        name: "virtuoso_assert",
        description: expect.stringContaining("Create assertion steps"),
        inputSchema: expect.objectContaining({
          type: "object",
          properties: expect.objectContaining({
            type: expect.objectContaining({
              enum: expect.arrayContaining(["exists", "not-exists", "equals"]),
            }),
          }),
        }),
      });
    });
  });

  describe("Assert Exists", () => {
    test("should create exists assertion", async () => {
      mockCli.execute.mockResolvedValue({
        success: true,
        data: { stepId: "step-123" },
      });

      const result = await callToolHandler({
        params: {
          name: "virtuoso_assert",
          arguments: {
            type: "exists",
            element: "Login button",
          },
        },
      });

      expect(mockCli.execute).toHaveBeenCalledWith(
        ["assert", "exists", "Login button"],
        { checkpoint: undefined, position: undefined },
      );

      expect(result.content[0]).toEqual({
        type: "text",
        text: expect.stringContaining(
          '✅ Created assertion: Element "Login button" exists',
        ),
      });
    });

    test("should use provided checkpoint and position", async () => {
      mockCli.execute.mockResolvedValue({
        success: true,
        data: { stepId: "step-123" },
      });

      await callToolHandler({
        params: {
          name: "virtuoso_assert",
          arguments: {
            type: "exists",
            element: "Submit",
            checkpoint: "67890",
            position: 5,
          },
        },
      });

      expect(mockCli.execute).toHaveBeenCalledWith(
        ["assert", "exists", "Submit"],
        { checkpoint: "67890", position: 5 },
      );
    });

    test("should handle missing element for exists assertion", async () => {
      const result = await callToolHandler({
        params: {
          name: "virtuoso_assert",
          arguments: {
            type: "exists",
          },
        },
      });

      expect(result.isError).toBe(true);
      expect(result.content[0].text).toContain(
        "Element selector required for exists assertion",
      );
    });
  });

  describe("Assert Equals", () => {
    test("should create equals assertion", async () => {
      mockCli.execute.mockResolvedValue({
        success: true,
        data: { stepId: "step-456" },
      });

      const result = await callToolHandler({
        params: {
          name: "virtuoso_assert",
          arguments: {
            type: "equals",
            element: "Username",
            value: "john@example.com",
          },
        },
      });

      expect(mockCli.execute).toHaveBeenCalledWith(
        ["assert", "equals", "Username", "john@example.com"],
        expect.any(Object),
      );

      expect(result.content[0].text).toContain(
        '✅ Created assertion: "Username" equals "john@example.com"',
      );
    });

    test("should handle missing value for equals assertion", async () => {
      const result = await callToolHandler({
        params: {
          name: "virtuoso_assert",
          arguments: {
            type: "equals",
            element: "Username",
          },
        },
      });

      expect(result.isError).toBe(true);
      expect(result.content[0].text).toContain(
        "Element and value required for equals assertion",
      );
    });
  });

  describe("Numeric Assertions", () => {
    test.each(["gt", "gte", "lt", "lte"])(
      "should create %s assertion",
      async (type) => {
        mockCli.execute.mockResolvedValue({
          success: true,
          data: { stepId: "step-789" },
        });

        const result = await callToolHandler({
          params: {
            name: "virtuoso_assert",
            arguments: {
              type,
              element: "Price",
              value: "100",
            },
          },
        });

        expect(mockCli.execute).toHaveBeenCalledWith(
          ["assert", type, "Price", "100"],
          expect.any(Object),
        );

        expect(result.content[0].text).toContain("✅ Created assertion:");
      },
    );

    test("should validate numeric value", async () => {
      const result = await callToolHandler({
        params: {
          name: "virtuoso_assert",
          arguments: {
            type: "gt",
            element: "Price",
            value: "not-a-number",
          },
        },
      });

      expect(result.isError).toBe(true);
      expect(result.content[0].text).toContain(
        "Value must be numeric for gt assertion",
      );
    });
  });

  describe("Pattern Matching", () => {
    test("should create matches assertion", async () => {
      mockCli.execute.mockResolvedValue({
        success: true,
        data: { stepId: "step-999" },
      });

      const result = await callToolHandler({
        params: {
          name: "virtuoso_assert",
          arguments: {
            type: "matches",
            element: "Email",
            value: "^[a-zA-Z0-9+_.-]+@[a-zA-Z0-9.-]+$",
          },
        },
      });

      expect(mockCli.execute).toHaveBeenCalledWith(
        ["assert", "matches", "Email", "^[a-zA-Z0-9+_.-]+@[a-zA-Z0-9.-]+$"],
        expect.any(Object),
      );

      expect(result.content[0].text).toContain(
        '✅ Created assertion: "Email" matches pattern',
      );
    });
  });

  describe("Variable Assertion", () => {
    test("should create variable assertion", async () => {
      mockCli.execute.mockResolvedValue({
        success: true,
        data: { stepId: "step-111" },
      });

      const result = await callToolHandler({
        params: {
          name: "virtuoso_assert",
          arguments: {
            type: "variable",
            variable: "userEmail",
            value: "test@example.com",
          },
        },
      });

      expect(mockCli.execute).toHaveBeenCalledWith(
        ["assert", "variable", "userEmail", "test@example.com"],
        expect.any(Object),
      );

      expect(result.content[0].text).toContain(
        '✅ Created assertion: Variable "userEmail" equals "test@example.com"',
      );
    });

    test("should handle missing variable name", async () => {
      const result = await callToolHandler({
        params: {
          name: "virtuoso_assert",
          arguments: {
            type: "variable",
            value: "test@example.com",
          },
        },
      });

      expect(result.isError).toBe(true);
      expect(result.content[0].text).toContain(
        "Variable name and expected value required",
      );
    });
  });

  describe("Error Handling", () => {
    test("should handle unknown tool name", async () => {
      await expect(
        callToolHandler({
          params: {
            name: "unknown_tool",
            arguments: {},
          },
        }),
      ).rejects.toThrow("Unknown tool: unknown_tool");
    });

    test("should handle CLI execution failure", async () => {
      mockCli.execute.mockResolvedValue({
        success: false,
        error: "CLI command failed",
      });

      const result = await callToolHandler({
        params: {
          name: "virtuoso_assert",
          arguments: {
            type: "exists",
            element: "Button",
          },
        },
      });

      expect(result.isError).toBe(true);
      expect(result.content[0].text).toContain(
        "❌ Assertion failed: CLI command failed",
      );
    });

    test("should handle validation errors", async () => {
      const result = await callToolHandler({
        params: {
          name: "virtuoso_assert",
          arguments: {
            type: "invalid-type",
            element: "Button",
          },
        },
      });

      expect(result.isError).toBe(true);
      expect(result.content[0].text).toContain("❌ Error:");
    });
  });

  describe("Special Assertions", () => {
    test.each(["checked", "selected"])(
      "should create %s assertion",
      async (type) => {
        mockCli.execute.mockResolvedValue({
          success: true,
          data: { stepId: "step-222" },
        });

        const result = await callToolHandler({
          params: {
            name: "virtuoso_assert",
            arguments: {
              type,
              element: "Checkbox",
            },
          },
        });

        expect(mockCli.execute).toHaveBeenCalledWith(
          ["assert", type, "Checkbox"],
          expect.any(Object),
        );

        expect(result.content[0].text).toContain(
          `✅ Created assertion: "Checkbox" is ${type}`,
        );
      },
    );

    test("should create not-exists assertion", async () => {
      mockCli.execute.mockResolvedValue({
        success: true,
        data: { stepId: "step-333" },
      });

      const result = await callToolHandler({
        params: {
          name: "virtuoso_assert",
          arguments: {
            type: "not-exists",
            element: "Error message",
          },
        },
      });

      expect(mockCli.execute).toHaveBeenCalledWith(
        ["assert", "not-exists", "Error message"],
        expect.any(Object),
      );

      expect(result.content[0].text).toContain(
        '✅ Created assertion: Element "Error message" does not exist',
      );
    });

    test("should create not-equals assertion", async () => {
      mockCli.execute.mockResolvedValue({
        success: true,
        data: { stepId: "step-444" },
      });

      const result = await callToolHandler({
        params: {
          name: "virtuoso_assert",
          arguments: {
            type: "not-equals",
            element: "Status",
            value: "Error",
          },
        },
      });

      expect(mockCli.execute).toHaveBeenCalledWith(
        ["assert", "not-equals", "Status", "Error"],
        expect.any(Object),
      );

      expect(result.content[0].text).toContain(
        '✅ Created assertion: "Status" not equals "Error"',
      );
    });
  });
});
