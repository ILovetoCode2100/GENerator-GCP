package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

// DataCommand implements the data command group using BaseCommand pattern
type DataCommand struct {
	*BaseCommand
	operation string
	subtype   string
}

// dataConfig contains configuration for each data operation
type dataConfig struct {
	stepType     string
	description  string
	usage        string
	examples     []string
	requiredArgs int
	buildMeta    func(args []string, flags map[string]interface{}) map[string]interface{}
	extraFlags   []flagConfig
}

// flagConfig defines additional flags for certain data operations
type flagConfig struct {
	name        string
	flagType    string // "string", "bool"
	defaultVal  interface{}
	description string
}

// dataConfigs maps data operations to their configurations
var dataConfigs = map[string]dataConfig{
	"store.element-text": {
		stepType:    "STORE",
		description: "Store element text in a variable",
		usage:       "data store element-text [checkpoint-id] <selector> <variable-name> [position]",
		examples: []string{
			`api-cli data store element-text cp_12345 "Username field" "current_user" 1`,
			`api-cli data store element-text "Order total" "order_amount"  # Uses session context`,
		},
		requiredArgs: 2,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"operation": "ELEMENT_TEXT",
				"selector":  args[0],
				"variable":  args[1],
			}
		},
	},
	"store.literal": {
		stepType:    "STORE",
		description: "Store a literal value in a variable",
		usage:       "data store literal [checkpoint-id] <value> <variable-name> [position]",
		examples: []string{
			`api-cli data store literal cp_12345 "test@example.com" "test_email" 1`,
			`api-cli data store literal "2024-01-01" "test_date"  # Uses session context`,
		},
		requiredArgs: 2,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"operation": "LITERAL",
				"value":     args[0],
				"variable":  args[1],
			}
		},
	},
	"store.attribute": {
		stepType:    "STORE",
		description: "Store element attribute value in a variable",
		usage:       "data store attribute [checkpoint-id] <selector> <attribute-name> <variable-name> [position]",
		examples: []string{
			`api-cli data store attribute cp_12345 "#link" "href" "link_url" 1`,
			`api-cli data store attribute "input[name='email']" "value" "email_value"  # Uses session context`,
		},
		requiredArgs: 3,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"operation": "ATTRIBUTE",
				"selector":  args[0],
				"attribute": args[1],
				"variable":  args[2],
			}
		},
	},
	"cookie.create": {
		stepType:    "ENVIRONMENT",
		description: "Create a cookie with specified name and value",
		usage:       "data cookie create [checkpoint-id] <name> <value> [position] [--domain <domain>] [--path <path>] [--secure] [--http-only]",
		examples: []string{
			`api-cli data cookie create cp_12345 "session" "abc123" 1`,
			`api-cli data cookie create "user_id" "12345" --domain ".example.com"`,
			`api-cli data cookie create "auth_token" "xyz789" --secure --http-only`,
		},
		requiredArgs: 2,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			meta := map[string]interface{}{
				"operation": "COOKIE_CREATE",
				"name":      args[0],
				"value":     args[1],
			}
			// Add optional cookie attributes
			if domain, ok := flags["domain"].(string); ok && domain != "" {
				meta["domain"] = domain
			}
			if path, ok := flags["path"].(string); ok && path != "" {
				meta["path"] = path
			}
			if secure, ok := flags["secure"].(bool); ok && secure {
				meta["secure"] = true
			}
			if httpOnly, ok := flags["http-only"].(bool); ok && httpOnly {
				meta["httpOnly"] = true
			}
			return meta
		},
		extraFlags: []flagConfig{
			{name: "domain", flagType: "string", defaultVal: "", description: "Cookie domain"},
			{name: "path", flagType: "string", defaultVal: "/", description: "Cookie path"},
			{name: "secure", flagType: "bool", defaultVal: false, description: "Set secure flag on cookie"},
			{name: "http-only", flagType: "bool", defaultVal: false, description: "Set httpOnly flag on cookie"},
		},
	},
	"cookie.delete": {
		stepType:    "ENVIRONMENT",
		description: "Delete a specific cookie",
		usage:       "data cookie delete [checkpoint-id] <name> [position]",
		examples: []string{
			`api-cli data cookie delete cp_12345 "session" 1`,
			`api-cli data cookie delete "auth_token"  # Uses session context`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"operation": "COOKIE_DELETE",
				"name":      args[0],
			}
		},
	},
	"cookie.clear-all": {
		stepType:    "ENVIRONMENT",
		description: "Clear all cookies",
		usage:       "data cookie clear-all [checkpoint-id] [position]",
		examples: []string{
			`api-cli data cookie clear-all cp_12345 1`,
			`api-cli data cookie clear-all  # Uses session context`,
		},
		requiredArgs: 0,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"operation": "COOKIE_CLEAR_ALL",
			}
		},
	},
}

