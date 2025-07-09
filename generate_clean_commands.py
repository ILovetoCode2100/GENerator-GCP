#!/usr/bin/env python3
"""
Script to generate clean step command files without backward compatibility.
This removes the need for manual checkpoint ID management.
"""

import os
import re
import glob

# Command patterns for different step types
COMMAND_PATTERNS = {
    # Single element parameter
    'single_element': {
        'args': 'ELEMENT [POSITION]',
        'cobra_args': 'cobra.RangeArgs(1, 2)',
        'position_index': 1,
        'params': ['element := args[0]'],
        'validation': ['if element == "" { return fmt.Errorf("element cannot be empty") }'],
        'extra': 'map[string]interface{}{"element": element}'
    },
    
    # Text and element parameters
    'text_element': {
        'args': 'TEXT ELEMENT [POSITION]',
        'cobra_args': 'cobra.RangeArgs(2, 3)',
        'position_index': 2,
        'params': ['text := args[0]', 'element := args[1]'],
        'validation': [
            'if text == "" { return fmt.Errorf("text cannot be empty") }',
            'if element == "" { return fmt.Errorf("element cannot be empty") }'
        ],
        'extra': 'map[string]interface{}{"text": text, "element": element}'
    },
    
    # Element and value parameters
    'element_value': {
        'args': 'ELEMENT VALUE [POSITION]',
        'cobra_args': 'cobra.RangeArgs(2, 3)',
        'position_index': 2,
        'params': ['element := args[0]', 'value := args[1]'],
        'validation': [
            'if element == "" { return fmt.Errorf("element cannot be empty") }',
            'if value == "" { return fmt.Errorf("value cannot be empty") }'
        ],
        'extra': 'map[string]interface{}{"element": element, "value": value}'
    },
    
    # Position only (no parameters)
    'position_only': {
        'args': '[POSITION]',
        'cobra_args': 'cobra.RangeArgs(0, 1)',
        'position_index': 0,
        'params': [],
        'validation': [],
        'extra': 'nil'
    },
    
    # Single value parameter
    'single_value': {
        'args': 'VALUE [POSITION]',
        'cobra_args': 'cobra.RangeArgs(1, 2)',
        'position_index': 1,
        'params': ['value := args[0]'],
        'validation': ['if value == "" { return fmt.Errorf("value cannot be empty") }'],
        'extra': 'map[string]interface{}{"value": value}'
    }
}

# Mapping of step commands to their patterns
STEP_COMMAND_MAPPING = {
    'create-step-wait-time': 'single_value',
    'create-step-wait-element': 'single_element',
    'create-step-window': 'position_only',
    'create-step-double-click': 'single_element',
    'create-step-hover': 'single_element',
    'create-step-right-click': 'single_element',
    'create-step-key': 'single_value',
    'create-step-pick': 'element_value',
    'create-step-upload': 'element_value',
    'create-step-scroll-top': 'position_only',
    'create-step-scroll-bottom': 'position_only',
    'create-step-scroll-element': 'single_element',
    'create-step-assert-not-exists': 'single_element',
    'create-step-assert-equals': 'element_value',
    'create-step-assert-checked': 'single_element',
    'create-step-store': 'element_value',
    'create-step-execute-js': 'single_value',
    'create-step-add-cookie': 'element_value',
    'create-step-dismiss-alert': 'position_only',
    'create-step-comment': 'single_value',
}

def generate_command_file(command_name, pattern_name):
    """Generate a clean command file for the given command and pattern."""
    pattern = COMMAND_PATTERNS[pattern_name]
    
    # Convert command name to function name and step type
    func_name = command_name.replace('-', '')
    step_type = command_name.replace('create-step-', '').upper().replace('-', '_')
    
    # Generate parameter assignments
    param_lines = '\n\t\t\t'.join(pattern['params'])
    validation_lines = '\n\t\t\t'.join(pattern['validation'])
    
    # Generate the command file content
    content = f'''package main

import (
\t"fmt"

\t"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
\t"github.com/spf13/cobra"
)

func new{func_name.title()}Cmd() *cobra.Command {{
\tvar checkpointFlag int
\t
\tcmd := &cobra.Command{{
\t\tUse:   "{command_name} {pattern['args']}",
\t\tShort: "Create a {step_type.lower().replace('_', ' ')} step at a specific position in a checkpoint",
\t\tLong: `Create a {step_type.lower().replace('_', ' ')} step at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli {command_name} [example args] 1
  api-cli {command_name} [example args]  # Auto-increment position
  
  # Override checkpoint explicitly
  api-cli {command_name} [example args] 1 --checkpoint 1678318`,
\t\tArgs: {pattern['cobra_args']},
\t\tRunE: func(cmd *cobra.Command, args []string) error {{
\t\t\t{param_lines}
\t\t\t
\t\t\t{validation_lines}
\t\t\t
\t\t\t// Resolve checkpoint and position
\t\t\tctx, err := resolveStepContext(args, checkpointFlag, {pattern['position_index']})
\t\t\tif err != nil {{
\t\t\t\treturn err
\t\t\t}}
\t\t\t
\t\t\t// Create Virtuoso client
\t\t\tclient := virtuoso.NewClient(cfg)
\t\t\t
\t\t\t// Create step using the enhanced client
\t\t\tstepID, err := client.Create{step_type.title().replace('_', '')}Step(ctx.CheckpointID, /* params */, ctx.Position)
\t\t\tif err != nil {{
\t\t\t\treturn fmt.Errorf("failed to create {step_type.lower().replace('_', ' ')} step: %w", err)
\t\t\t}}
\t\t\t
\t\t\t// Save config if position was auto-incremented
\t\t\tsaveStepContext(ctx)
\t\t\t
\t\t\t// Output result
\t\t\toutput := &StepOutput{{
\t\t\t\tStatus:       "success",
\t\t\t\tStepType:     "{step_type}",
\t\t\t\tCheckpointID: ctx.CheckpointID,
\t\t\t\tStepID:       stepID,
\t\t\t\tPosition:     ctx.Position,
\t\t\t\tParsedStep:   fmt.Sprintf("[generated description]"),
\t\t\t\tUsingContext: ctx.UsingContext,
\t\t\t\tAutoPosition: ctx.AutoPosition,
\t\t\t\tExtra:        {pattern['extra']},
\t\t\t}}
\t\t\t
\t\t\treturn outputStepResult(output)
\t\t}},
\t}}
\t
\taddCheckpointFlag(cmd, &checkpointFlag)
\t
\treturn cmd
}}
'''
    
    return content

def main():
    """Generate clean command files for all step commands."""
    base_dir = '/Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd'
    
    print("Generating clean step command files...")
    
    for command_name, pattern_name in STEP_COMMAND_MAPPING.items():
        filename = f"{command_name}.go"
        filepath = os.path.join(base_dir, filename)
        
        print(f"Generating {filename}...")
        
        # Generate the clean command file
        content = generate_command_file(command_name, pattern_name)
        
        # Write the file
        with open(filepath, 'w') as f:
            f.write(content)
    
    print(f"Generated {len(STEP_COMMAND_MAPPING)} clean command files")
    print("Note: You'll need to manually adjust the client method calls and parameters")

if __name__ == '__main__':
    main()