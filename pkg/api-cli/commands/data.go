package commands

import (
	"fmt"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

// dataType represents the type of data operation
type dataType string

const (
	dataStoreElementText dataType = "store-element-text"
	dataStoreLiteral     dataType = "store-literal"
	dataStoreAttribute   dataType = "store-attribute"
	dataCookieCreate     dataType = "cookie-create"
	dataCookieDelete     dataType = "cookie-delete"
	dataCookieClearAll   dataType = "cookie-clear-all"
)

// dataCommandInfo contains metadata about each data operation type
type dataCommandInfo struct {
	stepType    string
	description string
	usage       string
	examples    []string
	argsCount   []int // Valid argument counts (excluding position)
	parseStep   func(args []string, flags map[string]interface{}) string
}

// dataCommands maps data operation types to their metadata
var dataCommands = map[dataType]dataCommandInfo{
	dataStoreElementText: {
		stepType:    "STORE",
		description: "Store element text in a variable",
		usage:       "data store element-text SELECTOR VARIABLE_NAME [POSITION]",
		examples: []string{
			`api-cli data store element-text "Username field" "current_user" 1`,
			`api-cli data store element-text "Order total" "order_amount"  # Auto-increment position`,
		},
		argsCount: []int{2},
		parseStep: func(args []string, flags map[string]interface{}) string {
			return fmt.Sprintf("store text from \"%s\" in $%s", args[0], args[1])
		},
	},
	dataStoreLiteral: {
		stepType:    "STORE",
		description: "Store a literal value in a variable",
		usage:       "data store literal VALUE VARIABLE_NAME [POSITION]",
		examples: []string{
			`api-cli data store literal "test@example.com" "test_email" 1`,
			`api-cli data store literal "2024-01-01" "test_date"  # Auto-increment position`,
		},
		argsCount: []int{2},
		parseStep: func(args []string, flags map[string]interface{}) string {
			return fmt.Sprintf("store \"%s\" in $%s", args[0], args[1])
		},
	},
	dataStoreAttribute: {
		stepType:    "STORE",
		description: "Store element attribute value in a variable",
		usage:       "data store attribute SELECTOR ATTRIBUTE_NAME VARIABLE_NAME [POSITION]",
		examples: []string{
			`api-cli data store attribute "#link" "href" "link_url" 1`,
			`api-cli data store attribute "input[name='email']" "value" "email_value"  # Auto-increment position`,
		},
		argsCount: []int{3},
		parseStep: func(args []string, flags map[string]interface{}) string {
			return fmt.Sprintf("store attribute \"%s\" from \"%s\" in $%s", args[1], args[0], args[2])
		},
	},
	dataCookieCreate: {
		stepType:    "ENVIRONMENT",
		description: "Create a cookie with specified name and value",
		usage:       "data cookie create NAME VALUE [POSITION]",
		examples: []string{
			`api-cli data cookie create "session" "abc123" 1`,
			`api-cli data cookie create "user_id" "12345" --domain ".example.com"`,
			`api-cli data cookie create "auth_token" "xyz789" --secure --http-only`,
		},
		argsCount: []int{2},
		parseStep: func(args []string, flags map[string]interface{}) string {
			parts := []string{fmt.Sprintf("create cookie \"%s\" with value \"%s\"", args[0], args[1])}
			if domain, ok := flags["domain"].(string); ok && domain != "" {
				parts = append(parts, fmt.Sprintf("for domain %s", domain))
			}
			if path, ok := flags["path"].(string); ok && path != "" {
				parts = append(parts, fmt.Sprintf("with path %s", path))
			}
			if secure, ok := flags["secure"].(bool); ok && secure {
				parts = append(parts, "secure")
			}
			if httpOnly, ok := flags["http-only"].(bool); ok && httpOnly {
				parts = append(parts, "http-only")
			}
			return strings.Join(parts, " ")
		},
	},
	dataCookieDelete: {
		stepType:    "ENVIRONMENT",
		description: "Delete a specific cookie",
		usage:       "data cookie delete NAME [POSITION]",
		examples: []string{
			`api-cli data cookie delete "session" 1`,
			`api-cli data cookie delete "auth_token"  # Auto-increment position`,
		},
		argsCount: []int{1},
		parseStep: func(args []string, flags map[string]interface{}) string {
			return fmt.Sprintf("delete cookie \"%s\"", args[0])
		},
	},
	dataCookieClearAll: {
		stepType:    "ENVIRONMENT",
		description: "Clear all cookies",
		usage:       "data cookie clear-all [POSITION]",
		examples: []string{
			`api-cli data cookie clear-all 1`,
			`api-cli data cookie clear-all  # Auto-increment position`,
		},
		argsCount: []int{0},
		parseStep: func(args []string, flags map[string]interface{}) string {
			return "clear all cookies"
		},
	},
}

// newDataCmd creates the consolidated data command with subcommands
func newDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data",
		Short: "Manage data storage and cookies in test steps",
		Long: `Create data management steps including storing values in variables and managing cookies.

This command consolidates all data-related operations into a single command with subcommands.

Available operations:
  - store: Store values in variables
    - element-text: Store element text in variable
    - literal: Store literal value in variable
  - cookie: Manage browser cookies
    - create: Create a new cookie
    - delete: Delete a specific cookie
    - clear-all: Clear all cookies`,
		Example: `  # Store element text in variable
  api-cli data store element-text "Username field" "current_user"

  # Store literal value in variable
  api-cli data store literal "test@example.com" "test_email"

  # Create a cookie
  api-cli data cookie create "session" "abc123"

  # Delete a cookie
  api-cli data cookie delete "session"

  # Clear all cookies
  api-cli data cookie clear-all`,
	}

	// Add store subcommand
	storeCmd := &cobra.Command{
		Use:   "store",
		Short: "Store values in variables",
		Long:  "Store element text or literal values in variables for use in subsequent test steps.",
	}

	// Add store subcommands
	storeCmd.AddCommand(newDataStoreSubCmd("element-text", dataStoreElementText, dataCommands[dataStoreElementText]))
	storeCmd.AddCommand(newDataStoreSubCmd("literal", dataStoreLiteral, dataCommands[dataStoreLiteral]))
	storeCmd.AddCommand(newDataStoreSubCmd("attribute", dataStoreAttribute, dataCommands[dataStoreAttribute]))

	// Add cookie subcommand
	cookieCmd := &cobra.Command{
		Use:   "cookie",
		Short: "Manage browser cookies",
		Long:  "Create, delete, or clear browser cookies in test steps.",
	}

	// Add cookie subcommands
	cookieCmd.AddCommand(newDataCookieCreateCmd())
	cookieCmd.AddCommand(newDataCookieSubCmd("delete", dataCookieDelete, dataCommands[dataCookieDelete]))
	cookieCmd.AddCommand(newDataCookieSubCmd("clear-all", dataCookieClearAll, dataCommands[dataCookieClearAll]))

	cmd.AddCommand(storeCmd)
	cmd.AddCommand(cookieCmd)

	return cmd
}

