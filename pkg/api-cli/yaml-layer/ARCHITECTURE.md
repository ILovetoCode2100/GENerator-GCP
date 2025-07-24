# Virtuoso YAML Test Layer Architecture

## Overview

The YAML test layer provides a token-efficient, AI-friendly interface for creating Virtuoso tests. It compiles concise YAML definitions into Virtuoso API commands while providing comprehensive validation and helpful error messages.

## Architecture Flow

```
User Input → YAML → Parser → Validator → Preprocessor → Compiler → Virtuoso API
     ↑                           ↓
     └── AI Templates ← Error Messages
```

## Core Components

### 1. YAML Parser

- Parses minimal YAML syntax
- Supports shortcuts and abbreviations
- Handles multi-document files

### 2. Schema Validator

- Validates YAML structure
- Type checking
- Required field validation
- Cross-reference validation

### 3. Semantic Validator

- Validates logical correctness
- Checks selector validity
- Ensures action sequences make sense
- Validates data references

### 4. Preprocessor

- Expands shortcuts
- Resolves variables
- Handles includes/imports
- Manages contexts

### 5. Compiler

- Converts YAML to API commands
- Optimizes command sequences
- Handles error recovery
- Generates execution plan

### 6. Execution Engine

- Manages API calls
- Handles retries
- Tracks progress
- Reports results

## Key Design Principles

1. **Minimal Syntax**: Every character counts
2. **Progressive Disclosure**: Simple things simple, complex things possible
3. **Fail Fast**: Validate early with helpful messages
4. **AI Guidance**: Templates and patterns for efficient generation
5. **Extensibility**: Easy to add new features without breaking existing tests
