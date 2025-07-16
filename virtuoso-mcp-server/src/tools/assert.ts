import { z } from "zod";
import { VirtuosoCliWrapper } from "../cli-wrapper.js";
import { formatToolResponse } from "../utils/formatting.js";
import { sanitizeInput } from "../utils/validation.js";
import { registerTool } from "./all-tools.js";

const assertToolSchema = z.object({
  type: z.enum([
    "exists",
    "not-exists",
    "equals",
    "not-equals",
    "checked",
    "selected",
    "gt",
    "gte",
    "lt",
    "lte",
    "matches",
    "variable",
  ] as const),
  element: z.string().optional().describe("Element selector or text"),
  value: z.string().optional().describe("Value for comparison assertions"),
  variable: z
    .string()
    .optional()
    .describe("Variable name for variable assertions"),
  checkpoint: z
    .string()
    .optional()
    .describe("Checkpoint ID (optional, uses session context if not provided)"),
  position: z
    .number()
    .optional()
    .describe("Step position (optional, auto-increments if not provided)"),
});

type AssertToolInput = z.infer<typeof assertToolSchema>;

export function registerAssertTools(cli: VirtuosoCliWrapper) {
  registerTool(
    {
      name: "virtuoso_assert",
      description:
        "Create assertion steps in Virtuoso tests. Supports various assertion types including element existence, value comparisons, and pattern matching.",
      inputSchema: {
        type: "object",
        properties: {
          type: {
            type: "string",
            enum: [
              "exists",
              "not-exists",
              "equals",
              "not-equals",
              "checked",
              "selected",
              "gt",
              "gte",
              "lt",
              "lte",
              "matches",
              "variable",
            ],
            description: "Type of assertion to perform",
          },
          element: {
            type: "string",
            description:
              "Element selector or text (required for most assertions)",
          },
          value: {
            type: "string",
            description:
              "Expected value (required for comparison and matches assertions)",
          },
          variable: {
            type: "string",
            description: "Variable name (required for variable assertions)",
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
        },
        required: ["type"],
        additionalProperties: false,
      },
    },
    async (args) => {
      try {
        // Parse and validate input
        const input = assertToolSchema.parse(args);

        // Build CLI command based on assertion type
        const cliArgs = buildAssertCommand(input);

        // Execute CLI command
        const result = await cli.execute(cliArgs, {
          checkpoint: input.checkpoint,
          position: input.position,
        });

        // Format response
        const response = formatToolResponse(result, cli.getContext());

        if (!response.success) {
          return {
            content: [
              {
                type: "text",
                text: `❌ Assertion failed: ${response.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: formatAssertSuccessMessage(input, response.data),
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
 * Build CLI command arguments for assert
 */
function buildAssertCommand(input: AssertToolInput): string[] {
  const args = ["assert", input.type];

  // Add arguments based on assertion type
  switch (input.type) {
    case "exists":
    case "not-exists":
    case "checked":
    case "selected":
      if (!input.element) {
        throw new Error(
          `Element selector required for ${input.type} assertion`,
        );
      }
      args.push(sanitizeInput(input.element));
      break;

    case "equals":
    case "not-equals":
      if (!input.element || !input.value) {
        throw new Error(
          `Element and value required for ${input.type} assertion`,
        );
      }
      args.push(sanitizeInput(input.element));
      args.push(sanitizeInput(input.value));
      break;

    case "gt":
    case "gte":
    case "lt":
    case "lte":
      if (!input.element || !input.value) {
        throw new Error(
          `Element and numeric value required for ${input.type} assertion`,
        );
      }
      // Validate numeric value
      if (isNaN(Number(input.value))) {
        throw new Error(`Value must be numeric for ${input.type} assertion`);
      }
      args.push(sanitizeInput(input.element));
      args.push(input.value);
      break;

    case "matches":
      if (!input.element || !input.value) {
        throw new Error(
          `Element and regex pattern required for matches assertion`,
        );
      }
      args.push(sanitizeInput(input.element));
      args.push(input.value); // Don't sanitize regex patterns
      break;

    case "variable":
      if (!input.variable || !input.value) {
        throw new Error(
          `Variable name and expected value required for variable assertion`,
        );
      }
      args.push(input.variable);
      args.push(sanitizeInput(input.value));
      break;
  }

  return args;
}

/**
 * Format success message based on assertion type
 */
function formatAssertSuccessMessage(input: AssertToolInput, data: any): string {
  const stepInfo = data.stepId ? ` (Step ${data.stepId})` : "";

  switch (input.type) {
    case "exists":
      return `✅ Created assertion: Element "${input.element}" exists${stepInfo}`;

    case "not-exists":
      return `✅ Created assertion: Element "${input.element}" does not exist${stepInfo}`;

    case "equals":
      return `✅ Created assertion: "${input.element}" equals "${input.value}"${stepInfo}`;

    case "not-equals":
      return `✅ Created assertion: "${input.element}" not equals "${input.value}"${stepInfo}`;

    case "checked":
      return `✅ Created assertion: "${input.element}" is checked${stepInfo}`;

    case "selected":
      return `✅ Created assertion: "${input.element}" is selected${stepInfo}`;

    case "gt":
      return `✅ Created assertion: "${input.element}" > ${input.value}${stepInfo}`;

    case "gte":
      return `✅ Created assertion: "${input.element}" >= ${input.value}${stepInfo}`;

    case "lt":
      return `✅ Created assertion: "${input.element}" < ${input.value}${stepInfo}`;

    case "lte":
      return `✅ Created assertion: "${input.element}" <= ${input.value}${stepInfo}`;

    case "matches":
      return `✅ Created assertion: "${input.element}" matches pattern "${input.value}"${stepInfo}`;

    case "variable":
      return `✅ Created assertion: Variable "${input.variable}" equals "${input.value}"${stepInfo}`;

    default:
      return `✅ Created ${input.type} assertion${stepInfo}`;
  }
}
