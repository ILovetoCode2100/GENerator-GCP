import { z } from "zod";
import { VirtuosoCliWrapper } from "../cli-wrapper.js";
import { formatToolResponse } from "../utils/formatting.js";
import { sanitizeInput } from "../utils/validation.js";
import { registerTool } from "./all-tools.js";

const fileToolSchema = z.object({
  url: z.string().url().describe("URL of the file to upload"),
  selector: z.string().describe("File input element selector"),
  checkpoint: z
    .string()
    .optional()
    .describe("Checkpoint ID (optional, uses session context if not provided)"),
  position: z
    .number()
    .optional()
    .describe("Step position (optional, auto-increments if not provided)"),
});

type FileToolInput = z.infer<typeof fileToolSchema>;

export function registerFileTools(cli: VirtuosoCliWrapper) {
  registerTool(
    {
      name: "virtuoso_file_upload",
      description:
        "Upload a file in Virtuoso tests. Uploads a file from a URL to a file input element.",
      inputSchema: {
        type: "object",
        properties: {
          url: {
            type: "string",
            description: "URL of the file to upload (must be a valid URL)",
          },
          selector: {
            type: "string",
            description: "File input element selector",
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
        required: ["url", "selector"],
        additionalProperties: false,
      },
    },
    async (args) => {
      try {
        // Parse and validate input
        const input = fileToolSchema.parse(args);

        // Build CLI command
        const cliArgs = buildFileCommand(input);

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
                text: `❌ File upload failed: ${response.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: formatFileSuccessMessage(input, response.data),
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
 * Build CLI command arguments for file operations
 */
function buildFileCommand(input: FileToolInput): string[] {
  return [
    "file",
    "upload",
    input.url, // URL is not sanitized since it needs to remain a valid URL
    sanitizeInput(input.selector),
  ];
}

/**
 * Format success message for file operations
 */
function formatFileSuccessMessage(input: FileToolInput, data: any): string {
  const stepInfo = data.stepId ? ` (Step ${data.stepId})` : "";
  const fileName = input.url.split("/").pop() || "file";

  return `✅ Created file upload: "${fileName}" to "${input.selector}"${stepInfo}`;
}
