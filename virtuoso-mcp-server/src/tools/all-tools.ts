import { Tool } from "@modelcontextprotocol/sdk/types.js";

// Global registry for all tools
export const toolRegistry: Tool[] = [];
export const toolHandlers = new Map<string, (args: any) => Promise<any>>();

export function registerTool(tool: Tool, handler: (args: any) => Promise<any>) {
  toolRegistry.push(tool);
  toolHandlers.set(tool.name, handler);
}