// newDataStoreSubCmd creates a subcommand for store operations
func newDataStoreSubCmd(name string, dType dataType, info dataCommandInfo) *cobra.Command {
	var checkpointFlag int

	// Extract arguments from usage
	usageParts := strings.Fields(info.usage)
	argsUsage := ""
	if len(usageParts) > 3 {
		argsUsage = strings.Join(usageParts[3:], " ")
	}

	cmd := &cobra.Command{
		Use:   name + " " + argsUsage,
		Short: info.description,
		Long: fmt.Sprintf(`%s

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
%s`, info.description, strings.Join(info.examples, "\n")),
		Args: func(cmd *cobra.Command, args []string) error {
			// Validate argument count
			validCounts := info.argsCount
			for _, count := range validCounts {
				if len(args) == count || len(args) == count+1 {
					return nil
				}
			}

			// Generate expected count message
			expectedCounts := []string{}
			for _, count := range validCounts {
				expectedCounts = append(expectedCounts, fmt.Sprintf("%d", count))
				expectedCounts = append(expectedCounts, fmt.Sprintf("%d", count+1))
			}

			return fmt.Errorf("accepts %s args, received %d", strings.Join(expectedCounts, " or "), len(args))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDataCommand(dType, info, args, checkpointFlag, nil)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}

// newDataCookieCreateCmd creates the cookie create subcommand with additional flags
func newDataCookieCreateCmd() *cobra.Command {
	var checkpointFlag int
	var domain, path string
	var secure, httpOnly bool

	info := dataCommands[dataCookieCreate]

	cmd := &cobra.Command{
		Use:   "create NAME VALUE [POSITION]",
		Short: info.description,
		Long: fmt.Sprintf(`%s

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Cookie options:
  --domain: Cookie domain (e.g., ".example.com")
  --path: Cookie path (default: "/")
  --secure: Set secure flag on cookie
  --http-only: Set httpOnly flag on cookie

Examples:
%s`, info.description, strings.Join(info.examples, "\n")),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 || len(args) > 3 {
				return fmt.Errorf("accepts 2 or 3 args, received %d", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := map[string]interface{}{
				"domain":    domain,
				"path":      path,
				"secure":    secure,
				"http-only": httpOnly,
			}
			return runDataCommand(dataCookieCreate, info, args, checkpointFlag, flags)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)
	cmd.Flags().StringVar(&domain, "domain", "", "Cookie domain")
	cmd.Flags().StringVar(&path, "path", "/", "Cookie path")
	cmd.Flags().BoolVar(&secure, "secure", false, "Set secure flag on cookie")
	cmd.Flags().BoolVar(&httpOnly, "http-only", false, "Set httpOnly flag on cookie")

	return cmd
}

// newDataCookieSubCmd creates a subcommand for cookie operations (delete, clear-all)
func newDataCookieSubCmd(name string, dType dataType, info dataCommandInfo) *cobra.Command {
	var checkpointFlag int

	// Extract arguments from usage
	usageParts := strings.Fields(info.usage)
	argsUsage := ""
	if len(usageParts) > 3 {
		argsUsage = strings.Join(usageParts[3:], " ")
	}

	cmd := &cobra.Command{
		Use:   name + " " + argsUsage,
		Short: info.description,
		Long: fmt.Sprintf(`%s

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
%s`, info.description, strings.Join(info.examples, "\n")),
		Args: func(cmd *cobra.Command, args []string) error {
			// Validate argument count
			validCounts := info.argsCount
			for _, count := range validCounts {
				if len(args) == count || len(args) == count+1 {
					return nil
				}
			}

			// Generate expected count message
			expectedCounts := []string{}
			for _, count := range validCounts {
				expectedCounts = append(expectedCounts, fmt.Sprintf("%d", count))
				expectedCounts = append(expectedCounts, fmt.Sprintf("%d", count+1))
			}

			return fmt.Errorf("accepts %s args, received %d", strings.Join(expectedCounts, " or "), len(args))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDataCommand(dType, info, args, checkpointFlag, nil)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}

// runDataCommand executes the data command logic
func runDataCommand(dType dataType, info dataCommandInfo, args []string, checkpointFlag int, flags map[string]interface{}) error {
	// Validate arguments based on data type
	if err := validateDataArgs(dType, args); err != nil {
		return err
	}

	// Resolve checkpoint and position
	// The position index depends on how many required arguments we have
	positionIndex := -1
	if len(info.argsCount) > 0 {
		positionIndex = info.argsCount[0] // Position comes after required args
	}
	ctx, err := resolveStepContext(args, checkpointFlag, positionIndex)
	if err != nil {
		return err
	}

	// Create Virtuoso client
	apiClient := client.NewClient(cfg)

	// Call the appropriate API method based on data type
	stepID, err := callDataAPI(apiClient, dType, ctx, args, flags)
	if err != nil {
		return fmt.Errorf("failed to create %s step: %w", info.stepType, err)
	}

	// Save config if position was auto-incremented
	saveStepContext(ctx)

	// Build extra data for output
	extra := buildDataExtraData(dType, args, flags)

	// Output result
	output := &StepOutput{
		Status:       "success",
		StepType:     info.stepType,
		CheckpointID: ctx.CheckpointID,
		StepID:       stepID,
		Position:     ctx.Position,
		ParsedStep:   info.parseStep(args, flags),
		UsingContext: ctx.UsingContext,
		AutoPosition: ctx.AutoPosition,
		Extra:        extra,
	}

	return outputStepResult(output)
}

// validateDataArgs validates arguments for a specific data operation type
func validateDataArgs(dType dataType, args []string) error {
	switch dType {
	case dataStoreElementText:
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("selector cannot be empty")
		}
		if len(args) < 2 || args[1] == "" {
			return fmt.Errorf("variable name cannot be empty")
		}
		if !isValidVariableName(args[1]) {
			return fmt.Errorf("invalid variable name: %s (must contain only letters, numbers, and underscores)", args[1])
		}
	case dataStoreLiteral:
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("value cannot be empty")
		}
		if len(args) < 2 || args[1] == "" {
			return fmt.Errorf("variable name cannot be empty")
		}
		if !isValidVariableName(args[1]) {
			return fmt.Errorf("invalid variable name: %s (must contain only letters, numbers, and underscores)", args[1])
		}
	case dataStoreAttribute:
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("selector cannot be empty")
		}
		if len(args) < 2 || args[1] == "" {
			return fmt.Errorf("attribute name cannot be empty")
		}
		if len(args) < 3 || args[2] == "" {
			return fmt.Errorf("variable name cannot be empty")
		}
		if !isValidVariableName(args[2]) {
			return fmt.Errorf("invalid variable name: %s (must contain only letters, numbers, and underscores)", args[2])
		}
	case dataCookieCreate:
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("cookie name cannot be empty")
		}
		if len(args) < 2 || args[1] == "" {
			return fmt.Errorf("cookie value cannot be empty")
		}
	case dataCookieDelete:
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("cookie name cannot be empty")
		}
	case dataCookieClearAll:
		// No arguments required
	}
	return nil
}

// isValidVariableName checks if a variable name is valid
func isValidVariableName(name string) bool {
	if len(name) == 0 {
		return false
	}
	for i, ch := range name {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9' && i > 0) ||
			ch == '_') {
			return false
		}
	}
	return true
}