// newStepDataCmd creates the new data command using BaseCommand pattern
func newStepDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "step-data",
		Short: "Manage data storage and cookies in test steps",
		Long: `Create data management steps including storing values in variables and managing cookies.

This command uses the standardized positional argument pattern:
- Optional checkpoint ID as first argument (falls back to session context)
- Required data operation arguments
- Optional position as last argument (auto-increments if not specified)

Available operations:
  - store: Store values in variables
    - element-text: Store element text in variable
    - literal: Store literal value in variable
    - attribute: Store element attribute in variable
  - cookie: Manage browser cookies
    - create: Create a new cookie
    - delete: Delete a specific cookie
    - clear-all: Clear all cookies`,
		Example: `  # Store element text in variable (with explicit checkpoint)
  api-cli data store element-text cp_12345 "Username field" "current_user" 1

  # Store element text in variable (using session context)
  api-cli data store element-text "Username field" "current_user"

  # Store literal value in variable
  api-cli data store literal "test@example.com" "test_email"

  # Create cookie with options
  api-cli data cookie create "session" "abc123" --domain ".example.com" --secure`,
	}

	// Add store subcommand
	storeCmd := &cobra.Command{
		Use:   "store",
		Short: "Store values in variables",
		Long:  "Store element text, literal values, or attributes in variables for later use",
	}

	// Add store subcommands
	storeCmd.AddCommand(newDataV2SubCmd("element-text", "store.element-text", dataConfigs["store.element-text"]))
	storeCmd.AddCommand(newDataV2SubCmd("literal", "store.literal", dataConfigs["store.literal"]))
	storeCmd.AddCommand(newDataV2SubCmd("attribute", "store.attribute", dataConfigs["store.attribute"]))

	// Add cookie subcommand
	cookieCmd := &cobra.Command{
		Use:   "cookie",
		Short: "Manage browser cookies",
		Long:  "Create, delete, or clear browser cookies",
	}

	// Add cookie subcommands
	cookieCmd.AddCommand(newDataV2SubCmd("create", "cookie.create", dataConfigs["cookie.create"]))
	cookieCmd.AddCommand(newDataV2SubCmd("delete", "cookie.delete", dataConfigs["cookie.delete"]))
	cookieCmd.AddCommand(newDataV2SubCmd("clear-all", "cookie.clear-all", dataConfigs["cookie.clear-all"]))

	cmd.AddCommand(storeCmd)
	cmd.AddCommand(cookieCmd)

	return cmd
}

// newDataV2SubCmd creates a subcommand for a specific data operation
func newDataV2SubCmd(name string, operationKey string, config dataConfig) *cobra.Command {
	// Store flag values
	flagValues := make(map[string]interface{})

	cmd := &cobra.Command{
		Use:   name + " " + extractDataUsageArgs(config.usage),
		Short: config.description,
		Long: fmt.Sprintf(`%s

%s

Examples:
%s`, config.description, config.usage, strings.Join(config.examples, "\n")),
		RunE: func(cmd *cobra.Command, args []string) error {
			dc := &DataCommand{
				BaseCommand: NewBaseCommand(),
				operation:   operationKey,
			}
			return dc.Execute(cmd, args, config, flagValues)
		},
	}

	// Add extra flags if defined
	for _, flag := range config.extraFlags {
		switch flag.flagType {
		case "string":
			var val string
			cmd.Flags().StringVar(&val, flag.name, flag.defaultVal.(string), flag.description)
			flagValues[flag.name] = &val
		case "bool":
			var val bool
			cmd.Flags().BoolVar(&val, flag.name, flag.defaultVal.(bool), flag.description)
			flagValues[flag.name] = &val
		}
	}

	return cmd
}

