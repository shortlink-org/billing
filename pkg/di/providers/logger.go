package providers

import (
	"github.com/shortlink-org/shortlink/pkg/logger"
)

// ProvideLogger provides a configured logger instance
func ProvideLogger() logger.Logger {
	// Initialize logger with default configuration
	log, _ := logger.NewLogger(logger.INFO, logger.JSON)
	return log
}