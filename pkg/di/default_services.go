//go:generate go tool wire
//go:build wireinject

/*
Default Services DI-package
This package provides common services that are used across all microservices
*/
package di

import (
	"context"

	"github.com/google/wire"
	"go.opentelemetry.io/otel/trace"

	"github.com/shortlink-org/shortlink/pkg/di/pkg/autoMaxPro"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/config"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/profiling"
	"github.com/shortlink-org/shortlink/pkg/di/providers"
	"github.com/shortlink-org/shortlink/pkg/logger"
	"github.com/shortlink-org/shortlink/pkg/observability/metrics"
)

// DefaultServices represents the common services available for all microservices
type DefaultServices struct {
	// Common
	Log        logger.Logger
	Config     *config.Config
	AutoMaxPro autoMaxPro.AutoMaxPro
	Context    context.Context

	// Observability
	Tracer        trace.TracerProvider
	Metrics       *metrics.Monitoring
	PprofEndpoint profiling.PprofEndpoint
}

// DefaultSet provides all common/default services for dependency injection
var DefaultSet = wire.NewSet(
	// Context
	providers.ProvideContext,

	// Common services
	providers.ProvideLogger,
	providers.ProvideConfig,
	providers.ProvideAutoMaxPro,

	// Observability
	providers.ProvideTracer,
	providers.ProvideMetrics,
	providers.ProvidePprofEndpoint,

	// Group all into DefaultServices struct
	NewDefaultServices,
)

// NewDefaultServices creates a new DefaultServices instance with all dependencies
func NewDefaultServices(
	ctx context.Context,
	log logger.Logger,
	config *config.Config,
	autoMaxPro autoMaxPro.AutoMaxPro,
	tracer trace.TracerProvider,
	metrics *metrics.Monitoring,
	pprofEndpoint profiling.PprofEndpoint,
) (*DefaultServices, error) {
	return &DefaultServices{
		Context:       ctx,
		Log:           log,
		Config:        config,
		AutoMaxPro:    autoMaxPro,
		Tracer:        tracer,
		Metrics:       metrics,
		PprofEndpoint: pprofEndpoint,
	}, nil
}

// InitializeDefaultServices wires up all default services
func InitializeDefaultServices() (*DefaultServices, func(), error) {
	panic(wire.Build(DefaultSet))
}