import { z } from "zod";
import { VirtuosoCliWrapper } from "../cli-wrapper.js";
import { cookieSameSiteSchema } from "../utils/validation.js";
import { formatToolResponse } from "../utils/formatting.js";
import { sanitizeInput, isValidVariableName } from "../utils/validation.js";
import { registerTool } from "./all-tools.js";

const dataToolSchema = z.object({
  action: z.enum([
    "store-text",
    "store-value",
    "cookie-create",
    "cookie-delete",
    "cookie-clear",
  ] as const),
  element: z
    .string()
    .optional()
    .describe("Element selector for store operations"),
  variable: z.string().optional().describe("Variable name to store data"),
  name: z.string().optional().describe("Cookie name"),
  value: z.string().optional().describe("Cookie value"),
  checkpoint: z.string().optional(),
  position: z.number().optional(),
  // Cookie-specific options
  domain: z.string().optional().describe("Cookie domain"),
  path: z.string().optional().describe("Cookie path"),
  secure: z.boolean().optional().describe("Secure cookie flag"),
  httpOnly: z.boolean().optional().describe("HTTP-only cookie flag"),
  sameSite: cookieSameSiteSchema
    .optional()
    .describe("Same-site cookie attribute"),
  maxAge: z.number().optional().describe("Max age in seconds"),
  expires: z.string().optional().describe("Expiration date"),
});

type DataToolInput = z.infer<typeof dataToolSchema>;

export function registerDataTools(cli: VirtuosoCliWrapper) {
  registerTool(
    {
      name: "virtuoso_data",
      description:
        "Manage data storage and cookies in Virtuoso tests. Store element text/values in variables and manage browser cookies.",
      inputSchema: {
        type: "object",
        properties: {
          action: {
            type: "string",
            enum: [
              "store-text",
              "store-value",
              "cookie-create",
              "cookie-delete",
              "cookie-clear",
            ],
            description: "Type of data operation",
          },
          element: {
            type: "string",
            description: "Element selector (required for store operations)",
          },
          variable: {
            type: "string",
            description: "Variable name (required for store operations)",
          },
          name: {
            type: "string",
            description:
              "Cookie name (required for cookie operations except clear)",
          },
          value: {
            type: "string",
            description: "Cookie value (required for cookie-create)",
          },
          checkpoint: {
            type: "string",
            description: "Checkpoint ID (optional)",
          },
          position: { type: "number", description: "Step position (optional)" },
          // Cookie options
          domain: { type: "string", description: "Cookie domain" },
          path: { type: "string", description: "Cookie path" },
          secure: { type: "boolean", description: "Secure cookie flag" },
          httpOnly: { type: "boolean", description: "HTTP-only cookie flag" },
          sameSite: {
            type: "string",
            enum: ["strict", "lax", "none"],
            description: "Same-site cookie attribute",
          },
          maxAge: { type: "number", description: "Max age in seconds" },
          expires: { type: "string", description: "Expiration date" },
        },
        required: ["action"],
        additionalProperties: false,
      },
    },
    async (args) => {
      try {
        const input = dataToolSchema.parse(args);

        // Validate required fields based on action
        validateDataInput(input);

        const cliArgs = buildDataCommand(input);

        // Build options object
        const options: any = {
          checkpoint: input.checkpoint,
          position: input.position,
        };

        // Add cookie-specific options
        if (input.action === "cookie-create") {
          if (input.domain) options.domain = input.domain;
          if (input.path) options.path = input.path;
          if (input.secure) options.secure = true;
          if (input.httpOnly) options["http-only"] = true;
          if (input.sameSite) options["same-site"] = input.sameSite;
          if (input.maxAge) options["max-age"] = input.maxAge;
          if (input.expires) options.expires = input.expires;
        }

        const result = await cli.execute(cliArgs, options);
        const response = formatToolResponse(result, cli.getContext());

        if (!response.success) {
          return {
            content: [
              {
                type: "text",
                text: `❌ Data operation failed: ${response.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: formatDataSuccessMessage(input, response.data),
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

function validateDataInput(input: DataToolInput) {
  switch (input.action) {
    case "store-text":
    case "store-value":
      if (!input.element || !input.variable) {
        throw new Error(
          `Element and variable are required for ${input.action}`,
        );
      }
      if (!isValidVariableName(input.variable)) {
        throw new Error(
          "Variable name must start with a letter or underscore and contain only alphanumeric characters",
        );
      }
      break;

    case "cookie-create":
      if (!input.name || !input.value) {
        throw new Error("Cookie name and value are required for cookie-create");
      }
      break;

    case "cookie-delete":
      if (!input.name) {
        throw new Error("Cookie name is required for cookie-delete");
      }
      break;

    // cookie-clear doesn't require additional parameters
  }
}

function buildDataCommand(input: DataToolInput): string[] {
  const args = ["data", input.action];

  switch (input.action) {
    case "store-text":
    case "store-value":
      args.push(sanitizeInput(input.element!));
      args.push(input.variable!);
      break;

    case "cookie-create":
      args.push(input.name!);
      args.push(sanitizeInput(input.value!));
      break;

    case "cookie-delete":
      args.push(input.name!);
      break;

    // cookie-clear doesn't need additional arguments
  }

  return args;
}

function formatDataSuccessMessage(input: DataToolInput, data: any): string {
  const stepInfo = data.stepId ? ` (Step ${data.stepId})` : "";

  switch (input.action) {
    case "store-text":
      return `✅ Created store text from "${input.element}" in variable "${input.variable}"${stepInfo}`;

    case "store-value":
      return `✅ Created store value from "${input.element}" in variable "${input.variable}"${stepInfo}`;

    case "cookie-create":
      const cookieDetails = [];
      if (input.domain) cookieDetails.push(`domain: ${input.domain}`);
      if (input.path) cookieDetails.push(`path: ${input.path}`);
      if (input.secure) cookieDetails.push("secure");
      if (input.httpOnly) cookieDetails.push("httpOnly");
      const details =
        cookieDetails.length > 0 ? ` (${cookieDetails.join(", ")})` : "";
      return `✅ Created cookie "${input.name}" with value "${input.value}"${details}${stepInfo}`;

    case "cookie-delete":
      return `✅ Created delete cookie "${input.name}"${stepInfo}`;

    case "cookie-clear":
      return `✅ Created clear all cookies${stepInfo}`;

    default:
      return `✅ Created ${input.action} operation${stepInfo}`;
  }
}
