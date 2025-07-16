import { VirtuosoStep, ToolResponse } from "../types/virtuoso.js";

/**
 * Format tool response for MCP
 */
export function formatToolResponse(result: any, context?: any): ToolResponse {
  // Handle different response formats from CLI
  if (result.success === false) {
    return {
      success: false,
      error: result.error || "Unknown error",
      context,
    };
  }

  // Format successful responses
  if (result.data) {
    return {
      success: true,
      data: formatResponseData(result.data),
      context,
    };
  }

  return {
    success: true,
    data: result,
    context,
  };
}

/**
 * Format response data based on type
 */
function formatResponseData(data: any): any {
  // Handle step creation responses
  if (data.step_id !== undefined || data.stepId !== undefined) {
    return {
      stepId: data.step_id || data.stepId,
      type: data.type,
      position: data.position,
      description: data.description,
      status: "created",
    };
  }

  // Handle list responses
  if (Array.isArray(data)) {
    return {
      items: data,
      count: data.length,
    };
  }

  // Handle raw responses
  if (typeof data === "string") {
    return {
      message: data,
    };
  }

  return data;
}

/**
 * Format step description for display
 */
export function formatStepDescription(step: VirtuosoStep): string {
  const parts = [`Step ${step.id}: ${step.type}`];

  if (step.description) {
    parts.push(`- ${step.description}`);
  }

  if (step.position) {
    parts.push(`(Position: ${step.position})`);
  }

  return parts.join(" ");
}

/**
 * Format list output for display
 */
export function formatList(items: any[], fields: string[]): string {
  if (items.length === 0) {
    return "No items found";
  }

  const lines: string[] = [];

  // Header
  lines.push(fields.join(" | "));
  lines.push("-".repeat(fields.join(" | ").length));

  // Items
  items.forEach((item) => {
    const values = fields.map((field) => {
      const value = item[field];
      return value !== undefined ? String(value) : "-";
    });
    lines.push(values.join(" | "));
  });

  return lines.join("\n");
}

/**
 * Format error details for debugging
 */
export function formatErrorDetails(error: any): string {
  const details: string[] = [];

  if (error.message) {
    details.push(`Error: ${error.message}`);
  }

  if (error.code) {
    details.push(`Code: ${error.code}`);
  }

  if (error.stack && process.env.DEBUG) {
    details.push(`Stack:\n${error.stack}`);
  }

  return details.join("\n");
}

/**
 * Format CLI arguments for logging
 */
export function formatCliArgs(args: string[]): string {
  return args
    .map((arg) => {
      // Quote arguments with spaces
      if (arg.includes(" ")) {
        return `"${arg}"`;
      }
      return arg;
    })
    .join(" ");
}

/**
 * Parse and format checkpoint context
 */
export function formatCheckpointContext(checkpoint: any): string {
  if (typeof checkpoint === "string") {
    return `Checkpoint ID: ${checkpoint}`;
  }

  if (checkpoint && typeof checkpoint === "object") {
    const parts = [];
    if (checkpoint.id) parts.push(`ID: ${checkpoint.id}`);
    if (checkpoint.name) parts.push(`Name: ${checkpoint.name}`);
    if (checkpoint.url) parts.push(`URL: ${checkpoint.url}`);
    return parts.join(", ");
  }

  return "No checkpoint context";
}
