import { z } from "zod";
import { VirtuosoCliWrapper } from "../cli-wrapper.js";
import { formatToolResponse } from "../utils/formatting.js";
import { sanitizeInput } from "../utils/validation.js";
import { registerTool } from "./all-tools.js";

const windowToolSchema = z.object({
  action: z.enum(["resize", "switch-tab", "switch-frame"] as const),
  // For resize
  width: z.number().optional().describe("Window width in pixels"),
  height: z.number().optional().describe("Window height in pixels"),
  // For switch-tab/switch-frame
  target: z
    .string()
    .optional()
    .describe('Tab index, "next"/"previous", or frame selector'),
  checkpoint: z.string().optional(),
  position: z.number().optional(),
});

type WindowToolInput = z.infer<typeof windowToolSchema>;

export function registerWindowTools(cli: VirtuosoCliWrapper) {
  registerTool(
    {
      name: "virtuoso_window",
      description:
        "Manage browser windows, tabs, and frames in Virtuoso tests.",
      inputSchema: {
        type: "object",
        properties: {
          action: {
            type: "string",
            enum: ["resize", "switch-tab", "switch-frame"],
            description: "Window management action to perform",
          },
          width: {
            type: "number",
            description: "Window width in pixels (required for resize)",
          },
          height: {
            type: "number",
            description: "Window height in pixels (required for resize)",
          },
          target: {
            type: "string",
            description:
              'Tab index, "next"/"previous", or frame selector (required for switch actions)',
          },
          checkpoint: {
            type: "string",
            description: "Checkpoint ID (uses current context if not provided)",
          },
          position: {
            type: "number",
            description:
              "Position in test sequence (auto-increments if not provided)",
          },
        },
        required: ["action"],
        anyOf: [
          {
            properties: { action: { const: "resize" } },
            required: ["action", "width", "height"],
          },
          {
            properties: { action: { const: "switch-tab" } },
            required: ["action", "target"],
          },
          {
            properties: { action: { const: "switch-frame" } },
            required: ["action", "target"],
          },
        ],
      },
    },
    async (args) => {
      try {
        const input = windowToolSchema.parse(args);

        // Get context if checkpoint not provided
        const context = cli.getContext();
        const checkpoint = input.checkpoint || context.checkpointId;

        if (!checkpoint) {
          throw new Error(
            "Checkpoint ID required. Use virtuoso_set_context first or provide checkpoint parameter.",
          );
        }

        // Build command arguments based on action
        let cliArgs: string[] = [];
        let description = "";

        switch (input.action) {
          case "resize":
            if (!input.width || !input.height) {
              throw new Error(
                "Width and height are required for resize action",
              );
            }
            cliArgs = [
              "window",
              "resize",
              checkpoint,
              String(input.width),
              String(input.height),
              String(input.position || context.position),
            ];
            description = `Resize window to ${input.width}x${input.height}`;
            break;

          case "switch-tab":
            if (!input.target) {
              throw new Error("Target is required for switch-tab action");
            }
            cliArgs = [
              "window",
              "switch-tab",
              checkpoint,
              sanitizeInput(input.target),
              String(input.position || context.position),
            ];

            if (input.target === "next" || input.target === "previous") {
              description = `Switch to ${input.target} tab`;
            } else {
              description = `Switch to tab ${input.target}`;
            }
            break;

          case "switch-frame":
            if (!input.target) {
              throw new Error("Target is required for switch-frame action");
            }
            cliArgs = [
              "window",
              "switch-frame",
              checkpoint,
              sanitizeInput(input.target),
              String(input.position || context.position),
            ];

            if (input.target === "parent" || input.target === "top") {
              description = `Switch to ${input.target} frame`;
            } else {
              description = `Switch to frame: ${input.target}`;
            }
            break;
        }

        // Execute command
        const result = await cli.execute(cliArgs);

        // Auto-increment position if using context
        if (!input.position && context.checkpointId) {
          cli.updateContext({ position: context.position + 1 });
        }

        const response = formatToolResponse(result, {
          checkpoint,
          position: input.position || context.position,
        });

        return {
          content: [
            {
              type: "text",
              text: `✅ Created ${
                input.action
              } step: ${description}\n\nStep ID: ${
                response.data?.stepId || "N/A"
              }\nCheckpoint: ${checkpoint}\nPosition: ${
                input.position || context.position
              }`,
            },
          ],
        };
      } catch (error) {
        return {
          content: [
            {
              type: "text",
              text: `❌ Error: ${
                error instanceof Error
                  ? error.message
                  : "Unknown error occurred"
              }`,
            },
          ],
        };
      }
    },
  );
}
