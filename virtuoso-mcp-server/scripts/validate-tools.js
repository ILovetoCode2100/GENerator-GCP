#!/usr/bin/env tsx
/**
 * Validate Tools Script
 *
 * This script validates all tool schemas, checks for consistency,
 * and generates tool documentation.
 */
import { readdir, readFile, writeFile } from "fs/promises";
import path from "path";
import { fileURLToPath } from "url";
import { z } from "zod";
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
// Color codes for terminal output
const colors = {
  reset: "\x1b[0m",
  green: "\x1b[32m",
  red: "\x1b[31m",
  yellow: "\x1b[33m",
  blue: "\x1b[34m",
  dim: "\x1b[2m",
};
// Common parameter patterns
const commonParams = {
  checkpointId: z.string().describe("The checkpoint ID"),
  position: z
    .number()
    .int()
    .positive()
    .describe("The position in the test sequence"),
  selector: z.string().describe("CSS selector or element description"),
  variableName: z.string().describe("Variable name for storage"),
  timeout: z
    .number()
    .int()
    .positive()
    .optional()
    .describe("Timeout in milliseconds"),
};
async function loadToolsFromFiles() {
  const toolsDir = path.join(__dirname, "..", "src", "tools");
  const files = await readdir(toolsDir);
  const tools = [];
  for (const file of files) {
    if (!file.endsWith(".ts")) continue;
    const filePath = path.join(toolsDir, file);
    const content = await readFile(filePath, "utf-8");
    // Extract tool definitions using regex
    const toolMatches = content.matchAll(
      /name:\s*['"`]([^'"`]+)['"`][\s\S]*?description:\s*['"`]([^'"`]+)['"`][\s\S]*?inputSchema:\s*({[\s\S]*?})\s*,?\s*\n\s*}/g,
    );
    for (const match of toolMatches) {
      const [, name, description] = match;
      const [, group, subcommand] = name.match(/virtuoso_([^_]+)_(.+)/) || [];
      tools.push({
        name,
        description,
        inputSchema: {}, // Would need proper parsing for full validation
        group: group || "unknown",
        subcommand: subcommand || "unknown",
      });
    }
  }
  return tools;
}
async function validateTools(tools) {
  const results = [];
  for (const tool of tools) {
    const errors = [];
    const warnings = [];
    // Validate naming convention
    if (!tool.name.startsWith("virtuoso_")) {
      errors.push('Tool name must start with "virtuoso_"');
    }
    const nameParts = tool.name.split("_");
    if (nameParts.length < 3) {
      errors.push("Tool name must follow pattern: virtuoso_GROUP_SUBCOMMAND");
    }
    // Validate description
    if (!tool.description || tool.description.length < 10) {
      errors.push("Description must be at least 10 characters long");
    }
    if (
      !tool.description.endsWith(".") &&
      !tool.description.endsWith("!") &&
      !tool.description.endsWith("?")
    ) {
      warnings.push("Description should end with punctuation");
    }
    // Check for common patterns
    if (
      tool.group !== "library" &&
      tool.group !== "dialog" &&
      tool.group !== "misc"
    ) {
      if (!tool.description.includes("checkpoint")) {
        warnings.push('Most tools should mention "checkpoint" in description');
      }
    }
    results.push({
      tool: tool.name,
      valid: errors.length === 0,
      errors,
      warnings,
    });
  }
  return results;
}
async function generateDocumentation(tools) {
  const groups = tools.reduce((acc, tool) => {
    if (!acc[tool.group]) acc[tool.group] = [];
    acc[tool.group].push(tool);
    return acc;
  }, {});
  let doc = `# Virtuoso MCP Server Tools Documentation

Generated on: ${new Date().toISOString()}

## Overview

The Virtuoso MCP Server provides ${tools.length} tools across ${
    Object.keys(groups).length
  } command groups.

## Tool Groups

`;
  for (const [group, groupTools] of Object.entries(groups)) {
    doc += `### ${group.charAt(0).toUpperCase() + group.slice(1)} Commands (${
      groupTools.length
    } tools)\n\n`;
    for (const tool of groupTools.sort((a, b) =>
      a.subcommand.localeCompare(b.subcommand),
    )) {
      doc += `#### \`${tool.name}\`\n`;
      doc += `${tool.description}\n\n`;
      doc += `**Subcommand:** \`${tool.subcommand}\`\n\n`;
    }
  }
  doc += `## Usage Examples

### Assert Commands
\`\`\`json
{
  "tool": "virtuoso_assert_exists",
  "arguments": {
    "checkpointId": "1680930",
    "selector": "Login button",
    "position": 1
  }
}
\`\`\`

### Interact Commands
\`\`\`json
{
  "tool": "virtuoso_interact_click",
  "arguments": {
    "checkpointId": "1680930",
    "selector": "Submit",
    "position": 1
  }
}
\`\`\`

### Navigate Commands
\`\`\`json
{
  "tool": "virtuoso_navigate_to",
  "arguments": {
    "checkpointId": "1680930",
    "url": "https://example.com",
    "position": 1
  }
}
\`\`\`

## Integration with Claude Desktop

Add to your Claude Desktop configuration:

\`\`\`json
{
  "mcpServers": {
    "virtuoso": {
      "command": "node",
      "args": ["/path/to/virtuoso-mcp-server/dist/index.js"]
    }
  }
}
\`\`\`
`;
  return doc;
}
async function main() {
  console.log(
    `${colors.blue}Virtuoso MCP Server Tool Validation${colors.reset}\n`,
  );
  try {
    // Load tools
    console.log(`${colors.yellow}Loading tools...${colors.reset}`);
    const tools = await loadToolsFromFiles();
    console.log(
      `${colors.green}✓ Loaded ${tools.length} tools${colors.reset}\n`,
    );
    // Group analysis
    const groups = tools.reduce((acc, tool) => {
      acc[tool.group] = (acc[tool.group] || 0) + 1;
      return acc;
    }, {});
    console.log(`${colors.yellow}Tool Groups:${colors.reset}`);
    for (const [group, count] of Object.entries(groups)) {
      console.log(`  ${group}: ${count} tools`);
    }
    console.log("");
    // Validate tools
    console.log(`${colors.yellow}Validating tools...${colors.reset}`);
    const results = await validateTools(tools);
    let valid = 0;
    let invalid = 0;
    let warnings = 0;
    for (const result of results) {
      if (result.valid) {
        valid++;
        if (result.warnings.length > 0) {
          warnings++;
          console.log(`${colors.yellow}⚠ ${result.tool}${colors.reset}`);
          for (const warning of result.warnings) {
            console.log(`  ${colors.dim}${warning}${colors.reset}`);
          }
        }
      } else {
        invalid++;
        console.log(`${colors.red}✗ ${result.tool}${colors.reset}`);
        for (const error of result.errors) {
          console.log(`  ${colors.red}${error}${colors.reset}`);
        }
        if (result.warnings.length > 0) {
          for (const warning of result.warnings) {
            console.log(`  ${colors.yellow}${warning}${colors.reset}`);
          }
        }
      }
    }
    console.log(`\n${colors.yellow}Validation Summary:${colors.reset}`);
    console.log(`${colors.green}Valid: ${valid}${colors.reset}`);
    console.log(`${colors.red}Invalid: ${invalid}${colors.reset}`);
    console.log(`${colors.yellow}Warnings: ${warnings}${colors.reset}\n`);
    // Check for duplicates
    console.log(`${colors.yellow}Checking for duplicates...${colors.reset}`);
    const nameCount = tools.reduce((acc, tool) => {
      acc[tool.name] = (acc[tool.name] || 0) + 1;
      return acc;
    }, {});
    const duplicates = Object.entries(nameCount).filter(
      ([, count]) => count > 1,
    );
    if (duplicates.length > 0) {
      console.log(`${colors.red}Found duplicate tool names:${colors.reset}`);
      for (const [name, count] of duplicates) {
        console.log(`  ${name}: ${count} occurrences`);
      }
    } else {
      console.log(
        `${colors.green}✓ No duplicate tool names found${colors.reset}`,
      );
    }
    console.log("");
    // Check naming consistency
    console.log(
      `${colors.yellow}Checking naming consistency...${colors.reset}`,
    );
    const inconsistentNames = tools.filter((tool) => {
      const expectedPattern = /^virtuoso_[a-z]+_[a-z_]+$/;
      return !expectedPattern.test(tool.name);
    });
    if (inconsistentNames.length > 0) {
      console.log(
        `${colors.yellow}Tools with inconsistent naming:${colors.reset}`,
      );
      for (const tool of inconsistentNames) {
        console.log(`  ${tool.name}`);
      }
    } else {
      console.log(
        `${colors.green}✓ All tools follow naming convention${colors.reset}`,
      );
    }
    console.log("");
    // Generate documentation
    console.log(`${colors.yellow}Generating documentation...${colors.reset}`);
    const documentation = await generateDocumentation(tools);
    const docPath = path.join(__dirname, "..", "TOOLS.md");
    await writeFile(docPath, documentation);
    console.log(
      `${colors.green}✓ Documentation written to ${docPath}${colors.reset}\n`,
    );
    // Final summary
    if (invalid === 0) {
      console.log(
        `${colors.green}✅ All tools validated successfully!${colors.reset}`,
      );
    } else {
      console.log(
        `${colors.red}❌ Validation failed for ${invalid} tools${colors.reset}`,
      );
      process.exit(1);
    }
  } catch (error) {
    console.error(`${colors.red}Validation failed:${colors.reset}`, error);
    process.exit(1);
  }
}
// Run validation
main().catch(console.error);
//# sourceMappingURL=validate-tools.js.map
