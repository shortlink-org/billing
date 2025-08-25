package providers

import (
	"github.com/shortlink-org/shortlink/pkg/di/pkg/config"
)

// ProvideConfig provides application configuration
func ProvideConfig() (*config.Config, error) {
	// Load configuration from environment or config files
	return config.New()
}