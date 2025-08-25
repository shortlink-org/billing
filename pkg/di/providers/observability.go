package providers

import (
	"go.opentelemetry.io/otel/trace"

	"github.com/shortlink-org/shortlink/pkg/di/pkg/profiling"
	"github.com/shortlink-org/shortlink/pkg/observability/metrics"
)

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