// extractDataUsageArgs extracts the arguments portion from the usage string
func extractDataUsageArgs(usage string) string {
	parts := strings.Fields(usage)
	// Find where the actual args start (after "data store element-text" or similar)
	for i, part := range parts {
		if strings.HasPrefix(part, "[checkpoint-id]") {
			return strings.Join(parts[i:], " ")
		}
	}
	return ""
}

// Execute runs the data command
func (dc *DataCommand) Execute(cmd *cobra.Command, args []string, config dataConfig, flagValues map[string]interface{}) error {
	// Initialize base command
	if err := dc.Init(cmd); err != nil {
		return err
	}

	// Resolve checkpoint and position
	remainingArgs, err := dc.ResolveCheckpointAndPosition(args, config.requiredArgs)
	if err != nil {
		return fmt.Errorf("failed to resolve arguments: %w", err)
	}

	// Validate we have the required number of arguments
	if len(remainingArgs) != config.requiredArgs {
		return fmt.Errorf("expected %d arguments, got %d", config.requiredArgs, len(remainingArgs))
	}

	// Resolve flag values
	resolvedFlags := make(map[string]interface{})
	for name, ptr := range flagValues {
		switch v := ptr.(type) {
		case *string:
			resolvedFlags[name] = *v
		case *bool:
			resolvedFlags[name] = *v
		}
	}

	// Build request metadata
	meta := config.buildMeta(remainingArgs, resolvedFlags)

	// Create the step
	stepResult, err := dc.createDataStep(config.stepType, meta)
	if err != nil {
		return fmt.Errorf("failed to create %s step: %w", config.stepType, err)
	}

	// Format and output the result
	output, err := dc.FormatOutput(stepResult, dc.OutputFormat)
	if err != nil {
		return err
	}

	fmt.Print(output)
	return nil
}

