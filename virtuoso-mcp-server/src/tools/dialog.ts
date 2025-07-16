import { z } from "zod";
import { VirtuosoCliWrapper } from "../cli-wrapper.js";
import { formatToolResponse } from "../utils/formatting.js";
import { sanitizeInput } from "../utils/validation.js";
import { registerTool } from "./all-tools.js";

const dialogToolSchema = z.object({
  action: z.enum([
    "dismiss-alert",
    "dismiss-confirm",
    "dismiss-prompt",
  ] as const),
  text: z.string().optional().describe("Text to enter in prompt dialog"),
  accept: z
    .boolean()
    .optional()
    .describe("Accept the dialog (for confirm dialogs)"),
  cancel: z.boolean().optional().describe("Cancel the dialog"),
  checkpoint: z.string().optional(),
  position: z.number().optional(),
});

type DialogToolInput = z.infer<typeof dialogToolSchema>;

export function registerDialogTools(cli: VirtuosoCliWrapper) {
  registerTool(
    {
      name: "virtuoso_dialog",
      description:
        "Handle browser dialogs in Virtuoso tests including alerts, confirms, and prompts.",
      inputSchema: {
        type: "object",
        properties: {
          action: {
            type: "string",
            enum: ["dismiss-alert", "dismiss-confirm", "dismiss-prompt"],
            description: "Type of dialog to dismiss",
          },
          text: {
            type: "string",
            description: "Text to enter in prompt dialog (optional)",
          },
          accept: {
            type: "boolean",
            description:
              "Accept the dialog - for confirm dialogs (optional, defaults to true)",
          },
          cancel: {
            type: "boolean",
            description: "Cancel the dialog (optional)",
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
        const input = dialogToolSchema.parse(args);

        // Validate conflicting options
        if (input.accept && input.cancel) {
          throw new Error("Cannot both accept and cancel a dialog");
        }

        const cliArgs = buildDialogCommand(input);

        // Build options object
        const options: any = {
          checkpoint: input.checkpoint,
          position: input.position,
        };

        // Add cancel flag if needed
        if (
          input.cancel ||
          (input.action === "dismiss-confirm" && input.accept === false)
        ) {
          options.cancel = true;
        }

        const result = await cli.execute(cliArgs, options);
        const response = formatToolResponse(result, cli.getContext());

        if (!response.success) {
          return {
            content: [
              {
                type: "text",
                text: `❌ Dialog operation failed: ${response.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: formatDialogSuccessMessage(input, response.data),
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

function buildDialogCommand(input: DialogToolInput): string[] {
  const args = ["dialog", input.action];

  // Add text for prompt dialogs
  if (input.action === "dismiss-prompt" && input.text !== undefined) {
    args.push(sanitizeInput(input.text));
  }

  return args;
}

function formatDialogSuccessMessage(input: DialogToolInput, data: any): string {
  const stepInfo = data.stepId ? ` (Step ${data.stepId})` : "";

  switch (input.action) {
    case "dismiss-alert":
      return `✅ Created dismiss alert dialog${stepInfo}`;

    case "dismiss-confirm":
      const confirmAction =
        input.cancel || input.accept === false ? "cancel" : "accept";
      return `✅ Created dismiss confirm dialog (${confirmAction})${stepInfo}`;

    case "dismiss-prompt":
      if (input.cancel) {
        return `✅ Created dismiss prompt dialog (cancel)${stepInfo}`;
      } else if (input.text !== undefined) {
        return `✅ Created dismiss prompt dialog with text "${input.text}"${stepInfo}`;
      } else {
        return `✅ Created dismiss prompt dialog${stepInfo}`;
      }

    default:
      return `✅ Created ${input.action} dialog operation${stepInfo}`;
  }
}
