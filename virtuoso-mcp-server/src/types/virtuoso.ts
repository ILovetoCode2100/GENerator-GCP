// Common types for Virtuoso API CLI MCP Server

export interface VirtuosoStep {
  id: number;
  type: string;
  position: number;
  description?: string;
  meta?: Record<string, any>;
}

export interface VirtuosoCheckpoint {
  id: number;
  name: string;
  url: string;
  journeyId: number;
  position: number;
}

export interface VirtuosoJourney {
  id: number;
  name: string;
  goalId: number;
  checkpoints?: VirtuosoCheckpoint[];
}

export interface VirtuosoGoal {
  id: number;
  name: string;
  projectId: number;
  journeys?: VirtuosoJourney[];
}

export interface VirtuosoProject {
  id: number;
  name: string;
  organizationId: number;
}

// Tool parameter types
export type AssertType =
  | "exists"
  | "not-exists"
  | "equals"
  | "not-equals"
  | "checked"
  | "selected"
  | "gt"
  | "gte"
  | "lt"
  | "lte"
  | "matches"
  | "variable";

export type InteractAction =
  | "click"
  | "double-click"
  | "right-click"
  | "hover"
  | "write"
  | "key";

export type NavigateAction =
  | "to"
  | "scroll-top"
  | "scroll-bottom"
  | "scroll-element"
  | "scroll-position";

export type DataAction =
  | "store-text"
  | "store-value"
  | "cookie-create"
  | "cookie-delete"
  | "cookie-clear";

export type DialogAction =
  | "dismiss-alert"
  | "dismiss-confirm"
  | "dismiss-prompt";

export type WaitType = "element" | "time";

export type WindowAction = "resize" | "switch-tab" | "switch-frame";

export type MouseAction =
  | "move-to"
  | "move-by"
  | "move"
  | "down"
  | "up"
  | "enter";

export type SelectAction = "option" | "index" | "last";

export type FileAction = "upload";

export type MiscAction = "comment" | "execute-script" | "key";

export type LibraryAction =
  | "add"
  | "get"
  | "attach"
  | "move-step"
  | "remove-step"
  | "update";

// Position types
export type ClickPosition =
  | "CENTER"
  | "TOP_LEFT"
  | "TOP_CENTER"
  | "TOP_RIGHT"
  | "CENTER_LEFT"
  | "CENTER_RIGHT"
  | "BOTTOM_LEFT"
  | "BOTTOM_CENTER"
  | "BOTTOM_RIGHT";

export type ScrollBlock = "start" | "center" | "end" | "nearest";
export type ScrollInline = "start" | "center" | "end" | "nearest";

// Common options
export interface BaseToolOptions {
  checkpoint?: string;
  position?: number;
}

export interface ToolResponse {
  success: boolean;
  data?: any;
  error?: string;
  context?: {
    checkpointId?: string;
    position?: number;
  };
}
