// Package config provides configuration management for the Virtuoso API CLI.
// It handles loading and validating configuration from files and environment variables.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// VirtuosoConfig holds all configuration for the Virtuoso API
type VirtuosoConfig struct {
	API     APIConfig     `mapstructure:"api"`
	Org     OrgConfig     `mapstructure:"organization"`
	Headers HeadersConfig `mapstructure:"headers"`
	Rules   BusinessRules `mapstructure:"business_rules"`
	Output  OutputConfig  `mapstructure:"output"`
	HTTP    HTTPConfig    `mapstructure:"http"`
	Logging LoggingConfig `mapstructure:"logging"`
	Session SessionConfig `mapstructure:"session"`
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

// SessionConfig holds session context information
type SessionConfig struct {
	CurrentProjectID    *int `mapstructure:"current_project_id"`
	CurrentGoalID       *int `mapstructure:"current_goal_id"`
	CurrentSnapshotID   *int `mapstructure:"current_snapshot_id"`
	CurrentJourneyID    *int `mapstructure:"current_journey_id"`
	CurrentCheckpointID *int `mapstructure:"current_checkpoint_id"`
	AutoIncrementPos    bool `mapstructure:"auto_increment_position"`
	NextPosition        int  `mapstructure:"next_position"`
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
	viper.SetDefault("session.auto_increment_position", true)
	viper.SetDefault("session.next_position", 1)

	// Allow environment variables
	viper.SetEnvPrefix("VIRTUOSO")
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; use defaults and warn user
			fmt.Printf("Warning: No config file found. Using defaults. " +
				"To create a config file, run: mkdir -p ./config && touch ./config/virtuoso-config.yaml\n")
		} else {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
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
		"Authorization":          fmt.Sprintf("Bearer %s", c.API.AuthToken),
		"X-Virtuoso-Client-ID":   c.Headers.ClientID,
		"X-Virtuoso-Client-Name": c.Headers.ClientName,
		"Content-Type":           "application/json",
	}
}

// Session management methods

// SetCurrentCheckpoint sets the current checkpoint ID and resets position
func (c *VirtuosoConfig) SetCurrentCheckpoint(checkpointID int) {
	c.Session.CurrentCheckpointID = &checkpointID
	c.Session.NextPosition = 1
}

// GetCurrentCheckpoint returns the current checkpoint ID
func (c *VirtuosoConfig) GetCurrentCheckpoint() *int {
	return c.Session.CurrentCheckpointID
}

// SetCurrentJourney sets the current journey ID and clears checkpoint
func (c *VirtuosoConfig) SetCurrentJourney(journeyID int) {
	c.Session.CurrentJourneyID = &journeyID
	c.Session.CurrentCheckpointID = nil
	c.Session.NextPosition = 1
}

// GetCurrentJourney returns the current journey ID
func (c *VirtuosoConfig) GetCurrentJourney() *int {
	return c.Session.CurrentJourneyID
}

// SetCurrentGoal sets the current goal ID and clears journey/checkpoint
func (c *VirtuosoConfig) SetCurrentGoal(goalID int) {
	c.Session.CurrentGoalID = &goalID
	c.Session.CurrentJourneyID = nil
	c.Session.CurrentCheckpointID = nil
	c.Session.NextPosition = 1
}

// GetCurrentGoal returns the current goal ID
func (c *VirtuosoConfig) GetCurrentGoal() *int {
	return c.Session.CurrentGoalID
}

// SetCurrentSnapshot sets the current snapshot ID
func (c *VirtuosoConfig) SetCurrentSnapshot(snapshotID int) {
	c.Session.CurrentSnapshotID = &snapshotID
}

// GetCurrentSnapshot returns the current snapshot ID
func (c *VirtuosoConfig) GetCurrentSnapshot() *int {
	return c.Session.CurrentSnapshotID
}

// SetCurrentProject sets the current project ID and clears all other context
func (c *VirtuosoConfig) SetCurrentProject(projectID int) {
	c.Session.CurrentProjectID = &projectID
	c.Session.CurrentGoalID = nil
	c.Session.CurrentSnapshotID = nil
	c.Session.CurrentJourneyID = nil
	c.Session.CurrentCheckpointID = nil
	c.Session.NextPosition = 1
}

// GetCurrentProject returns the current project ID
func (c *VirtuosoConfig) GetCurrentProject() *int {
	return c.Session.CurrentProjectID
}

// GetNextPosition returns the next position and increments it if auto-increment is enabled
func (c *VirtuosoConfig) GetNextPosition() int {
	pos := c.Session.NextPosition
	if c.Session.AutoIncrementPos {
		c.Session.NextPosition++
	}
	return pos
}

// SetNextPosition sets the next position
func (c *VirtuosoConfig) SetNextPosition(position int) {
	c.Session.NextPosition = position
}

// ClearSession clears all session context
func (c *VirtuosoConfig) ClearSession() {
	c.Session = SessionConfig{
		AutoIncrementPos: c.Session.AutoIncrementPos,
		NextPosition:     1,
	}
}

// SaveConfig saves the current configuration to the config file
func (c *VirtuosoConfig) SaveConfig() error {
	// Get the config file path
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		// Use default config path if no config file was loaded
		configPath = "./config/virtuoso-config.yaml"
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Update viper with current config values
	viper.Set("session.current_project_id", c.Session.CurrentProjectID)
	viper.Set("session.current_goal_id", c.Session.CurrentGoalID)
	viper.Set("session.current_snapshot_id", c.Session.CurrentSnapshotID)
	viper.Set("session.current_journey_id", c.Session.CurrentJourneyID)
	viper.Set("session.current_checkpoint_id", c.Session.CurrentCheckpointID)
	viper.Set("session.auto_increment_position", c.Session.AutoIncrementPos)
	viper.Set("session.next_position", c.Session.NextPosition)

	// Write config to file
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}