// createDataStep creates a data operation step via the API
func (dc *DataCommand) createDataStep(stepType string, meta map[string]interface{}) (*StepResult, error) {
	// Convert checkpoint ID from string to int
	checkpointID, err := strconv.Atoi(dc.CheckpointID)
	if err != nil {
		return nil, fmt.Errorf("invalid checkpoint ID: %s", dc.CheckpointID)
	}

	// Create context with timeout
	ctx, cancel := dc.CommandContext()
	defer cancel()

	var stepID int
	operation := meta["operation"].(string)

	// Route to appropriate client method based on operation
	switch operation {
	case "ELEMENT_TEXT":
		stepID, err = dc.Client.CreateStepStoreElementTextWithContext(ctx, checkpointID,
			meta["selector"].(string),
			meta["variable"].(string),
			dc.Position)
	case "LITERAL":
		stepID, err = dc.Client.CreateStepStoreLiteralValueWithContext(ctx, checkpointID,
			meta["value"].(string),
			meta["variable"].(string),
			dc.Position)
	case "ATTRIBUTE":
		stepID, err = dc.Client.CreateStepStoreAttributeWithContext(ctx, checkpointID,
			meta["selector"].(string),
			meta["attribute"].(string),
			meta["variable"].(string),
			dc.Position)
	case "COOKIE_CREATE":
		// Build cookie options
		options := make(map[string]interface{})
		if domain, ok := meta["domain"]; ok {
			options["domain"] = domain
		}
		if path, ok := meta["path"]; ok {
			options["path"] = path
		}
		if secure, ok := meta["secure"]; ok {
			options["secure"] = secure
		}
		if httpOnly, ok := meta["httpOnly"]; ok {
			options["httpOnly"] = httpOnly
		}
		// Use the method with options if we have any
		if len(options) > 0 {
			stepID, err = dc.Client.CreateStepCookieCreateWithOptionsWithContext(ctx, checkpointID,
				meta["name"].(string),
				meta["value"].(string),
				options,
				dc.Position)
		} else {
			stepID, err = dc.Client.CreateStepCookieCreateWithContext(ctx, checkpointID,
				meta["name"].(string),
				meta["value"].(string),
				dc.Position)
		}
	case "COOKIE_DELETE":
		stepID, err = dc.Client.CreateStepCookieDeleteWithContext(ctx, checkpointID,
			meta["name"].(string),
			dc.Position)
	case "COOKIE_CLEAR_ALL":
		stepID, err = dc.Client.CreateStepCookieClearAllWithContext(ctx, checkpointID, dc.Position)
	default:
		return nil, fmt.Errorf("unknown data operation: %s", operation)
	}

	if err != nil {
		// Handle different error types
		var apiErr *client.APIError
		var clientErr *client.ClientError

		if errors.As(err, &apiErr) {
			// API errors - provide user-friendly messages
			switch apiErr.Status {
			case 400:
				return nil, fmt.Errorf("invalid request: %s", apiErr.Message)
			case 401:
				return nil, fmt.Errorf("authentication failed: please check your API token")
			case 403:
				return nil, fmt.Errorf("access denied: you don't have permission to perform this operation")
			case 404:
				return nil, fmt.Errorf("checkpoint not found: %s", dc.CheckpointID)
			case 429:
				if apiErr.RetryAfter > 0 {
					return nil, fmt.Errorf("rate limit exceeded: please wait %d seconds before retrying", apiErr.RetryAfter)
				}
				return nil, fmt.Errorf("rate limit exceeded: please wait before retrying")
			case 500, 502, 503, 504:
				return nil, fmt.Errorf("server error: %s (please try again later)", apiErr.Message)
			default:
				return nil, fmt.Errorf("API error (%d): %s", apiErr.Status, apiErr.Message)
			}
		} else if errors.As(err, &clientErr) {
			// Client errors - provide context-specific messages
			switch clientErr.Kind {
			case client.KindTimeout:
				return nil, fmt.Errorf("request timed out: the operation took too long to complete")
			case client.KindContextCanceled:
				return nil, fmt.Errorf("operation canceled")
			case client.KindConnectionFailed:
				return nil, fmt.Errorf("connection failed: please check your network and try again")
			default:
				return nil, fmt.Errorf("client error: %s", clientErr.Message)
			}
		}

		// Generic error
		return nil, fmt.Errorf("failed to create %s step: %w", stepType, err)
	}

	// Build the result
	result := &StepResult{
		ID:           fmt.Sprintf("%d", stepID),
		CheckpointID: dc.CheckpointID,
		Type:         stepType,
		Position:     dc.Position,
		Description:  dc.buildDescription(operation, meta),
		Selector:     dc.extractSelector(meta),
		Meta:         meta,
	}

	// Save session state if position was auto-incremented
	if dc.Position == -1 && cfg.Session.AutoIncrementPos {
		if err := cfg.SaveConfig(); err != nil {
			// Don't fail the command, just warn
			// Note: In production, this warning would be sent to stderr
		}
	}

	return result, nil
}

// buildDescription creates a human-readable description for the step
func (dc *DataCommand) buildDescription(operation string, meta map[string]interface{}) string {
	switch operation {
	case "ELEMENT_TEXT":
		return fmt.Sprintf("store text from \"%s\" in $%s", meta["selector"], meta["variable"])
	case "LITERAL":
		return fmt.Sprintf("store \"%s\" in $%s", meta["value"], meta["variable"])
	case "ATTRIBUTE":
		return fmt.Sprintf("store attribute \"%s\" from \"%s\" in $%s", meta["attribute"], meta["selector"], meta["variable"])
	case "COOKIE_CREATE":
		desc := fmt.Sprintf("create cookie \"%s\" with value \"%s\"", meta["name"], meta["value"])
		if domain, ok := meta["domain"]; ok {
			desc += fmt.Sprintf(" for domain %s", domain)
		}
		return desc
	case "COOKIE_DELETE":
		return fmt.Sprintf("delete cookie \"%s\"", meta["name"])
	case "COOKIE_CLEAR_ALL":
		return "clear all cookies"
	default:
		return operation
	}
}

// extractSelector extracts the selector from metadata if present
func (dc *DataCommand) extractSelector(meta map[string]interface{}) string {
	if selector, ok := meta["selector"].(string); ok {
		return selector
	}
	return ""
}
