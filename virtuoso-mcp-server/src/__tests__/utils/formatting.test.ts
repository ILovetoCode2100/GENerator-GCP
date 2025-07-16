import { jest } from "@jest/globals";
import {
  formatToolResponse,
  formatStepDescription,
  formatList,
  formatErrorDetails,
  formatCliArgs,
  formatCheckpointContext,
} from "../../utils/formatting.js";

describe("Formatting Utilities", () => {
  describe("formatToolResponse", () => {
    test("should format error responses", () => {
      const result = formatToolResponse({
        success: false,
        error: "Command failed",
      });

      expect(result).toEqual({
        success: false,
        error: "Command failed",
        context: undefined,
      });
    });

    test("should format successful responses with data", () => {
      const result = formatToolResponse({
        success: true,
        data: { message: "Success" },
      });

      expect(result).toEqual({
        success: true,
        data: { message: "Success" },
        context: undefined,
      });
    });

    test("should include context when provided", () => {
      const context = { checkpointId: "12345" };
      const result = formatToolResponse(
        {
          success: true,
          data: { id: 1 },
        },
        context,
      );

      expect(result.context).toEqual(context);
    });

    test("should format step creation responses", () => {
      const result = formatToolResponse({
        success: true,
        data: {
          step_id: "step-123",
          type: "assert",
          position: 1,
          description: "Test assertion",
        },
      });

      expect(result.data).toEqual({
        stepId: "step-123",
        type: "assert",
        position: 1,
        description: "Test assertion",
        status: "created",
      });
    });

    test("should handle stepId variant", () => {
      const result = formatToolResponse({
        success: true,
        data: {
          stepId: "step-456",
          type: "click",
        },
      });

      expect(result.data.stepId).toBe("step-456");
      expect(result.data.status).toBe("created");
    });

    test("should format array responses", () => {
      const items = [{ id: 1 }, { id: 2 }];
      const result = formatToolResponse({
        success: true,
        data: items,
      });

      expect(result.data).toEqual({
        items: items,
        count: 2,
      });
    });

    test("should format string responses", () => {
      const result = formatToolResponse({
        success: true,
        data: "Command completed",
      });

      expect(result.data).toEqual({
        message: "Command completed",
      });
    });

    test("should pass through other data types unchanged", () => {
      const complexData = {
        nested: { value: 123 },
        array: [1, 2, 3],
      };

      const result = formatToolResponse({
        success: true,
        data: complexData,
      });

      expect(result.data).toEqual(complexData);
    });

    test("should handle responses without explicit success flag", () => {
      const result = formatToolResponse({
        data: { id: 1 },
      });

      expect(result).toEqual({
        success: true,
        data: { id: 1 },
        context: undefined,
      });
    });
  });

  describe("formatStepDescription", () => {
    test("should format basic step", () => {
      const step = {
        id: 123,
        type: "assert",
        position: 1,
      };

      expect(formatStepDescription(step)).toBe("Step 123: assert");
    });

    test("should include description when available", () => {
      const step = {
        id: 123,
        type: "assert",
        position: 1,
        description: "Check login button exists",
      };

      expect(formatStepDescription(step)).toBe(
        "Step 123: assert - Check login button exists",
      );
    });

    test("should include position when available", () => {
      const step = {
        id: 123,
        type: "assert",
        position: 5,
      };

      expect(formatStepDescription(step)).toBe(
        "Step 123: assert (Position: 5)",
      );
    });

    test("should include all fields when available", () => {
      const step = {
        id: 123,
        type: "assert",
        description: "Check element",
        position: 3,
      };

      expect(formatStepDescription(step)).toBe(
        "Step 123: assert - Check element (Position: 3)",
      );
    });
  });

  describe("formatList", () => {
    test("should return message for empty list", () => {
      expect(formatList([], ["id", "name"])).toBe("No items found");
    });

    test("should format list with headers", () => {
      const items = [
        { id: 1, name: "Test 1" },
        { id: 2, name: "Test 2" },
      ];

      const result = formatList(items, ["id", "name"]);
      const lines = result.split("\n");

      expect(lines[0]).toBe("id | name");
      expect(lines[1]).toBe("-".repeat("id | name".length));
      expect(lines[2]).toBe("1 | Test 1");
      expect(lines[3]).toBe("2 | Test 2");
    });

    test("should handle missing fields with dash", () => {
      const items = [
        { id: 1, name: "Test" },
        { id: 2 }, // missing name
      ];

      const result = formatList(items, ["id", "name"]);
      const lines = result.split("\n");

      expect(lines[3]).toBe("2 | -");
    });

    test("should convert non-string values", () => {
      const items = [{ id: 1, active: true, count: 42 }];

      const result = formatList(items, ["id", "active", "count"]);
      const lines = result.split("\n");

      expect(lines[2]).toBe("1 | true | 42");
    });
  });

  describe("formatErrorDetails", () => {
    const originalEnv = process.env.DEBUG;

    afterEach(() => {
      process.env.DEBUG = originalEnv;
    });

    test("should format basic error message", () => {
      const error = { message: "Something went wrong" };
      expect(formatErrorDetails(error)).toBe("Error: Something went wrong");
    });

    test("should include error code", () => {
      const error = {
        message: "Connection failed",
        code: "ECONNREFUSED",
      };

      const result = formatErrorDetails(error);
      expect(result).toContain("Error: Connection failed");
      expect(result).toContain("Code: ECONNREFUSED");
    });

    test("should include stack trace in debug mode", () => {
      process.env.DEBUG = "true";
      const error = {
        message: "Test error",
        stack: "Error: Test error\n    at test.js:10:5",
      };

      const result = formatErrorDetails(error);
      expect(result).toContain("Stack:");
      expect(result).toContain("at test.js:10:5");
    });

    test("should not include stack trace when not in debug mode", () => {
      process.env.DEBUG = "false";
      const error = {
        message: "Test error",
        stack: "Error: Test error\n    at test.js:10:5",
      };

      const result = formatErrorDetails(error);
      expect(result).not.toContain("Stack:");
    });

    test("should handle errors without message", () => {
      const error = { code: "UNKNOWN" };
      expect(formatErrorDetails(error)).toBe("Code: UNKNOWN");
    });
  });

  describe("formatCliArgs", () => {
    test("should format simple arguments", () => {
      const args = ["assert", "exists", "button"];
      expect(formatCliArgs(args)).toBe("assert exists button");
    });

    test("should quote arguments with spaces", () => {
      const args = ["assert", "exists", "Login button"];
      expect(formatCliArgs(args)).toBe('assert exists "Login button"');
    });

    test("should handle multiple arguments with spaces", () => {
      const args = ["interact", "write", "Email field", "test@example.com"];
      expect(formatCliArgs(args)).toBe(
        'interact write "Email field" test@example.com',
      );
    });

    test("should handle empty arguments array", () => {
      expect(formatCliArgs([])).toBe("");
    });
  });

  describe("formatCheckpointContext", () => {
    test("should format string checkpoint", () => {
      expect(formatCheckpointContext("12345")).toBe("Checkpoint ID: 12345");
    });

    test("should format object checkpoint with id", () => {
      const checkpoint = { id: "12345" };
      expect(formatCheckpointContext(checkpoint)).toBe("ID: 12345");
    });

    test("should format object checkpoint with multiple fields", () => {
      const checkpoint = {
        id: "12345",
        name: "Login Test",
        url: "https://example.com",
      };

      expect(formatCheckpointContext(checkpoint)).toBe(
        "ID: 12345, Name: Login Test, URL: https://example.com",
      );
    });

    test("should handle partial checkpoint objects", () => {
      const checkpoint = {
        name: "Test Checkpoint",
      };

      expect(formatCheckpointContext(checkpoint)).toBe("Name: Test Checkpoint");
    });

    test("should handle null/undefined checkpoint", () => {
      expect(formatCheckpointContext(null)).toBe("No checkpoint context");
      expect(formatCheckpointContext(undefined)).toBe("No checkpoint context");
    });

    test("should handle empty object", () => {
      expect(formatCheckpointContext({})).toBe("No checkpoint context");
    });
  });
});
