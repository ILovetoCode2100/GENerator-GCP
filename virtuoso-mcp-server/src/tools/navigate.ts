import { z } from "zod";
import { VirtuosoCliWrapper } from "../cli-wrapper.js";
import {
  urlSchema,
  selectorSchema,
  scrollBlockSchema,
  scrollInlineSchema,
} from "../utils/validation.js";
import { formatToolResponse } from "../utils/formatting.js";
import { sanitizeInput } from "../utils/validation.js";
import { registerTool } from "./all-tools.js";

const navigateToolSchema = z.object({
  action: z.enum([
    "to",
    "scroll-top",
    "scroll-bottom",
    "scroll-element",
    "scroll-position",
  ] as const),
  url: z.string().optional().describe("URL to navigate to"),
  selector: z
    .string()
    .optional()
    .describe("Element selector for scroll-element"),
  direction: z.enum(["up", "down"]).optional().describe("Scroll direction"),
  x: z.number().optional().describe("X coordinate for scroll-position"),
  y: z.number().optional().describe("Y coordinate for scroll-position"),
  checkpoint: z.string().optional(),
  position: z.number().optional(),
  // Optional parameters
  newTab: z.boolean().optional().describe("Open URL in new tab"),
  wait: z.boolean().optional().describe("Wait for page to load"),
  smooth: z.boolean().optional().describe("Use smooth scrolling"),
  block: scrollBlockSchema.optional().describe("Vertical alignment"),
  inline: scrollInlineSchema.optional().describe("Horizontal alignment"),
});

type NavigateToolInput = z.infer<typeof navigateToolSchema>;

export function registerNavigateTools(cli: VirtuosoCliWrapper) {
  const toolDefinition = {
    name: "virtuoso_navigate",
    description:
      "Navigate to URLs and control page scrolling in Virtuoso tests.",
    inputSchema: {
      type: "object" as const,
      properties: {
        action: {
          type: "string" as const,
          enum: [
            "to",
            "scroll-top",
            "scroll-bottom",
            "scroll-element",
            "scroll-position",
          ],
          description: "Type of navigation action",
        },
        url: {
          type: "string" as const,
          description: 'URL to navigate to (required for "to" action)',
        },
        selector: {
          type: "string" as const,
          description: "Element selector (required for scroll-element)",
        },
        direction: {
          type: "string" as const,
          enum: ["up", "down"],
          description: "Scroll direction for scroll-element",
        },
        x: {
          type: "number" as const,
          description: "X coordinate for scroll-position",
        },
        y: {
          type: "number" as const,
          description: "Y coordinate for scroll-position",
        },
        checkpoint: {
          type: "string" as const,
          description: "Checkpoint ID (optional)",
        },
        position: {
          type: "number" as const,
          description: "Step position (optional)",
        },
        newTab: {
          type: "boolean" as const,
          description: "Open URL in new tab",
        },
        wait: {
          type: "boolean" as const,
          description: "Wait for page to load",
        },
        smooth: {
          type: "boolean" as const,
          description: "Use smooth scrolling",
        },
        block: {
          type: "string" as const,
          enum: ["start", "center", "end", "nearest"],
          description: "Vertical alignment for scroll-element",
        },
        inline: {
          type: "string" as const,
          enum: ["start", "center", "end", "nearest"],
          description: "Horizontal alignment for scroll-element",
        },
      },
      required: ["action"],
      additionalProperties: false,
    },
  };

  const handler = async (args: unknown) => {
    try {
      const input = navigateToolSchema.parse(args);

      // Validate required fields based on action
      validateNavigateInput(input);

      const cliArgs = buildNavigateCommand(input);

      // Build options object
      const options: any = {
        checkpoint: input.checkpoint,
        position: input.position,
      };

      // Add optional flags
      if (input.newTab) options["new-tab"] = true;
      if (input.wait !== undefined) options.wait = input.wait;
      if (input.smooth) options.smooth = true;
      if (input.block) options.block = input.block;
      if (input.inline) options.inline = input.inline;

      const result = await cli.execute(cliArgs, options);
      const response = formatToolResponse(result, cli.getContext());

      if (!response.success) {
        return {
          content: [
            {
              type: "text",
              text: `❌ Navigation failed: ${response.error}`,
            },
          ],
          isError: true,
        };
      }

      return {
        content: [
          {
            type: "text",
            text: formatNavigateSuccessMessage(input, response.data),
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
  };

  registerTool(toolDefinition, handler);
}

function validateNavigateInput(input: NavigateToolInput) {
  switch (input.action) {
    case "to":
      if (!input.url) {
        throw new Error("URL is required for navigate to action");
      }
      // Validate URL format
      try {
        new URL(input.url);
      } catch {
        throw new Error("Invalid URL format");
      }
      break;

    case "scroll-element":
      if (!input.selector) {
        throw new Error("Selector is required for scroll-element action");
      }
      if (!input.direction) {
        throw new Error(
          "Direction (up/down) is required for scroll-element action",
        );
      }
      break;

    case "scroll-position":
      if (input.x === undefined || input.y === undefined) {
        throw new Error(
          "Both x and y coordinates are required for scroll-position action",
        );
      }
      break;
  }
}

function buildNavigateCommand(input: NavigateToolInput): string[] {
  const args = ["navigate", input.action];

  switch (input.action) {
    case "to":
      args.push(input.url!);
      break;

    case "scroll-element":
      args.push(sanitizeInput(input.selector!));
      args.push(input.direction!);
      break;

    case "scroll-position":
      args.push(String(input.x));
      args.push(String(input.y));
      break;

    // scroll-top and scroll-bottom don't need additional arguments
  }

  return args;
}

function formatNavigateSuccessMessage(
  input: NavigateToolInput,
  data: any,
): string {
  const stepInfo = data.stepId ? ` (Step ${data.stepId})` : "";

  switch (input.action) {
    case "to":
      const tabInfo = input.newTab ? " in new tab" : "";
      return `✅ Created navigation to ${input.url}${tabInfo}${stepInfo}`;

    case "scroll-top":
      return `✅ Created scroll to top of page${stepInfo}`;

    case "scroll-bottom":
      return `✅ Created scroll to bottom of page${stepInfo}`;

    case "scroll-element":
      return `✅ Created scroll ${input.direction} on element "${input.selector}"${stepInfo}`;

    case "scroll-position":
      return `✅ Created scroll to position (${input.x}, ${input.y})${stepInfo}`;

    default:
      return `✅ Created ${input.action} navigation${stepInfo}`;
  }
}
