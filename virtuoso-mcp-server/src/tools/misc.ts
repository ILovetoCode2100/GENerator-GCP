import { z } from "zod";
import { VirtuosoCliWrapper } from "../cli-wrapper.js";
import { formatToolResponse } from "../utils/formatting.js";
import { sanitizeInput } from "../utils/validation.js";
import { registerTool } from "./all-tools.js";

const miscToolSchema = z.object({
  action: z.enum(["comment", "execute-script", "key"] as const),
  text: z.string().optional().describe("Comment text (for comment action)"),
  script: z
    .string()
    .optional()
    .describe("JavaScript code to execute (for execute-script action)"),
  key: z.string().optional().describe("Key to press (for key action)"),
  checkpoint: z
    .string()
    .optional()
    .describe("Checkpoint ID (optional, uses session context if not provided)"),
  position: z
    .number()
    .optional()
    .describe("Step position (optional, auto-increments if not provided)"),
  // Optional parameters
  target: z.string().optional().describe("Target selector for key action"),
});

type MiscToolInput = z.infer<typeof miscToolSchema>;

export function registerMiscTools(cli: VirtuosoCliWrapper) {
  registerTool(
    {
      name: "virtuoso_misc",
      description:
        "Miscellaneous Virtuoso test actions including comments, script execution, and keyboard shortcuts.",
      inputSchema: {
        type: "object",
        properties: {
          action: {
            type: "string",
            enum: ["comment", "execute-script", "key"],
            description: "Type of miscellaneous action to perform",
          },
          text: {
            type: "string",
            description: "Comment text (required for comment action)",
          },
          script: {
            type: "string",
            description:
              "JavaScript code to execute (required for execute-script action)",
          },
          key: {
            type: "string",
            description:
              "Key or key combination to press (required for key action, e.g., CTRL+S, ALT+TAB)",
          },
          checkpoint: {
            type: "string",
            description:
              "Checkpoint ID (optional, uses session context if not provided)",
          },
          position: {
            type: "number",
            description:
              "Step position (optional, auto-increments if not provided)",
          },
          target: {
            type: "string",
            description: "Target selector for key action (optional)",
          },
        },
        required: ["action"],
        additionalProperties: false,
      },
    },
    async (args) => {
      try {
        // Parse and validate input
        const input = miscToolSchema.parse(args);

        // Build CLI command based on action type
        const cliArgs = buildMiscCommand(input);

        // Build options object
        const options: any = {
          checkpoint: input.checkpoint,
          position: input.position,
        };

        // Add optional parameters for key action
        if (input.action === "key" && input.target) {
          options.target = input.target;
        }

        // Execute CLI command
        const result = await cli.execute(cliArgs, options);

        // Format response
        const response = formatToolResponse(result, cli.getContext());

        if (!response.success) {
          return {
            content: [
              {
                type: "text",
                text: `❌ Misc action failed: ${response.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: formatMiscSuccessMessage(input, response.data),
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

/**
 * Build CLI command arguments for misc actions
 */
function buildMiscCommand(input: MiscToolInput): string[] {
  const args = ["misc", input.action];

  switch (input.action) {
    case "comment":
      if (!input.text) {
        throw new Error("Text required for comment action");
      }
      args.push(sanitizeInput(input.text));
      break;

    case "execute-script":
      if (!input.script) {
        throw new Error("Script required for execute-script action");
      }
      // Don't sanitize JavaScript code as it may contain quotes and special characters
      args.push(input.script);
      break;

    case "key":
      if (!input.key) {
        throw new Error("Key required for key action");
      }
      // Convert key to uppercase and handle special formatting
      args.push(input.key.toUpperCase());
      break;

    default:
      throw new Error(`Unknown misc action: ${input.action}`);
  }

  return args;
}

/**
 * Format success message based on misc action type
 */
function formatMiscSuccessMessage(input: MiscToolInput, data: any): string {
  const stepInfo = data.stepId ? ` (Step ${data.stepId})` : "";

  switch (input.action) {
    case "comment":
      return `✅ Added comment: "${input.text}"${stepInfo}`;

    case "execute-script":
      const scriptPreview =
        input.script!.length > 50
          ? input.script!.substring(0, 50) + "..."
          : input.script!;
      return `✅ Created script execution: "${scriptPreview}"${stepInfo}`;

    case "key":
      const targetInfo = input.target ? ` on "${input.target}"` : "";
      return `✅ Created key press: ${input.key}${targetInfo}${stepInfo}`;

    default:
      return `✅ Created ${input.action} action${stepInfo}`;
  }
}