// callDataAPI calls the appropriate client API method for the data operation type
func callDataAPI(apiClient *client.Client, dType dataType, ctx *StepContext, args []string, flags map[string]interface{}) (int, error) {
	switch dType {
	case dataStoreElementText:
		return apiClient.CreateStepStoreElementText(ctx.CheckpointID, args[0], args[1], ctx.Position)
	case dataStoreLiteral:
		return apiClient.CreateStepStoreLiteralValue(ctx.CheckpointID, args[0], args[1], ctx.Position)
	case dataStoreAttribute:
		return apiClient.CreateStepStoreAttribute(ctx.CheckpointID, args[0], args[1], args[2], ctx.Position)
	case dataCookieCreate:
		if flags != nil && (flags["domain"] != nil || flags["path"] != nil || flags["secure"] != nil || flags["http-only"] != nil) {
			options := map[string]interface{}{
				"domain":   flags["domain"],
				"path":     flags["path"],
				"secure":   flags["secure"],
				"httpOnly": flags["http-only"],
			}
			return apiClient.CreateStepCookieCreateWithOptions(ctx.CheckpointID, args[0], args[1], options, ctx.Position)
		}
		return apiClient.CreateStepCookieCreate(ctx.CheckpointID, args[0], args[1], ctx.Position)
	case dataCookieDelete:
		return apiClient.CreateStepDeleteCookie(ctx.CheckpointID, args[0], ctx.Position)
	case dataCookieClearAll:
		return apiClient.CreateStepCookieWipeAll(ctx.CheckpointID, ctx.Position)
	default:
		return 0, fmt.Errorf("unsupported data operation type: %s", dType)
	}
}

