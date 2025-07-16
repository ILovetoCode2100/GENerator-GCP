import { jest } from "@jest/globals";
import { spawn } from "child_process";
import { EventEmitter } from "events";
import { VirtuosoCliWrapper } from "../cli-wrapper.js";

// Mock child_process module
jest.mock("child_process", () => ({
  spawn: jest.fn(),
}));

describe("VirtuosoCliWrapper", () => {
  let mockSpawn: jest.MockedFunction<typeof spawn>;
  let wrapper: VirtuosoCliWrapper;

  beforeEach(() => {
    mockSpawn = spawn as jest.MockedFunction<typeof spawn>;
    jest.clearAllMocks();
    wrapper = new VirtuosoCliWrapper("/path/to/cli", "/path/to/config.yaml");
  });

  describe("Session Context Management", () => {
    test("should initialize with default context", () => {
      const context = wrapper.getContext();
      expect(context).toEqual({
        position: 1,
        checkpointId: undefined,
        journeyId: undefined,
        goalId: undefined,
      });
    });

    test("should update context correctly", () => {
      wrapper.updateContext({
        checkpointId: "12345",
        position: 5,
        journeyId: "67890",
      });

      const context = wrapper.getContext();
      expect(context).toEqual({
        position: 5,
        checkpointId: "12345",
        journeyId: "67890",
        goalId: undefined,
      });
    });

    test("should merge partial context updates", () => {
      wrapper.updateContext({ checkpointId: "12345" });
      wrapper.updateContext({ position: 3 });

      const context = wrapper.getContext();
      expect(context.checkpointId).toBe("12345");
      expect(context.position).toBe(3);
    });

    test("should return a copy of context to prevent external modifications", () => {
      const context1 = wrapper.getContext();
      context1.position = 999;

      const context2 = wrapper.getContext();
      expect(context2.position).toBe(1);
    });
  });

  describe("Command Execution", () => {
    let mockProcess: any;

    beforeEach(() => {
      // Create a mock child process
      mockProcess = new EventEmitter();
      mockProcess.stdout = new EventEmitter();
      mockProcess.stderr = new EventEmitter();
      mockProcess.stdin = { end: jest.fn() };

      mockSpawn.mockReturnValue(mockProcess as any);
    });

    test("should execute basic command with JSON output", async () => {
      const executePromise = wrapper.execute(["list-projects"]);

      // Simulate successful execution
      mockProcess.stdout.emit("data", JSON.stringify({ projects: [] }));
      mockProcess.emit("close", 0);

      const result = await executePromise;

      expect(mockSpawn).toHaveBeenCalledWith(
        "/path/to/cli",
        [
          "list-projects",
          "--output",
          "json",
          "--config",
          "/path/to/config.yaml",
        ],
        expect.objectContaining({
          env: expect.any(Object),
          stdio: ["pipe", "pipe", "pipe"],
        }),
      );

      expect(result).toEqual({
        success: true,
        data: { projects: [] },
        raw: JSON.stringify({ projects: [] }),
      });
    });

    test("should add checkpoint from session context", async () => {
      wrapper.updateContext({ checkpointId: "12345" });

      const executePromise = wrapper.execute([
        "assert",
        "exists",
        "Login button",
      ]);

      mockProcess.stdout.emit("data", JSON.stringify({ success: true }));
      mockProcess.emit("close", 0);

      await executePromise;

      expect(mockSpawn).toHaveBeenCalledWith(
        "/path/to/cli",
        expect.arrayContaining([
          "assert",
          "exists",
          "Login button",
          "--output",
          "json",
          "--config",
          "/path/to/config.yaml",
          "--checkpoint",
          "12345",
          "1", // position
        ]),
        expect.any(Object),
      );
    });

    test("should use positional checkpoint for certain commands", async () => {
      wrapper.updateContext({ checkpointId: "12345" });

      const executePromise = wrapper.execute(["interact", "click", "Submit"]);

      mockProcess.stdout.emit("data", JSON.stringify({ success: true }));
      mockProcess.emit("close", 0);

      await executePromise;

      expect(mockSpawn).toHaveBeenCalledWith(
        "/path/to/cli",
        expect.arrayContaining([
          "interact",
          "click",
          "12345",
          "Submit", // checkpoint as positional arg
          "--output",
          "json",
          "--config",
          "/path/to/config.yaml",
          "1", // position
        ]),
        expect.any(Object),
      );
    });

    test("should auto-increment position after successful step creation", async () => {
      const executePromise = wrapper.execute(["assert", "exists", "Button"], {
        checkpoint: "12345",
      });

      mockProcess.stdout.emit("data", JSON.stringify({ success: true }));
      mockProcess.emit("close", 0);

      await executePromise;

      expect(wrapper.getContext().position).toBe(2);
    });

    test("should not increment position on failed command", async () => {
      const executePromise = wrapper.execute(["assert", "exists", "Button"], {
        checkpoint: "12345",
      });

      mockProcess.stderr.emit("data", "Error: Command failed");
      mockProcess.emit("close", 1);

      await executePromise;

      expect(wrapper.getContext().position).toBe(1);
    });

    test("should handle custom options", async () => {
      const executePromise = wrapper.execute(["interact", "click", "Button"], {
        checkpoint: "12345",
        position: 5,
        variable: "buttonClicked",
        "new-tab": true,
      });

      mockProcess.stdout.emit("data", JSON.stringify({ success: true }));
      mockProcess.emit("close", 0);

      await executePromise;

      expect(mockSpawn).toHaveBeenCalledWith(
        "/path/to/cli",
        expect.arrayContaining(["--variable", "buttonClicked", "--new-tab"]),
        expect.any(Object),
      );
    });

    test("should handle non-JSON output", async () => {
      const executePromise = wrapper.execute(["list-projects"]);

      mockProcess.stdout.emit("data", "Plain text output");
      mockProcess.emit("close", 0);

      const result = await executePromise;

      expect(result).toEqual({
        success: true,
        data: { output: "Plain text output" },
        raw: "Plain text output",
      });
    });

    test("should handle command errors", async () => {
      const executePromise = wrapper.execute(["invalid-command"]);

      mockProcess.stderr.emit("data", "Unknown command: invalid-command");
      mockProcess.emit("close", 1);

      const result = await executePromise;

      expect(result).toEqual({
        success: false,
        error: "Unknown command: invalid-command",
        raw: "",
      });
    });

    test("should handle spawn errors", async () => {
      mockSpawn.mockImplementationOnce(() => {
        throw new Error("ENOENT: command not found");
      });

      const result = await wrapper.execute(["list-projects"]);

      expect(result).toEqual({
        success: false,
        error: "Failed to start CLI: ENOENT: command not found",
      });
    });

    test("should not add position for commands that do not require it", async () => {
      const executePromise = wrapper.execute(["list-projects"]);

      mockProcess.stdout.emit("data", JSON.stringify({ projects: [] }));
      mockProcess.emit("close", 0);

      await executePromise;

      const args = mockSpawn.mock.calls[0][1];
      expect(args).not.toContain("1");
      expect(args).not.toContain("--position");
    });

    test("should respect explicit position over session context", async () => {
      wrapper.updateContext({ position: 10 });

      const executePromise = wrapper.execute(["assert", "exists", "Button"], {
        checkpoint: "12345",
        position: 5,
      });

      mockProcess.stdout.emit("data", JSON.stringify({ success: true }));
      mockProcess.emit("close", 0);

      await executePromise;

      const args = mockSpawn.mock.calls[0][1];
      expect(args).toContain("5");
      expect(args).not.toContain("10");
    });
  });

  describe("Command Classification", () => {
    test("should identify step creation commands", () => {
      const stepCommands = [
        ["assert", "exists", "Element"],
        ["interact", "click", "Button"],
        ["navigate", "to", "https://example.com"],
        ["wait", "element", "#loader"],
        ["window", "resize", "1024", "768"],
      ];

      stepCommands.forEach((args) => {
        wrapper.updateContext({ position: 1 });

        const mockProcess = new EventEmitter();
        mockProcess.stdout = new EventEmitter();
        mockProcess.stderr = new EventEmitter();
        mockSpawn.mockReturnValue(mockProcess as any);

        const executePromise = wrapper.execute(args, { checkpoint: "12345" });

        mockProcess.stdout.emit("data", JSON.stringify({ success: true }));
        mockProcess.emit("close", 0);

        executePromise.then(() => {
          expect(wrapper.getContext().position).toBe(2);
        });
      });
    });

    test("should not increment position for non-step commands", () => {
      const nonStepCommands = [
        ["list-projects"],
        ["list-goals", "12345"],
        ["create-project", "Test Project"],
        ["library", "get", "7023"],
      ];

      nonStepCommands.forEach((args) => {
        wrapper.updateContext({ position: 1 });

        const mockProcess = new EventEmitter();
        mockProcess.stdout = new EventEmitter();
        mockProcess.stderr = new EventEmitter();
        mockSpawn.mockReturnValue(mockProcess as any);

        const executePromise = wrapper.execute(args);

        mockProcess.stdout.emit("data", JSON.stringify({ success: true }));
        mockProcess.emit("close", 0);

        executePromise.then(() => {
          expect(wrapper.getContext().position).toBe(1);
        });
      });
    });
  });

  describe("Multiple Command Execution", () => {
    test("should maintain position across multiple commands", async () => {
      const mockProcess = new EventEmitter();
      mockProcess.stdout = new EventEmitter();
      mockProcess.stderr = new EventEmitter();
      mockSpawn.mockReturnValue(mockProcess as any);

      wrapper.updateContext({ checkpointId: "12345" });

      // First command
      const promise1 = wrapper.execute(["assert", "exists", "Button1"]);
      mockProcess.stdout.emit("data", JSON.stringify({ success: true }));
      mockProcess.emit("close", 0);
      await promise1;

      expect(wrapper.getContext().position).toBe(2);

      // Reset mock for second command
      const mockProcess2 = new EventEmitter();
      mockProcess2.stdout = new EventEmitter();
      mockProcess2.stderr = new EventEmitter();
      mockSpawn.mockReturnValue(mockProcess2 as any);

      // Second command
      const promise2 = wrapper.execute(["interact", "click", "Button2"]);
      mockProcess2.stdout.emit("data", JSON.stringify({ success: true }));
      mockProcess2.emit("close", 0);
      await promise2;

      expect(wrapper.getContext().position).toBe(3);

      // Verify second command used position 2
      const secondCallArgs = mockSpawn.mock.calls[1][1];
      expect(secondCallArgs).toContain("2");
    });
  });
});
