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
	ProvideContext,

	// Common services
	ProvideLogger,
	ProvideConfig,
	ProvideAutoMaxPro,

	// Observability
	ProvideTracer,
	ProvideMetrics,
	ProvidePprofEndpoint,

	// Group all into DefaultServices struct
	NewDefaultServices,
)

// Provider functions for each service

// ProvideContext provides a background context
func ProvideContext() context.Context {
	return context.Background()
}

// ProvideLogger provides a configured logger instance
func ProvideLogger() logger.Logger {
	// Initialize logger with default configuration
	log, _ := logger.NewLogger(logger.INFO, logger.JSON)
	return log
}

// ProvideConfig provides application configuration
func ProvideConfig() (*config.Config, error) {
	// Load configuration from environment or config files
	return config.New()
}

// ProvideAutoMaxPro provides automatic max process configuration
func ProvideAutoMaxPro() autoMaxPro.AutoMaxPro {
	return autoMaxPro.New()
}

// ProvideTracer provides OpenTelemetry tracer
func ProvideTracer() trace.TracerProvider {
	// Initialize tracer with default configuration
	return trace.NewNoopTracerProvider()
}

// ProvideMetrics provides metrics monitoring
func ProvideMetrics() (*metrics.Monitoring, error) {
	// Initialize metrics monitoring
	return metrics.New()
}

// ProvidePprofEndpoint provides profiling endpoint
func ProvidePprofEndpoint() profiling.PprofEndpoint {
	return profiling.New()
}

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