// buildDataExtraData builds the extra data map for output based on data operation type
func buildDataExtraData(dType dataType, args []string, flags map[string]interface{}) map[string]interface{} {
	extra := make(map[string]interface{})

	switch dType {
	case dataStoreElementText:
		extra["selector"] = args[0]
		extra["variable_name"] = args[1]
	case dataStoreLiteral:
		extra["value"] = args[0]
		extra["variable_name"] = args[1]
	case dataStoreAttribute:
		extra["selector"] = args[0]
		extra["attribute"] = args[1]
		extra["variable_name"] = args[2]
	case dataCookieCreate:
		extra["cookie_name"] = args[0]
		extra["cookie_value"] = args[1]
		if flags != nil {
			if domain, ok := flags["domain"].(string); ok && domain != "" {
				extra["domain"] = domain
			}
			if path, ok := flags["path"].(string); ok && path != "" {
				extra["path"] = path
			}
			if secure, ok := flags["secure"].(bool); ok && secure {
				extra["secure"] = secure
			}
			if httpOnly, ok := flags["http-only"].(bool); ok && httpOnly {
				extra["http_only"] = httpOnly
			}
		}
	case dataCookieDelete:
		extra["cookie_name"] = args[0]
	case dataCookieClearAll:
		// No extra data needed
	}

	return extra
}
