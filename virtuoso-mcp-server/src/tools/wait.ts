import { z } from "zod";
import { VirtuosoCliWrapper } from "../cli-wrapper.js";
import { formatToolResponse } from "../utils/formatting.js";
import { sanitizeInput } from "../utils/validation.js";
import { registerTool } from "./all-tools.js";

const waitToolSchema = z.object({
  type: z.enum(["element", "time"] as const),
  selector: z.string().optional().describe("Element selector to wait for"),
  milliseconds: z.number().optional().describe("Time to wait in milliseconds"),
  checkpoint: z.string().optional(),
  position: z.number().optional(),
  // Optional parameters for element wait
  timeout: z
    .number()
    .optional()
    .describe("Timeout in milliseconds for element wait"),
  visible: z.boolean().optional().describe("Wait for element to be visible"),
  hidden: z.boolean().optional().describe("Wait for element to be hidden"),
});

type WaitToolInput = z.infer<typeof waitToolSchema>;

export function registerWaitTools(cli: VirtuosoCliWrapper) {
  registerTool(
    {
      name: "virtuoso_wait",
      description:
        "Add wait conditions in Virtuoso tests. Wait for elements to appear/disappear or wait for a specific duration.",
      inputSchema: {
        type: "object",
        properties: {
          type: {
            type: "string",
            enum: ["element", "time"],
            description: "Type of wait operation",
          },
          selector: {
            type: "string",
            description: "Element selector (required for element wait)",
          },
          milliseconds: {
            type: "number",
            description:
              "Time to wait in milliseconds (required for time wait)",
          },
          checkpoint: {
            type: "string",
            description: "Checkpoint ID (optional)",
          },
          position: { type: "number", description: "Step position (optional)" },
          timeout: {
            type: "number",
            description: "Timeout in milliseconds for element wait (optional)",
          },
          visible: {
            type: "boolean",
            description: "Wait for element to be visible (optional)",
          },
          hidden: {
            type: "boolean",
            description: "Wait for element to be hidden (optional)",
          },
        },
        required: ["type"],
        additionalProperties: false,
      },
    },
    async (args) => {
      try {
        const input = waitToolSchema.parse(args);

        // Validate required fields based on type
        validateWaitInput(input);

        const cliArgs = buildWaitCommand(input);

        // Build options object
        const options: any = {
          checkpoint: input.checkpoint,
          position: input.position,
        };

        // Add optional flags for element wait
        if (input.type === "element") {
          if (input.timeout !== undefined) options.timeout = input.timeout;
          if (input.visible) options.visible = true;
          if (input.hidden) options.hidden = true;
        }

        const result = await cli.execute(cliArgs, options);
        const response = formatToolResponse(result, cli.getContext());

        if (!response.success) {
          return {
            content: [
              {
                type: "text",
                text: `❌ Wait operation failed: ${response.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: formatWaitSuccessMessage(input, response.data),
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

function validateWaitInput(input: WaitToolInput) {
  switch (input.type) {
    case "element":
      if (!input.selector) {
        throw new Error("Selector is required for element wait");
      }
      if (input.visible && input.hidden) {
        throw new Error(
          "Cannot wait for element to be both visible and hidden",
        );
      }
      break;

    case "time":
      if (input.milliseconds === undefined) {
        throw new Error("Milliseconds is required for time wait");
      }
      if (input.milliseconds <= 0) {
        throw new Error("Milliseconds must be a positive number");
      }
      break;
  }
}

function buildWaitCommand(input: WaitToolInput): string[] {
  const args = ["wait", input.type];

  switch (input.type) {
    case "element":
      args.push(sanitizeInput(input.selector!));
      break;

    case "time":
      args.push(String(input.milliseconds));
      break;
  }

  return args;
}

function formatWaitSuccessMessage(input: WaitToolInput, data: any): string {
  const stepInfo = data.stepId ? ` (Step ${data.stepId})` : "";

  switch (input.type) {
    case "element":
      let waitCondition = "appear";
      if (input.visible) waitCondition = "be visible";
      if (input.hidden) waitCondition = "be hidden";

      const timeoutInfo = input.timeout ? ` (timeout: ${input.timeout}ms)` : "";
      return `✅ Created wait for element "${input.selector}" to ${waitCondition}${timeoutInfo}${stepInfo}`;

    case "time":
      return `✅ Created wait for ${input.milliseconds}ms${stepInfo}`;

    default:
      return `✅ Created ${input.type} wait${stepInfo}`;
  }
}
