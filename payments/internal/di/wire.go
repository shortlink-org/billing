//go:generate go tool wire
//go:build wireinject

package di

import (
	"context"

	"github.com/google/wire"
	"go.opentelemetry.io/otel/trace"

	"github.com/shortlink-org/shortlink/pkg/di"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/autoMaxPro"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/config"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/profiling"
	"github.com/shortlink-org/shortlink/pkg/logger"
	"github.com/shortlink-org/shortlink/pkg/observability/metrics"

	"github.com/shortlink-org/billing/payments/internal/application/payments/usecase/create"
	"github.com/shortlink-org/billing/payments/internal/application/payments/usecase/refund"
)

// PaymentService represents the payment service with common services and usecases
type PaymentService struct {
	// Common services
	Log        logger.Logger
	Config     *config.Config
	AutoMaxPro autoMaxPro.AutoMaxPro
	Context    context.Context

	// Observability
	Tracer        trace.TracerProvider
	Metrics       *metrics.Monitoring
	PprofEndpoint profiling.PprofEndpoint

	// Payment usecases
	CreatePayment *create.Handler
	RefundPayment *refund.Handler
}

// Wire provider sets for dependency injection
var (
	// InfrastructureSet provides all infrastructure dependencies
	InfrastructureSet = wire.NewSet(
		ProvidePaymentRepository,
		ProvidePaymentProvider,
	)

	// UsecaseSet provides all usecase handlers
	UsecaseSet = wire.NewSet(
		ProvideCreateHandler,
		ProvideRefundHandler,
	)

	// PaymentServiceSet combines default services with payment-specific providers
	PaymentServiceSet = wire.NewSet(
		di.DefaultSet,
		InfrastructureSet,
		UsecaseSet,
		NewPaymentService,
	)

	// AllProviders combines all provider sets (backward compatibility)
	AllProviders = wire.NewSet(
		InfrastructureSet,
		UsecaseSet,
	)
)

// NewPaymentService creates a new PaymentService with all dependencies
func NewPaymentService(
	// Common services
	ctx context.Context,
	log logger.Logger,
	config *config.Config,
	autoMaxPro autoMaxPro.AutoMaxPro,

	// Observability
	tracer trace.TracerProvider,
	metrics *metrics.Monitoring,
	pprofEndpoint profiling.PprofEndpoint,

	// Payment usecases
	createPayment *create.Handler,
	refundPayment *refund.Handler,
) (*PaymentService, error) {
	log.Info("Initializing PaymentService")

	return &PaymentService{
		// Common services
		Context:    ctx,
		Log:        log,
		Config:     config,
		AutoMaxPro: autoMaxPro,

		// Observability
		Tracer:        tracer,
		Metrics:       metrics,
		PprofEndpoint: pprofEndpoint,

		// Payment usecases
		CreatePayment: createPayment,
		RefundPayment: refundPayment,
	}, nil
}

// InitializePaymentService wires up the complete payment service with default services
func InitializePaymentService() (*PaymentService, func(), error) {
	panic(wire.Build(PaymentServiceSet))
}

// InitializeCreateHandler wires up the create payment usecase handler
func InitializeCreateHandler(cfg PaymentConfig) (*create.Handler, error) {
	wire.Build(AllProviders)
	return nil, nil
}

// InitializeRefundHandler wires up the refund payment usecase handler
func InitializeRefundHandler(cfg PaymentConfig) (*refund.Handler, error) {
	wire.Build(AllProviders)
	return nil, nil
}

// InitializePaymentUsecases wires up all payment usecase handlers
func InitializePaymentUsecases(cfg PaymentConfig) (*PaymentUsecases, error) {
	wire.Build(
		AllProviders,
		wire.Struct(new(PaymentUsecases), "*"),
	)
	return nil, nil
}
