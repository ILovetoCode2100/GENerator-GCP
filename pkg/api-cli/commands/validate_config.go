package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

func newValidateConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-config",
		Short: "Validate API configuration and connectivity",
		Long: `Validate that the configuration file is properly set up and that the API is accessible.

This command checks:
- Configuration file exists and is valid
- API endpoint is reachable
- Authentication credentials are valid
- Shows current organization details`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Step 1: Check configuration file
			switch cfg.Output.DefaultFormat {
			case "json":
				fmt.Fprintf(os.Stdout, `{"step": "checking_config", "status": "in_progress"}`+"\n")
			default:
				fmt.Println("🔍 Checking configuration file...")
			}

			if cfg == nil {
				return fmt.Errorf("configuration not loaded")
			}

			// Check if config file exists
			configPath := "./config/virtuoso-config.yaml"
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				switch cfg.Output.DefaultFormat {
				case "json":
					fmt.Fprintf(os.Stdout, `{"step": "config_file_check", "status": "warning", "message": "Config file not found, using defaults"}`+"\n")
				default:
					fmt.Printf("⚠️  Config file not found at %s, using defaults\n", configPath)
				}
			}

			// Validate required fields
			if cfg.API.BaseURL == "" {
				return fmt.Errorf("base URL not configured")
			}

			if cfg.API.AuthToken == "" {
				return fmt.Errorf("auth token not configured")
			}

			if cfg.Org.ID == "" {
				return fmt.Errorf("organization ID not configured")
			}

			// Step 2: Test API connectivity
			switch cfg.Output.DefaultFormat {
			case "json":
				fmt.Fprintf(os.Stdout, `{"step": "testing_api", "status": "in_progress"}`+"\n")
			default:
				fmt.Println("🌐 Testing API connectivity...")
			}

			client := client.NewClient(cfg)

			// Step 3: Validate authentication by listing projects
			switch cfg.Output.DefaultFormat {
			case "json":
				fmt.Fprintf(os.Stdout, `{"step": "validating_auth", "status": "in_progress"}`+"\n")
			default:
				fmt.Println("🔐 Validating authentication...")
			}

			// We'll implement a simple health check by trying to list projects
			// This will be replaced with ListProjects when we implement it
			resp, err := client.TestConnection()
			if err != nil {
				return fmt.Errorf("API connection failed: %w", err)
			}

			// Output results
			switch cfg.Output.DefaultFormat {
			case "json":
				result := map[string]interface{}{
					"status": "valid",
					"config": map[string]interface{}{
						"base_url":        cfg.API.BaseURL,
						"organization_id": cfg.Org.ID,
						"headers":         cfg.Headers,
					},
					"api_test": map[string]interface{}{
						"reachable":     true,
						"authenticated": true,
						"response_time": resp,
					},
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				encoder.Encode(result)

			case "yaml":
				fmt.Println("status: valid")
				fmt.Printf("base_url: %s\n", cfg.API.BaseURL)
				fmt.Printf("organization_id: %s\n", cfg.Org.ID)
				fmt.Println("api_reachable: true")
				fmt.Println("authenticated: true")

			case "ai":
				fmt.Println("Configuration validation successful!")
				fmt.Printf("\nConfiguration Details:\n")
				fmt.Printf("- Base URL: %s\n", cfg.API.BaseURL)
				fmt.Printf("- Organization ID: %s\n", cfg.Org.ID)
				fmt.Printf("- Auth Token: %s...%s (hidden)\n", cfg.API.AuthToken[:8], cfg.API.AuthToken[len(cfg.API.AuthToken)-4:])
				fmt.Printf("\nAPI Status:\n")
				fmt.Printf("- API is reachable\n")
				fmt.Printf("- Authentication is valid\n")
				fmt.Printf("- Response time: %s\n", resp)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. List projects: api-cli list-projects\n")
				fmt.Printf("2. Create a project: api-cli create-project \"My Project\"\n")

			default: // human
				fmt.Println("✅ Configuration is valid!")
				fmt.Printf("\n📋 Configuration:\n")
				fmt.Printf("   Base URL: %s\n", cfg.API.BaseURL)
				fmt.Printf("   Organization ID: %s\n", cfg.Org.ID)
				fmt.Printf("   Auth Token: ***...%s\n", cfg.API.AuthToken[len(cfg.API.AuthToken)-4:])
				fmt.Println("\n✅ API connection successful")
				fmt.Println("✅ Authentication valid")
			}

			return nil
		},
	}

	return cmd
}
