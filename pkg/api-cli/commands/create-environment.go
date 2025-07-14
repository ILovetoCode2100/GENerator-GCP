package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/marklovelady/api-cli-generator/pkg/api-cli/config"
	"github.com/spf13/cobra"
)

// EnvironmentOutput represents the output structure for environment operations
type EnvironmentOutput struct {
	Status        string                 `json:"status"`
	EnvironmentID string                 `json:"environment_id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	Variables     map[string]interface{} `json:"variables"`
	VariableCount int                    `json:"variable_count"`
	CreatedAt     time.Time              `json:"created_at"`
	NextSteps     []string               `json:"next_steps,omitempty"`
}

func newCreateEnvironmentCmd() *cobra.Command {
	var nameFlag string
	var descriptionFlag string
	var variablesFlag string
	var variablesFileFlag string
	var copyFromFlag string

	cmd := &cobra.Command{
		Use:   "create-environment",
		Short: "Create a new test environment with variables",
		Long: `Create a new test environment with configuration variables for testing.

The command creates test environments with support for:
- Environment variables and configuration settings
- JSON file import for bulk variable configuration
- Copying variables from existing environments
- Secure handling of sensitive variables
- Environment-specific test configurations

Examples:
  # Create basic environment
  api-cli create-environment --name "Production" --description "Production environment"

  # Create environment with variables
  api-cli create-environment --name "Staging" --variables "BASE_URL=https://staging.example.com,API_KEY=secret123"

  # Create environment from JSON file
  api-cli create-environment --name "Development" --variables-file dev-config.json

  # Copy environment from existing one
  api-cli create-environment --name "Production-Copy" --copy-from env_456`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			if nameFlag == "" {
				return fmt.Errorf("environment name is required")
			}

			client := client.NewClient(cfg)

			// Parse variables from different sources
			variables := make(map[string]interface{})

			// Parse variables from flag
			if variablesFlag != "" {
				vars, err := parseVariablesString(variablesFlag)
				if err != nil {
					return fmt.Errorf("failed to parse variables: %w", err)
				}
				for k, v := range vars {
					variables[k] = v
				}
			}

			// Parse variables from file
			if variablesFileFlag != "" {
				vars, err := parseVariablesFile(variablesFileFlag)
				if err != nil {
					return fmt.Errorf("failed to parse variables file: %w", err)
				}
				for k, v := range vars {
					variables[k] = v
				}
			}

			// TODO: Handle copy-from functionality
			if copyFromFlag != "" {
				return fmt.Errorf("copy-from functionality not yet implemented")
			}

			// Create environment
			environment, err := client.CreateEnvironment(nameFlag, descriptionFlag, variables)
			if err != nil {
				return fmt.Errorf("failed to create environment: %w", err)
			}

			// Prepare output
			output := &EnvironmentOutput{
				Status:        "success",
				EnvironmentID: environment.ID,
				Name:          environment.Name,
				Description:   environment.Description,
				Variables:     environment.Variables,
				VariableCount: len(environment.Variables),
				CreatedAt:     environment.CreatedAt,
				NextSteps: []string{
					"Use environment ID '" + environment.ID + "' in goal configurations",
					"Reference variables in test steps using ${VARIABLE_NAME} syntax",
					"Use 'api-cli update-environment " + environment.ID + "' to modify variables",
				},
			}

			return outputEnvironmentResult(output, cfg.Output.DefaultFormat)
		},
	}

	cmd.Flags().StringVar(&nameFlag, "name", "", "Name of the environment (required)")
	cmd.Flags().StringVar(&descriptionFlag, "description", "", "Description of the environment")
	cmd.Flags().StringVar(&variablesFlag, "variables", "", "Comma-separated key=value pairs")
	cmd.Flags().StringVar(&variablesFileFlag, "variables-file", "", "JSON file containing variables")
	cmd.Flags().StringVar(&copyFromFlag, "copy-from", "", "Environment ID to copy variables from")

	return cmd
}

// parseVariablesString parses a comma-separated string of key=value pairs
func parseVariablesString(varsStr string) (map[string]interface{}, error) {
	variables := make(map[string]interface{})

	if varsStr == "" {
		return variables, nil
	}

	pairs := strings.Split(varsStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid variable format: %s (expected key=value)", pair)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("empty variable key in: %s", pair)
		}

		variables[key] = value
	}

	return variables, nil
}

// parseVariablesFile parses a JSON file containing variables
func parseVariablesFile(filePath string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open variables file: %w", err)
	}
	defer file.Close()

	var variables map[string]interface{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&variables)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON variables file: %w", err)
	}

	return variables, nil
}

// outputEnvironmentResult formats and outputs the environment result
func outputEnvironmentResult(output *EnvironmentOutput, format string) error {
	switch format {
	case "json":
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(jsonData))

	case "yaml":
		// Convert to YAML-friendly format
		yamlData := map[string]interface{}{
			"status":         output.Status,
			"environment_id": output.EnvironmentID,
			"name":           output.Name,
			"description":    output.Description,
			"variables":      output.Variables,
			"variable_count": output.VariableCount,
			"created_at":     output.CreatedAt.Format(time.RFC3339),
			"next_steps":     output.NextSteps,
		}

		jsonData, err := json.MarshalIndent(yamlData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		fmt.Println(string(jsonData))

	case "ai":
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal AI format: %w", err)
		}
		fmt.Println(string(jsonData))

	default: // human
		fmt.Printf("✅ Environment created successfully\n")
		fmt.Printf("🆔 Environment ID: %s\n", output.EnvironmentID)
		fmt.Printf("📋 Name: %s\n", output.Name)

		if output.Description != "" {
			fmt.Printf("📝 Description: %s\n", output.Description)
		}

		fmt.Printf("🔢 Variables: %d\n", output.VariableCount)

		if len(output.Variables) > 0 {
			fmt.Printf("\n🔐 Environment Variables:\n")
			for key, value := range output.Variables {
				// Mask sensitive values
				displayValue := fmt.Sprintf("%v", value)
				if isSensitiveKey(key) {
					displayValue = maskValue(displayValue)
				}
				fmt.Printf("  • %s: %s\n", key, displayValue)
			}
		}

		fmt.Printf("\n⏰ Created: %s\n", output.CreatedAt.Format("2006-01-02 15:04:05"))

		if len(output.NextSteps) > 0 {
			fmt.Printf("\n💡 Next Steps:\n")
			for _, step := range output.NextSteps {
				fmt.Printf("  • %s\n", step)
			}
		}
	}

	return nil
}

// isSensitiveKey checks if a key contains sensitive information
func isSensitiveKey(key string) bool {
	key = strings.ToLower(key)
	sensitiveKeys := []string{
		"password", "token", "key", "secret", "auth", "credential",
		"api_key", "access_token", "client_secret", "private_key",
	}

	for _, sensitive := range sensitiveKeys {
		if strings.Contains(key, sensitive) {
			return true
		}
	}

	return false
}

// maskValue masks sensitive values for display
func maskValue(value string) string {
	if len(value) <= 4 {
		return "***"
	}

	return value[:2] + "***" + value[len(value)-2:]
}
