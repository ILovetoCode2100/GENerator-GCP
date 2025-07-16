import { z } from "zod";

/**
 * Common validation schemas for Virtuoso MCP tools
 */

// Base schemas
export const checkpointSchema = z
  .string()
  .regex(/^\d+$/, "Checkpoint ID must be numeric");
export const positionSchema = z.number().int().positive();
export const selectorSchema = z.string().min(1, "Selector cannot be empty");
export const urlSchema = z.string().url("Invalid URL format");

// Element position schema
export const clickPositionSchema = z.enum([
  "CENTER",
  "TOP_LEFT",
  "TOP_CENTER",
  "TOP_RIGHT",
  "CENTER_LEFT",
  "CENTER_RIGHT",
  "BOTTOM_LEFT",
  "BOTTOM_CENTER",
  "BOTTOM_RIGHT",
]);

// Scroll schemas
export const scrollBlockSchema = z.enum(["start", "center", "end", "nearest"]);
export const scrollInlineSchema = z.enum(["start", "center", "end", "nearest"]);

// Cookie schemas
export const cookieSameSiteSchema = z.enum(["strict", "lax", "none"]);

// Key modifier schema
export const keyModifierSchema = z.enum([
  "ctrl",
  "shift",
  "alt",
  "meta",
  "cmd",
]);

/**
 * Validate selector format
 */
export function isValidSelector(selector: string): boolean {
  // Check for CSS selector patterns
  const cssPatterns = [
    /^#[\w-]+$/, // ID selector
    /^\.[\w-]+$/, // Class selector
    /^\w+$/, // Tag selector
    /^\[[\w-]+(=[^\]]+)?\]$/, // Attribute selector
    /^[\w\s>+~,:.#\[\]="'-]+$/, // Complex selector
  ];

  return cssPatterns.some((pattern) => pattern.test(selector));
}

/**
 * Validate XPath format
 */
export function isValidXPath(xpath: string): boolean {
  return xpath.startsWith("//") || xpath.startsWith("/");
}

/**
 * Validate regex pattern
 */
export function isValidRegex(pattern: string): boolean {
  try {
    new RegExp(pattern);
    return true;
  } catch {
    return false;
  }
}

/**
 * Sanitize user input for CLI arguments
 */
export function sanitizeInput(input: string): string {
  // Remove potential command injection characters
  return input
    .replace(/[;&|`$<>]/g, "")
    .replace(/\\/g, "\\\\")
    .replace(/"/g, '\\"')
    .trim();
}

/**
 * Parse position from various formats
 */
export function parsePosition(
  position: string | number | undefined,
): number | undefined {
  if (position === undefined) return undefined;
  if (typeof position === "number") return position;

  const parsed = parseInt(position, 10);
  return isNaN(parsed) ? undefined : parsed;
}

/**
 * Format error messages for MCP responses
 */
export function formatError(error: Error | string): string {
  if (typeof error === "string") return error;

  // Extract meaningful error message
  const message = error.message || "Unknown error";

  // Common error patterns
  if (message.includes("checkpoint not found")) {
    return "Checkpoint not found. Please verify the checkpoint ID.";
  }
  if (message.includes("authentication")) {
    return "Authentication failed. Please check your API token.";
  }
  if (message.includes("network")) {
    return "Network error. Please check your connection.";
  }

  return message;
}

/**
 * Validate variable name format
 */
export function isValidVariableName(name: string): boolean {
  return /^[a-zA-Z_][a-zA-Z0-9_]*$/.test(name);
}
