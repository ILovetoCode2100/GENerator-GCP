# Virtuoso API CLI Documentation CSV Files

This directory contains comprehensive CSV documentation for the Virtuoso API CLI, generated through detailed analysis of the codebase and testing results.

## CSV Files Overview

### 1. **cli-commands-master.csv** (Main Reference)

- Complete inventory of all 120 CLI commands
- Includes command syntax, API endpoints, status, parameters
- Shows which commands work vs fail due to API limitations
- Primary reference for all CLI functionality

### 2. **api-endpoints-analysis.csv**

- Maps API endpoints to command groups
- Shows request/response structures
- Documents which endpoints serve multiple commands
- Useful for understanding API architecture

### 3. **parameter-patterns.csv**

- Documents different parameter patterns across command groups
- Shows checkpoint specification methods (flag vs positional)
- Explains session context support
- Lists all optional parameters by command type

### 4. **special-parameters-reference.csv**

- Complete reference for all enums and special values
- Position enums, keyboard modifiers, output formats
- Time values, coordinates, boolean flags
- Validation rules and constraints

### 5. **status-summary.csv**

- High-level success/failure statistics
- Breakdown by command group and failure type
- Overall success rate: 90.8% (109/120 commands)
- Categorizes failures: API limitations vs implementation

### 6. **Individual Command Group CSVs**

- interact-commands-inventory.csv
- navigate-commands-inventory.csv
- data-commands-inventory.csv
- Additional detailed analysis for specific command groups

## Key Findings

### Success Rate: 90.8% (109/120 commands working)

### Failures by Category:

- **API Limitations (8)**: Commands the API doesn't support
  - Navigate: back, forward, refresh (5)
  - Window: frame-index, frame-name, main-content (3)
- **File System Limitation (1)**: Local file upload not supported
- **External Dependencies (2)**: Library commands need valid IDs
- **Window Close (1)**: Not supported by API

### 100% Working Command Groups:

- Assert (12/12)
- Interact (30/30)
- Data (12/12)
- Dialog (6/6)
- Wait (6/6)
- Mouse (6/6)
- Select (3/3)
- Misc (2/2)

## Usage

These CSV files can be:

- Imported into spreadsheet software for analysis
- Used as reference documentation
- Converted to other formats (JSON, markdown tables)
- Used for generating automated tests
- Referenced for API integration planning

## Maintenance

When adding new commands or updating existing ones:

1. Update the relevant CSV files
2. Ensure consistency across all documentation
3. Update status-summary.csv with new totals
4. Test and verify all status indicators
