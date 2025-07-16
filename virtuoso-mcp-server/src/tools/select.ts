import { z } from "zod";
import { VirtuosoCliWrapper } from "../cli-wrapper.js";
import { formatToolResponse } from "../utils/formatting.js";
import { sanitizeInput } from "../utils/validation.js";
import { registerTool } from "./all-tools.js";

const selectToolSchema = z.object({
  action: z.enum(["option", "index", "last"] as const),
  selector: z.string().describe("Dropdown/select element selector"),
  optionText: z.string().optional().describe("Option text or value to select"),
  index: z.number().optional().describe("Option index to select"),
  checkpoint: z.string().optional(),
  position: z.number().optional(),
  // Optional parameters
  byValue: z
    .boolean()
    .optional()
    .describe("Select by value instead of text (for option action)"),
});

type SelectToolInput = z.infer<typeof selectToolSchema>;

export function registerSelectTools(cli: VirtuosoCliWrapper) {
  registerTool(
    {
      name: "virtuoso_select",
      description:
        "Select dropdown options in Virtuoso tests by text, value, index, or select the last option.",
      inputSchema: {
        type: "object",
        properties: {
          action: {
            type: "string",
            enum: ["option", "index", "last"],
            description: "Type of select operation",
          },
          selector: {
            type: "string",
            description: "Dropdown/select element selector (required)",
          },
          optionText: {
            type: "string",
            description:
              "Option text or value to select (required for option action)",
          },
          index: {
            type: "number",
            description:
              "Option index to select (required for index action, 0-based)",
          },
          checkpoint: {
            type: "string",
            description: "Checkpoint ID (optional)",
          },
          position: { type: "number", description: "Step position (optional)" },
          byValue: {
            type: "boolean",
            description: "Select by value instead of text (for option action)",
          },
        },
        required: ["action", "selector"],
        additionalProperties: false,
      },
    },
    async (args) => {
      try {
        const input = selectToolSchema.parse(args);

        // Validate required fields based on action
        validateSelectInput(input);

        const cliArgs = buildSelectCommand(input);

        // Build options object
        const options: any = {
          checkpoint: input.checkpoint,
          position: input.position,
        };

        // Add optional flags
        if (input.action === "option" && input.byValue) {
          options["by-value"] = true;
        }

        const result = await cli.execute(cliArgs, options);
        const response = formatToolResponse(result, cli.getContext());

        if (!response.success) {
          return {
            content: [
              {
                type: "text",
                text: `❌ Select operation failed: ${response.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: formatSelectSuccessMessage(input, response.data),
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

function validateSelectInput(input: SelectToolInput) {
  if (!input.selector) {
    throw new Error("Selector is required for all select operations");
  }

  switch (input.action) {
    case "option":
      if (!input.optionText) {
        throw new Error("Option text is required for option action");
      }
      break;

    case "index":
      if (input.index === undefined || input.index === null) {
        throw new Error("Index is required for index action");
      }
      if (input.index < 0) {
        throw new Error("Index must be a non-negative number");
      }
      break;

    case "last":
      // No additional validation needed for last action
      break;
  }
}

function buildSelectCommand(input: SelectToolInput): string[] {
  const args = ["select", input.action];

  // Selector is always required
  args.push(sanitizeInput(input.selector));

  switch (input.action) {
    case "option":
      args.push(sanitizeInput(input.optionText!));
      break;

    case "index":
      args.push(input.index!.toString());
      break;

    case "last":
      // No additional arguments needed
      break;
  }

  return args;
}

function formatSelectSuccessMessage(input: SelectToolInput, data: any): string {
  const stepInfo = data.stepId ? ` (Step ${data.stepId})` : "";

  switch (input.action) {
    case "option":
      const selectBy = input.byValue ? "value" : "text";
      return `✅ Created select option by ${selectBy}: "${input.optionText}" from "${input.selector}"${stepInfo}`;

    case "index":
      return `✅ Created select option by index ${input.index} from "${input.selector}"${stepInfo}`;

    case "last":
      return `✅ Created select last option from "${input.selector}"${stepInfo}`;

    default:
      return `✅ Created ${input.action} select operation${stepInfo}`;
  }
}
