import { z } from "zod";
import { VirtuosoCliWrapper } from "../cli-wrapper.js";
import { formatToolResponse } from "../utils/formatting.js";
import { sanitizeInput } from "../utils/validation.js";
import { registerTool } from "./all-tools.js";

const mouseToolSchema = z.object({
  action: z.enum([
    "move-to",
    "move-by",
    "move",
    "down",
    "up",
    "enter",
  ] as const),
  // For move-to
  x: z
    .union([z.string(), z.number()])
    .optional()
    .describe("X coordinate for move-to"),
  y: z
    .union([z.string(), z.number()])
    .optional()
    .describe("Y coordinate for move-to"),
  // For move-by
  dx: z
    .union([z.string(), z.number()])
    .optional()
    .describe("Horizontal offset for move-by"),
  dy: z
    .union([z.string(), z.number()])
    .optional()
    .describe("Vertical offset for move-by"),
  // For move
  selector: z.string().optional().describe("Element selector for move action"),
  // For down/up
  button: z
    .enum(["left", "middle", "right"])
    .optional()
    .describe("Mouse button (default: left)"),
  // Common parameters
  checkpoint: z.string().optional(),
  position: z.number().optional(),
});

type MouseToolInput = z.infer<typeof mouseToolSchema>;

export function registerMouseTools(cli: VirtuosoCliWrapper) {
  registerTool(
    {
      name: "virtuoso_mouse",
      description:
        "Perform mouse operations including movement, clicks, and viewport entry.",
      inputSchema: {
        type: "object",
        properties: {
          action: {
            type: "string",
            enum: ["move-to", "move-by", "move", "down", "up", "enter"],
            description: "Type of mouse action to perform",
          },
          x: {
            type: ["string", "number"],
            description: "X coordinate in pixels (required for move-to)",
          },
          y: {
            type: ["string", "number"],
            description: "Y coordinate in pixels (required for move-to)",
          },
          dx: {
            type: ["string", "number"],
            description: "Horizontal offset in pixels (required for move-by)",
          },
          dy: {
            type: ["string", "number"],
            description: "Vertical offset in pixels (required for move-by)",
          },
          selector: {
            type: "string",
            description: "Element selector (required for move action)",
          },
          button: {
            type: "string",
            enum: ["left", "middle", "right"],
            description: "Mouse button for down/up actions (default: left)",
          },
          checkpoint: {
            type: "string",
            description: "Checkpoint ID (optional)",
          },
          position: { type: "number", description: "Step position (optional)" },
        },
        required: ["action"],
        additionalProperties: false,
      },
    },
    async (args) => {
      try {
        const input = mouseToolSchema.parse(args);
        const cliArgs = buildMouseCommand(input);

        // Build options object
        const options: any = {
          checkpoint: input.checkpoint,
          position: input.position,
        };

        // Add button flag for down/up actions
        if (
          (input.action === "down" || input.action === "up") &&
          input.button &&
          input.button !== "left"
        ) {
          options.button = input.button;
        }

        const result = await cli.execute(cliArgs, options);
        const response = formatToolResponse(result, cli.getContext());

        if (!response.success) {
          return {
            content: [
              {
                type: "text",
                text: `❌ Mouse action failed: ${response.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: formatMouseSuccessMessage(input, response.data),
            },
          ],
        };
      } catch (error) {
        return {
          content: [
            {
              type: "text",
              text: `❌ Error: ${
                error instanceof Error ? error.message : String(error)
              }`,
            },
          ],
          isError: true,
        };
      }
    },
  );
}

function buildMouseCommand(input: MouseToolInput): string[] {
  const args = ["mouse", input.action];

  switch (input.action) {
    case "move-to":
      if (input.x === undefined || input.y === undefined) {
        throw new Error("X and Y coordinates required for move-to action");
      }
      args.push(String(input.x), String(input.y));
      break;

    case "move-by":
      if (input.dx === undefined || input.dy === undefined) {
        throw new Error("DX and DY offsets required for move-by action");
      }
      args.push(String(input.dx), String(input.dy));
      break;

    case "move":
      if (!input.selector) {
        throw new Error("Selector required for move action");
      }
      args.push(sanitizeInput(input.selector));
      break;

    case "down":
    case "up":
    case "enter":
      // No additional arguments required
      break;
  }

  return args;
}

function formatMouseSuccessMessage(input: MouseToolInput, data: any): string {
  const stepInfo = data.stepId ? ` (Step ${data.stepId})` : "";

  switch (input.action) {
    case "move-to":
      return `✅ Created mouse move-to action: coordinates (${input.x}, ${input.y})${stepInfo}`;
    case "move-by":
      return `✅ Created mouse move-by action: offset (${input.dx}, ${input.dy})${stepInfo}`;
    case "move":
      return `✅ Created mouse move action to element: "${input.selector}"${stepInfo}`;
    case "down":
      return `✅ Created mouse ${
        input.button || "left"
      } button down action${stepInfo}`;
    case "up":
      return `✅ Created mouse ${
        input.button || "left"
      } button up action${stepInfo}`;
    case "enter":
      return `✅ Created mouse enter viewport action${stepInfo}`;
    default:
      return `✅ Created mouse ${input.action} action${stepInfo}`;
  }
}
