package commands

import (
	"github.com/marklovelady/api-cli-generator/pkg/api-cli/config"
)

// Global config variable for the commands package
var cfg *config.VirtuosoConfig

// SetConfig sets the global config for the commands package
func SetConfig(c *config.VirtuosoConfig) {
	cfg = c
}
