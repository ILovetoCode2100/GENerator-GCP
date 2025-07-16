import { z } from "zod";
import { VirtuosoCliWrapper } from "../cli-wrapper.js";
import { formatToolResponse } from "../utils/formatting.js";
import { registerTool } from "./all-tools.js";

const libraryToolSchema = z.object({
  action: z.enum([
    "add",
    "get",
    "attach",
    "move-step",
    "remove-step",
    "update",
  ] as const),
  checkpointId: z
    .string()
    .optional()
    .describe("Checkpoint ID to add to library"),
  libraryCheckpointId: z.string().optional().describe("Library checkpoint ID"),
  journeyId: z.string().optional().describe("Journey ID for attach action"),
  stepId: z
    .string()
    .optional()
    .describe("Test step ID for move/remove actions"),
  position: z.number().optional().describe("Position for attach or move-step"),
  title: z.string().optional().describe("New title for update action"),
});

type LibraryToolInput = z.infer<typeof libraryToolSchema>;

export function registerLibraryTools(cli: VirtuosoCliWrapper) {
  registerTool(
    {
      name: "virtuoso_library",
      description:
        "Manage Virtuoso library checkpoints. Add checkpoints to library, get details, attach to journeys, and manage test steps within library checkpoints.",
      inputSchema: {
        type: "object",
        properties: {
          action: {
            type: "string",
            enum: [
              "add",
              "get",
              "attach",
              "move-step",
              "remove-step",
              "update",
            ],
            description: "Library action to perform",
          },
          checkpointId: {
            type: "string",
            description:
              "Checkpoint ID to add to library (required for add action)",
          },
          libraryCheckpointId: {
            type: "string",
            description:
              "Library checkpoint ID (required for get, attach, move-step, remove-step, update)",
          },
          journeyId: {
            type: "string",
            description: "Journey ID (required for attach action)",
          },
          stepId: {
            type: "string",
            description:
              "Test step ID (required for move-step and remove-step actions)",
          },
          position: {
            type: "number",
            description: "Position for attach or move-step actions",
          },
          title: {
            type: "string",
            description:
              "New title for library checkpoint (required for update action)",
          },
        },
        required: ["action"],
        additionalProperties: false,
      },
    },
    async (args) => {
      try {
        const input = libraryToolSchema.parse(args);

        // Validate required parameters for each action
        validateLibraryAction(input);

        const cliArgs = buildLibraryCommand(input);

        // Library commands don't use standard checkpoint/position context
        const result = await cli.execute(cliArgs, {});
        const response = formatToolResponse(result, cli.getContext());

        if (!response.success) {
          return {
            content: [
              {
                type: "text",
                text: `❌ Library operation failed: ${response.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: formatLibrarySuccessMessage(input, response.data),
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

function validateLibraryAction(input: LibraryToolInput) {
  switch (input.action) {
    case "add":
      if (!input.checkpointId) {
        throw new Error("Checkpoint ID is required for add action");
      }
      break;

    case "get":
      if (!input.libraryCheckpointId) {
        throw new Error("Library checkpoint ID is required for get action");
      }
      break;

    case "attach":
      if (
        !input.journeyId ||
        !input.libraryCheckpointId ||
        input.position === undefined
      ) {
        throw new Error(
          "Journey ID, library checkpoint ID, and position are required for attach action",
        );
      }
      break;

    case "move-step":
      if (
        !input.libraryCheckpointId ||
        !input.stepId ||
        input.position === undefined
      ) {
        throw new Error(
          "Library checkpoint ID, step ID, and position are required for move-step action",
        );
      }
      break;

    case "remove-step":
      if (!input.libraryCheckpointId || !input.stepId) {
        throw new Error(
          "Library checkpoint ID and step ID are required for remove-step action",
        );
      }
      break;

    case "update":
      if (!input.libraryCheckpointId || !input.title) {
        throw new Error(
          "Library checkpoint ID and title are required for update action",
        );
      }
      break;
  }
}

function buildLibraryCommand(input: LibraryToolInput): string[] {
  const args = ["library", input.action];

  switch (input.action) {
    case "add":
      args.push(input.checkpointId!);
      break;

    case "get":
      args.push(input.libraryCheckpointId!);
      break;

    case "attach":
      args.push(input.journeyId!);
      args.push(input.libraryCheckpointId!);
      args.push(String(input.position));
      break;

    case "move-step":
      args.push(input.libraryCheckpointId!);
      args.push(input.stepId!);
      args.push(String(input.position));
      break;

    case "remove-step":
      args.push(input.libraryCheckpointId!);
      args.push(input.stepId!);
      break;

    case "update":
      args.push(input.libraryCheckpointId!);
      args.push(input.title!);
      break;
  }

  return args;
}

function formatLibrarySuccessMessage(
  input: LibraryToolInput,
  data: any,
): string {
  switch (input.action) {
    case "add":
      return `✅ Added checkpoint ${input.checkpointId} to library${
        data.libraryCheckpointId
          ? ` (Library ID: ${data.libraryCheckpointId})`
          : ""
      }`;

    case "get":
      return `✅ Retrieved library checkpoint ${input.libraryCheckpointId}`;

    case "attach":
      return `✅ Attached library checkpoint ${input.libraryCheckpointId} to journey ${input.journeyId} at position ${input.position}`;

    case "move-step":
      return `✅ Moved step ${input.stepId} to position ${input.position} in library checkpoint ${input.libraryCheckpointId}`;

    case "remove-step":
      return `✅ Removed step ${input.stepId} from library checkpoint ${input.libraryCheckpointId}`;

    case "update":
      return `✅ Updated library checkpoint ${input.libraryCheckpointId} title to "${input.title}"`;

    default:
      return `✅ Completed ${input.action} operation`;
  }
}
