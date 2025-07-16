import { spawn } from "child_process";
import path from "path";
import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export interface CliOptions {
  checkpoint?: string;
  position?: number;
  output?: "json" | "yaml" | "human" | "ai";
  [key: string]: any;
}

export interface CliResult {
  success: boolean;
  data?: any;
  error?: string;
  raw?: string;
}

export class VirtuosoCliWrapper {
  private cliPath: string;
  private config: string;
  private sessionContext: {
    checkpointId?: string;
    position: number;
    journeyId?: string;
    goalId?: string;
  } = { position: 1 };

  constructor(cliPath: string, configPath?: string) {
    this.cliPath = cliPath;
    this.config =
      configPath || path.join(process.cwd(), "virtuoso-config.yaml");
  }

  /**
   * Update session context for subsequent commands
   */
  updateContext(context: Partial<typeof this.sessionContext>) {
    this.sessionContext = { ...this.sessionContext, ...context };
    if (context.position !== undefined) {
      this.sessionContext.position = context.position;
    }
  }

  /**
   * Get current session context
   */
  getContext() {
    return { ...this.sessionContext };
  }

  /**
   * Auto-increment position after successful command
   */
  private incrementPosition() {
    this.sessionContext.position++;
  }

  /**
   * Execute a CLI command with proper error handling
   */
  async execute(args: string[], options?: CliOptions): Promise<CliResult> {
    const fullArgs = [...args];

    // Always use JSON output for MCP
    fullArgs.push("--output", "json");

    // Add config if specified
    if (this.config) {
      fullArgs.push("--config", this.config);
    }

    // Use session context if no explicit checkpoint provided
    const checkpoint = options?.checkpoint || this.sessionContext.checkpointId;
    if (checkpoint && this.requiresCheckpoint(args)) {
      // For commands that use positional checkpoint argument
      if (this.usesPositionalCheckpoint(args[0])) {
        // Insert checkpoint ID after the subcommand
        fullArgs.splice(2, 0, checkpoint);
      } else {
        // Use --checkpoint flag
        fullArgs.push("--checkpoint", checkpoint);
      }
    }

    // Add position if not explicitly provided
    if (options?.position !== undefined) {
      fullArgs.push(String(options.position));
    } else if (this.requiresPosition(args)) {
      fullArgs.push(String(this.sessionContext.position));
    }

    // Add any additional options
    if (options) {
      Object.entries(options).forEach(([key, value]) => {
        if (
          !["checkpoint", "position", "output"].includes(key) &&
          value !== undefined
        ) {
          fullArgs.push(`--${key}`);
          if (value !== true) {
            fullArgs.push(String(value));
          }
        }
      });
    }

    return new Promise((resolve) => {
      console.error(`Executing: ${this.cliPath} ${fullArgs.join(" ")}`);

      const cli = spawn(this.cliPath, fullArgs, {
        env: { ...process.env },
        stdio: ["pipe", "pipe", "pipe"],
      });

      let stdout = "";
      let stderr = "";

      cli.stdout.on("data", (data) => {
        stdout += data;
      });
      cli.stderr.on("data", (data) => {
        stderr += data;
      });

      cli.on("close", (code) => {
        if (code !== 0) {
          resolve({
            success: false,
            error: stderr || `CLI exited with code ${code}`,
            raw: stdout,
          });
        } else {
          try {
            const result = JSON.parse(stdout);
            // Auto-increment position on successful step creation
            if (this.isStepCreationCommand(args)) {
              this.incrementPosition();
            }
            resolve({
              success: true,
              data: result,
              raw: stdout,
            });
          } catch (e) {
            // If not JSON, return raw output
            resolve({
              success: true,
              data: { output: stdout },
              raw: stdout,
            });
          }
        }
      });

      cli.on("error", (err) => {
        resolve({
          success: false,
          error: `Failed to start CLI: ${err.message}`,
        });
      });
    });
  }

  /**
   * Check if command requires checkpoint ID
   */
  private requiresCheckpoint(args: string[]): boolean {
    const noCheckpointCommands = [
      "list-projects",
      "list-goals",
      "list-journeys",
      "create-project",
      "create-goal",
      "library",
    ];
    return !noCheckpointCommands.includes(args[0]);
  }

  /**
   * Check if command uses positional checkpoint argument
   */
  private usesPositionalCheckpoint(command: string): boolean {
    const positionalCommands = [
      "interact",
      "navigate",
      "data",
      "dialog",
      "window",
      "mouse",
      "select",
      "file",
      "misc",
    ];
    return positionalCommands.includes(command);
  }

  /**
   * Check if command requires position
   */
  private requiresPosition(args: string[]): boolean {
    const stepCommands = [
      "assert",
      "interact",
      "navigate",
      "data",
      "dialog",
      "wait",
      "window",
      "mouse",
      "select",
      "file",
      "misc",
    ];
    return stepCommands.includes(args[0]);
  }

  /**
   * Check if command creates a step
   */
  private isStepCreationCommand(args: string[]): boolean {
    return this.requiresPosition(args);
  }
}
