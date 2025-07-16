import {
  describe,
  it,
  expect,
  jest,
  beforeEach,
  afterEach,
} from "@jest/globals";
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { VirtuosoMcpServer } from "../../server.js";
import { VirtuosoCliWrapper } from "../../cli-wrapper.js";
import path from "path";
import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Mock the CLI wrapper
jest.mock("../../cli-wrapper.js");

describe("End-to-End Integration Tests", () => {
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

    // Setup default context
    mockCliWrapper.getContext.mockReturnValue({ position: 1 });
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  describe("Complete Test Scenario", () => {
    it("should execute a complete login test flow", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");

      // Step 1: Set context for the test
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

      // Step 2: Navigate to login page
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "nav-1", type: "NAVIGATION" },
        raw: JSON.stringify({ id: "nav-1", type: "NAVIGATION" }),
      });

      let result = await callToolHandler({
        params: {
          name: "virtuoso_navigate",
          arguments: {
            action: "to",
            url: "https://example.com/login",
          },
        },
      });

      expect(result.content[0].text).toContain("Successfully executed");
      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "navigate",
        ["to", "https://example.com/login"],
        expect.objectContaining({
          checkpoint: "123456",
          position: 1,
        }),
      );

      // Position should auto-increment
      expect(mockCliWrapper.updateContext).toHaveBeenCalledWith({
        position: 2,
      });

      // Step 3: Wait for page to load
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "123456",
        position: 2,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "wait-1", type: "WAIT" },
        raw: JSON.stringify({ id: "wait-1", type: "WAIT" }),
      });

      result = await callToolHandler({
        params: {
          name: "virtuoso_wait",
          arguments: {
            action: "element",
            selector: "#login-form",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "wait",
        ["element", "#login-form"],
        expect.objectContaining({
          checkpoint: "123456",
          position: 2,
        }),
      );

      // Step 4: Enter username
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "123456",
        position: 3,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "write-1", type: "UI_INTERACTION" },
        raw: JSON.stringify({ id: "write-1", type: "UI_INTERACTION" }),
      });

      result = await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "write",
            selector: "Username field",
            text: "testuser@example.com",
            clear: true,
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "interact",
        ["write", "Username field", "testuser@example.com"],
        expect.objectContaining({
          checkpoint: "123456",
          position: 3,
          clear: true,
        }),
      );

      // Step 5: Enter password
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "123456",
        position: 4,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "write-2", type: "UI_INTERACTION" },
        raw: JSON.stringify({ id: "write-2", type: "UI_INTERACTION" }),
      });

      result = await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "write",
            selector: "Password field",
            text: "securepassword123",
            clear: true,
          },
        },
      });

      // Step 6: Click login button
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "123456",
        position: 5,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "click-1", type: "UI_INTERACTION" },
        raw: JSON.stringify({ id: "click-1", type: "UI_INTERACTION" }),
      });

      result = await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Login button",
          },
        },
      });

      // Step 7: Assert login successful
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "123456",
        position: 6,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "assert-1", type: "ASSERTION" },
        raw: JSON.stringify({ id: "assert-1", type: "ASSERTION" }),
      });

      result = await callToolHandler({
        params: {
          name: "virtuoso_assert",
          arguments: {
            action: "exists",
            selector: "Welcome message",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "assert",
        ["exists", "Welcome message"],
        expect.objectContaining({
          checkpoint: "123456",
          position: 6,
        }),
      );

      // Verify total number of calls
      expect(mockCliWrapper.execute).toHaveBeenCalledTimes(6);

      // Verify final position
      expect(mockCliWrapper.updateContext).toHaveBeenLastCalledWith({
        position: 7,
      });
    });
  });

  describe("Multi-Tool Sequences", () => {
    it("should handle form filling sequence with data storage", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");

      // Set initial context
      mockCliWrapper.updateContext.mockImplementation((context) => {
        if (context.checkpointId) {
          mockCliWrapper.getContext.mockReturnValue({
            checkpointId: context.checkpointId,
            position: context.position || 1,
          });
        } else if (context.position !== undefined) {
          const current = mockCliWrapper.getContext();
          mockCliWrapper.getContext.mockReturnValue({
            ...current,
            position: context.position,
          });
        }
      });

      await callToolHandler({
        params: {
          name: "virtuoso_set_context",
          arguments: {
            checkpointId: "789",
            position: 1,
          },
        },
      });

      // Step 1: Store existing value
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "store-1", type: "DATA_STORAGE" },
        raw: JSON.stringify({ id: "store-1", type: "DATA_STORAGE" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_data",
          arguments: {
            action: "store-text",
            selector: "Current value field",
            variable: "originalValue",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "data",
        ["store-text", "Current value field", "originalValue"],
        expect.objectContaining({
          checkpoint: "789",
          position: 1,
        }),
      );

      // Step 2: Clear and write new value
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "789",
        position: 2,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "write-1", type: "UI_INTERACTION" },
        raw: JSON.stringify({ id: "write-1", type: "UI_INTERACTION" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "write",
            selector: "Current value field",
            text: "New test value",
            clear: true,
          },
        },
      });

      // Step 3: Select from dropdown
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "789",
        position: 3,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "select-1", type: "UI_INTERACTION" },
        raw: JSON.stringify({ id: "select-1", type: "UI_INTERACTION" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_select",
          arguments: {
            action: "option",
            selector: "#country-dropdown",
            value: "United States",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "select",
        ["option", "#country-dropdown", "United States"],
        expect.objectContaining({
          checkpoint: "789",
          position: 3,
        }),
      );
    });

    it("should handle window and frame navigation", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");

      // Setup context
      await callToolHandler({
        params: {
          name: "virtuoso_set_context",
          arguments: {
            checkpointId: "456",
            position: 1,
          },
        },
      });

      // Step 1: Open link in new tab
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "nav-1", type: "NAVIGATION" },
        raw: JSON.stringify({ id: "nav-1", type: "NAVIGATION" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_navigate",
          arguments: {
            action: "to",
            url: "https://example.com/page-with-iframe",
            newTab: true,
          },
        },
      });

      // Step 2: Switch to new tab
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "456",
        position: 2,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "window-1", type: "WINDOW_OPERATION" },
        raw: JSON.stringify({ id: "window-1", type: "WINDOW_OPERATION" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_window",
          arguments: {
            action: "switch-tab",
            target: "next",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "window",
        ["switch-tab", "next"],
        expect.objectContaining({
          checkpoint: "456",
          position: 2,
        }),
      );

      // Step 3: Switch to iframe
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "456",
        position: 3,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "window-2", type: "WINDOW_OPERATION" },
        raw: JSON.stringify({ id: "window-2", type: "WINDOW_OPERATION" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_window",
          arguments: {
            action: "switch-frame",
            frameIdentifier: "#payment-iframe",
          },
        },
      });

      // Step 4: Interact within iframe
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "456",
        position: 4,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "write-1", type: "UI_INTERACTION" },
        raw: JSON.stringify({ id: "write-1", type: "UI_INTERACTION" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "write",
            selector: "Card number",
            text: "4111111111111111",
          },
        },
      });
    });
  });

  describe("Error Recovery", () => {
    it("should handle partial failures in a sequence", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");

      // Setup context
      await callToolHandler({
        params: {
          name: "virtuoso_set_context",
          arguments: {
            checkpointId: "999",
            position: 1,
          },
        },
      });

      // Step 1: Successful navigation
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "nav-1", type: "NAVIGATION" },
        raw: JSON.stringify({ id: "nav-1", type: "NAVIGATION" }),
      });

      let result = await callToolHandler({
        params: {
          name: "virtuoso_navigate",
          arguments: {
            action: "to",
            url: "https://example.com/form",
          },
        },
      });

      expect(result.content[0].text).toContain("Successfully executed");
      expect(mockCliWrapper.updateContext).toHaveBeenCalledWith({
        position: 2,
      });

      // Step 2: Failed interaction (element not found)
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "999",
        position: 2,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: false,
        error: "Element not found: Non-existent button",
        raw: "",
      });

      result = await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Non-existent button",
          },
        },
      });

      expect(result.isError).toBe(true);
      expect(result.content[0].text).toContain("Error");

      // Position should NOT increment on error
      const lastUpdateCall =
        mockCliWrapper.updateContext.mock.calls[
          mockCliWrapper.updateContext.mock.calls.length - 1
        ];
      expect(lastUpdateCall[0].position).not.toBe(3);

      // Step 3: Recovery - try alternative selector
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "click-1", type: "UI_INTERACTION" },
        raw: JSON.stringify({ id: "click-1", type: "UI_INTERACTION" }),
      });

      result = await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Alternative button",
          },
        },
      });

      expect(result.content[0].text).toContain("Successfully executed");
      // Now position should increment
      expect(mockCliWrapper.updateContext).toHaveBeenCalledWith({
        position: 3,
      });
    });

    it("should maintain session consistency across errors", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");

      // Track context changes
      let currentContext = { checkpointId: "111", position: 1 };
      mockCliWrapper.getContext.mockImplementation(() => currentContext);
      mockCliWrapper.updateContext.mockImplementation((update) => {
        currentContext = { ...currentContext, ...update };
      });

      // Set initial context
      await callToolHandler({
        params: {
          name: "virtuoso_set_context",
          arguments: {
            checkpointId: "111",
            position: 5,
          },
        },
      });

      expect(currentContext).toEqual({ checkpointId: "111", position: 5 });

      // Execute successful command
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-1" },
        raw: JSON.stringify({ id: "step-1" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Button",
          },
        },
      });

      expect(currentContext.position).toBe(6);

      // Execute failing command
      mockCliWrapper.execute.mockResolvedValue({
        success: false,
        error: "Failed",
        raw: "",
      });

      await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Bad button",
          },
        },
      });

      // Position should still be 6
      expect(currentContext.position).toBe(6);

      // Execute another successful command
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-2" },
        raw: JSON.stringify({ id: "step-2" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Good button",
          },
        },
      });

      // Position should now be 7
      expect(currentContext.position).toBe(7);
    });
  });

  describe("Resource Integration", () => {
    it("should provide accurate session context through resources", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");
      const readResourceHandler = handlers.get("resources/read");

      // Set context
      await callToolHandler({
        params: {
          name: "virtuoso_set_context",
          arguments: {
            checkpointId: "777",
            position: 10,
            journeyId: "888",
            goalId: "999",
          },
        },
      });

      // Mock getContext to return the set values
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "777",
        position: 10,
        journeyId: "888",
        goalId: "999",
      });

      // Read context resource
      const result = await readResourceHandler({
        params: {
          uri: "virtuoso://session/context",
        },
      });

      const contextData = JSON.parse(result.contents[0].text);
      expect(contextData).toEqual({
        checkpointId: "777",
        position: 10,
        journeyId: "888",
        goalId: "999",
      });

      // Execute a command
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "step-1" },
        raw: JSON.stringify({ id: "step-1" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Button",
          },
        },
      });

      // Update mock to reflect incremented position
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "777",
        position: 11,
        journeyId: "888",
        goalId: "999",
      });

      // Read context again
      const updatedResult = await readResourceHandler({
        params: {
          uri: "virtuoso://session/context",
        },
      });

      const updatedContextData = JSON.parse(updatedResult.contents[0].text);
      expect(updatedContextData.position).toBe(11);
    });
  });

  describe("Complex Workflows", () => {
    it("should handle file upload workflow", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");

      // Setup
      await callToolHandler({
        params: {
          name: "virtuoso_set_context",
          arguments: {
            checkpointId: "222",
            position: 1,
          },
        },
      });

      // Navigate to upload page
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "nav-1" },
        raw: JSON.stringify({ id: "nav-1" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_navigate",
          arguments: {
            action: "to",
            url: "https://example.com/upload",
          },
        },
      });

      // Click upload button to open file dialog
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "222",
        position: 2,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "click-1" },
        raw: JSON.stringify({ id: "click-1" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Choose file button",
          },
        },
      });

      // Upload file
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "222",
        position: 3,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "file-1", type: "FILE_UPLOAD" },
        raw: JSON.stringify({ id: "file-1", type: "FILE_UPLOAD" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_file",
          arguments: {
            action: "upload",
            url: "https://example.com/files/test.pdf",
            selector: "#file-input",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "file",
        ["upload", "https://example.com/files/test.pdf", "#file-input"],
        expect.objectContaining({
          checkpoint: "222",
          position: 3,
        }),
      );

      // Wait for upload to complete
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "222",
        position: 4,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "wait-1" },
        raw: JSON.stringify({ id: "wait-1" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_wait",
          arguments: {
            action: "element",
            selector: "Upload complete message",
            timeout: 10000,
          },
        },
      });

      // Assert upload successful
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "222",
        position: 5,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "assert-1" },
        raw: JSON.stringify({ id: "assert-1" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_assert",
          arguments: {
            action: "exists",
            selector: "File uploaded successfully",
          },
        },
      });
    });

    it("should handle dialog interaction workflow", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");

      // Setup
      await callToolHandler({
        params: {
          name: "virtuoso_set_context",
          arguments: {
            checkpointId: "333",
            position: 1,
          },
        },
      });

      // Click button that triggers confirm dialog
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "click-1" },
        raw: JSON.stringify({ id: "click-1" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_interact",
          arguments: {
            action: "click",
            selector: "Delete button",
          },
        },
      });

      // Handle confirm dialog
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "333",
        position: 2,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "dialog-1", type: "DIALOG_INTERACTION" },
        raw: JSON.stringify({ id: "dialog-1", type: "DIALOG_INTERACTION" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_dialog",
          arguments: {
            action: "dismiss-confirm",
            accept: true,
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "dialog",
        ["dismiss-confirm"],
        expect.objectContaining({
          checkpoint: "333",
          position: 2,
          accept: true,
        }),
      );

      // Wait for deletion to complete
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "333",
        position: 3,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "wait-1" },
        raw: JSON.stringify({ id: "wait-1" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_wait",
          arguments: {
            action: "time",
            milliseconds: 1000,
          },
        },
      });

      // Assert item was deleted
      mockCliWrapper.getContext.mockReturnValue({
        checkpointId: "333",
        position: 4,
      });
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: { id: "assert-1" },
        raw: JSON.stringify({ id: "assert-1" }),
      });

      await callToolHandler({
        params: {
          name: "virtuoso_assert",
          arguments: {
            action: "not-exists",
            selector: "Deleted item",
          },
        },
      });
    });
  });

  describe("Library Integration", () => {
    it("should handle library checkpoint workflow", async () => {
      const serverInstance = (server as any).server as Server;
      const handlers = (serverInstance as any)._requestHandlers;
      const callToolHandler = handlers.get("tools/call");

      // Create a library checkpoint from existing checkpoint
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: {
          id: "7023",
          title: "Login Flow",
          type: "LIBRARY_CHECKPOINT",
        },
        raw: JSON.stringify({
          id: "7023",
          title: "Login Flow",
          type: "LIBRARY_CHECKPOINT",
        }),
      });

      let result = await callToolHandler({
        params: {
          name: "virtuoso_library",
          arguments: {
            action: "add",
            checkpointId: "1680930",
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "library",
        ["add", "1680930"],
        expect.any(Object),
      );
      expect(result.content[0].text).toContain("Successfully executed");

      // Get library checkpoint details
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: {
          id: "7023",
          title: "Login Flow",
          steps: [
            { id: "19660498", description: "Navigate to login" },
            { id: "19660499", description: "Enter username" },
            { id: "19660500", description: "Enter password" },
            { id: "19660501", description: "Click login" },
          ],
        },
        raw: JSON.stringify({
          id: "7023",
          title: "Login Flow",
          steps: [
            { id: "19660498", description: "Navigate to login" },
            { id: "19660499", description: "Enter username" },
            { id: "19660500", description: "Enter password" },
            { id: "19660501", description: "Click login" },
          ],
        }),
      });

      result = await callToolHandler({
        params: {
          name: "virtuoso_library",
          arguments: {
            action: "get",
            libraryCheckpointId: "7023",
          },
        },
      });

      // Attach library checkpoint to a journey
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: {
          id: "attach-1",
          message: "Library checkpoint attached successfully",
        },
        raw: JSON.stringify({
          id: "attach-1",
          message: "Library checkpoint attached successfully",
        }),
      });

      result = await callToolHandler({
        params: {
          name: "virtuoso_library",
          arguments: {
            action: "attach",
            journeyId: "608926",
            libraryCheckpointId: "7023",
            position: 4,
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "library",
        ["attach", "608926", "7023", "4"],
        expect.any(Object),
      );

      // Move a step within the library checkpoint
      mockCliWrapper.execute.mockResolvedValue({
        success: true,
        data: {
          id: "move-1",
          message: "Step moved successfully",
        },
        raw: JSON.stringify({
          id: "move-1",
          message: "Step moved successfully",
        }),
      });

      result = await callToolHandler({
        params: {
          name: "virtuoso_library",
          arguments: {
            action: "move-step",
            libraryCheckpointId: "7023",
            stepId: "19660500",
            newPosition: 2,
          },
        },
      });

      expect(mockCliWrapper.execute).toHaveBeenCalledWith(
        "library",
        ["move-step", "7023", "19660500", "2"],
        expect.any(Object),
      );
    });
  });
});
