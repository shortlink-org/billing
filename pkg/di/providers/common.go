package providers

import (
	"context"

	"github.com/shortlink-org/shortlink/pkg/di/pkg/autoMaxPro"
)

// ProvideContext provides a background context
func ProvideContext() context.Context {
	return context.Background()
}

// ProvideAutoMaxPro provides automatic max process configuration
func ProvideAutoMaxPro() autoMaxPro.AutoMaxPro {
	return autoMaxPro.New()
}