package config

import (
	"fmt"
	"os"
	"github.com/spf13/viper"
)

// VirtuosoConfig holds all configuration for the Virtuoso API
type VirtuosoConfig struct {
	API      APIConfig      `mapstructure:"api"`
	Org      OrgConfig      `mapstructure:"organization"`
	Headers  HeadersConfig  `mapstructure:"headers"`
	Rules    BusinessRules  `mapstructure:"business_rules"`
	Output   OutputConfig   `mapstructure:"output"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// APIConfig holds API connection details
type APIConfig struct {
	BaseURL   string `mapstructure:"base_url"`
	AuthToken string `mapstructure:"auth_token"`
}

// OrgConfig holds organization details
type OrgConfig struct {
	ID string `mapstructure:"id"`
}

// HeadersConfig holds custom headers
type HeadersConfig struct {
	ClientID   string `mapstructure:"X-Virtuoso-Client-ID"`
	ClientName string `mapstructure:"X-Virtuoso-Client-Name"`
}

// BusinessRules holds Virtuoso-specific business rules
type BusinessRules struct {
	InitialCheckpointName string `mapstructure:"initial_checkpoint_name"`
	AutoAttachCheckpoints bool   `mapstructure:"auto_attach_checkpoints"`
	CreateInitialJourney  bool   `mapstructure:"create_initial_journey"`
}

// OutputConfig holds output formatting options
type OutputConfig struct {
	DefaultFormat string `mapstructure:"default_format"`
	Verbose       bool   `mapstructure:"verbose"`
}

// HTTPConfig holds HTTP client settings
type HTTPConfig struct {
	Timeout   int `mapstructure:"timeout"`
	Retries   int `mapstructure:"retries"`
	RetryWait int `mapstructure:"retry_wait"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

// LoadConfig loads configuration from file and environment
func LoadConfig() (*VirtuosoConfig, error) {
	viper.SetConfigName("virtuoso-config")
	viper.SetConfigType("yaml")
	
	// Add config paths
	viper.AddConfigPath("./config")
	viper.AddConfigPath("$HOME/.api-cli")
	viper.AddConfigPath(".")
	
	// Set defaults
	viper.SetDefault("api.base_url", "https://api-app2.virtuoso.qa/api")
	viper.SetDefault("organization.id", "2242")
	viper.SetDefault("headers.X-Virtuoso-Client-ID", "api-cli-generator")
	viper.SetDefault("headers.X-Virtuoso-Client-Name", "api-cli-generator")
	viper.SetDefault("business_rules.initial_checkpoint_name", "INITIAL_CHECKPOINT")
	viper.SetDefault("business_rules.auto_attach_checkpoints", true)
	viper.SetDefault("business_rules.create_initial_journey", true)
	viper.SetDefault("output.default_format", "human")
	viper.SetDefault("http.timeout", 30)
	viper.SetDefault("http.retries", 3)
	
	// Allow environment variables
	viper.SetEnvPrefix("VIRTUOSO")
	viper.AutomaticEnv()
	
	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
		// Config file not found; use defaults
	}
	
	var config VirtuosoConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}
	
	// Override with environment variables if set
	if token := os.Getenv("VIRTUOSO_API_TOKEN"); token != "" {
		config.API.AuthToken = token
	}
	
	return &config, nil
}

// GetHeaders returns all headers for API requests
func (c *VirtuosoConfig) GetHeaders() map[string]string {
	return map[string]string{
		"Authorization":         fmt.Sprintf("Bearer %s", c.API.AuthToken),
		"X-Virtuoso-Client-ID":  c.Headers.ClientID,
		"X-Virtuoso-Client-Name": c.Headers.ClientName,
		"Content-Type":          "application/json",
	}
}
