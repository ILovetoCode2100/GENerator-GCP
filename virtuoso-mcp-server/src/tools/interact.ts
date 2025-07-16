import { z } from "zod";
import { VirtuosoCliWrapper } from "../cli-wrapper.js";
import { clickPositionSchema, keyModifierSchema } from "../utils/validation.js";
import { formatToolResponse } from "../utils/formatting.js";
import { sanitizeInput } from "../utils/validation.js";
import { registerTool } from "./all-tools.js";

const interactToolSchema = z.object({
  action: z.enum([
    "click",
    "double-click",
    "right-click",
    "hover",
    "write",
    "key",
  ] as const),
  selector: z.string().optional().describe("Element selector or text"),
  text: z.string().optional().describe("Text to write (for write action)"),
  key: z.string().optional().describe("Key to press (for key action)"),
  checkpoint: z.string().optional(),
  position: z.number().optional(),
  // Optional parameters
  variable: z.string().optional().describe("Store element text in variable"),
  elementType: z
    .string()
    .optional()
    .describe("Element type (BUTTON, LINK, etc.)"),
  clickPosition: clickPositionSchema
    .optional()
    .describe("Click position on element"),
  clear: z.boolean().optional().describe("Clear field before writing"),
  delay: z.number().optional().describe("Delay in milliseconds"),
  target: z.string().optional().describe("Target selector for key action"),
  modifiers: z
    .array(keyModifierSchema)
    .optional()
    .describe("Key modifiers (ctrl, shift, alt, meta)"),
});

type InteractToolInput = z.infer<typeof interactToolSchema>;

export function registerInteractTools(cli: VirtuosoCliWrapper) {
  registerTool(
    {
      name: "virtuoso_interact",
      description:
        "Perform user interactions in Virtuoso tests including clicks, typing, hovering, and keyboard actions.",
      inputSchema: {
        type: "object",
        properties: {
          action: {
            type: "string",
            enum: [
              "click",
              "double-click",
              "right-click",
              "hover",
              "write",
              "key",
            ],
            description: "Type of interaction to perform",
          },
          selector: {
            type: "string",
            description: "Element selector or text (required for most actions)",
          },
          text: {
            type: "string",
            description: "Text to write (required for write action)",
          },
          key: {
            type: "string",
            description:
              "Key to press (required for key action, e.g., ENTER, TAB, ESCAPE)",
          },
          checkpoint: {
            type: "string",
            description: "Checkpoint ID (optional)",
          },
          position: { type: "number", description: "Step position (optional)" },
          variable: {
            type: "string",
            description: "Variable name to store element text",
          },
          elementType: {
            type: "string",
            description: "Element type (BUTTON, LINK, etc.)",
          },
          clickPosition: {
            type: "string",
            enum: [
              "CENTER",
              "TOP_LEFT",
              "TOP_CENTER",
              "TOP_RIGHT",
              "CENTER_LEFT",
              "CENTER_RIGHT",
              "BOTTOM_LEFT",
              "BOTTOM_CENTER",
              "BOTTOM_RIGHT",
            ],
            description: "Position on element to click",
          },
          clear: { type: "boolean", description: "Clear field before writing" },
          delay: { type: "number", description: "Delay in milliseconds" },
          target: {
            type: "string",
            description: "Target selector for key action",
          },
          modifiers: {
            type: "array",
            items: { type: "string", enum: ["ctrl", "shift", "alt", "meta"] },
            description: "Key modifiers",
          },
        },
        required: ["action"],
        additionalProperties: false,
      },
    },
    async (args: unknown) => {
      try {
        const input = interactToolSchema.parse(args);
        const cliArgs = buildInteractCommand(input);

        // Build options object
        const options: any = {
          checkpoint: input.checkpoint,
          position: input.position,
        };

        // Add optional flags
        if (input.variable) options.variable = input.variable;
        if (input.elementType) options["element-type"] = input.elementType;
        if (input.clickPosition) options.position = input.clickPosition;
        if (input.clear) options.clear = true;
        if (input.delay) options.delay = input.delay;
        if (input.target) options.target = input.target;

        const result = await cli.execute(cliArgs, options);
        const response = formatToolResponse(result, cli.getContext());

        if (!response.success) {
          return {
            content: [
              {
                type: "text",
                text: `❌ Interaction failed: ${response.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: formatInteractSuccessMessage(input, response.data),
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

function buildInteractCommand(input: InteractToolInput): string[] {
  const args = ["interact", input.action];

  switch (input.action) {
    case "click":
    case "double-click":
    case "right-click":
    case "hover":
      if (!input.selector) {
        throw new Error(`Selector required for ${input.action} action`);
      }
      args.push(sanitizeInput(input.selector));
      break;

    case "write":
      if (!input.selector || input.text === undefined) {
        throw new Error("Selector and text required for write action");
      }
      args.push(sanitizeInput(input.selector));
      args.push(sanitizeInput(input.text));
      break;

    case "key":
      if (!input.key) {
        throw new Error("Key required for key action");
      }
      // Handle key combinations
      let keyCombo = input.key.toUpperCase();
      if (input.modifiers && input.modifiers.length > 0) {
        const mods = input.modifiers.map((m) => m.toUpperCase()).join("+");
        keyCombo = `${mods}+${keyCombo}`;
      }
      args.push(keyCombo);
      break;
  }

  return args;
}

function formatInteractSuccessMessage(
  input: InteractToolInput,
  data: any,
): string {
  const stepInfo = data.stepId ? ` (Step ${data.stepId})` : "";

  switch (input.action) {
    case "click":
      return `✅ Created click action on "${input.selector}"${stepInfo}`;
    case "double-click":
      return `✅ Created double-click action on "${input.selector}"${stepInfo}`;
    case "right-click":
      return `✅ Created right-click action on "${input.selector}"${stepInfo}`;
    case "hover":
      return `✅ Created hover action on "${input.selector}"${stepInfo}`;
    case "write":
      return `✅ Created write action: "${input.text}" in "${input.selector}"${stepInfo}`;
    case "key":
      const keyInfo = input.target ? ` on "${input.target}"` : "";
      return `✅ Created key press: ${input.key}${keyInfo}${stepInfo}`;
    default:
      return `✅ Created ${input.action} action${stepInfo}`;
  }
}
