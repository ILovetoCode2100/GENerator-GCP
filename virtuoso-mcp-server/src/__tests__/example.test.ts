import { jest } from "@jest/globals";

describe("Example Test Suite", () => {
  test("should pass basic test", () => {
    expect(1 + 1).toBe(2);
  });

  test("should work with Jest globals", () => {
    const mockFn = jest.fn();
    mockFn("test");

    expect(mockFn).toHaveBeenCalledWith("test");
    expect(mockFn).toHaveBeenCalledTimes(1);
  });

  test("should have access to test helpers", () => {
    const mockResponse = global.testHelpers.createMockResponse({ id: 1 });

    expect(mockResponse).toEqual({
      success: true,
      data: { id: 1 },
      raw: JSON.stringify({ id: 1 }),
    });
  });

  test("should handle async operations", async () => {
    const asyncFunction = async () => {
      return new Promise((resolve) => {
        setTimeout(() => resolve("done"), 10);
      });
    };

    const result = await asyncFunction();
    expect(result).toBe("done");
  });
});
