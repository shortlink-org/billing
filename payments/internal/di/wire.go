//go:generate go tool wire
//go:build wireinject

// The build tag makes sure the stub is not built in the final build.
package di

import (
	"context"

	"github.com/google/wire"
	"go.opentelemetry.io/otel/trace"

	"github.com/shortlink-org/go-sdk/config"
	"github.com/shortlink-org/go-sdk/logger"
	"github.com/shortlink-org/shortlink/pkg/di"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/autoMaxPro"
	"github.com/shortlink-org/shortlink/pkg/di/pkg/profiling"
	"github.com/shortlink-org/shortlink/pkg/observability/metrics"

	"github.com/shortlink-org/billing/payments/internal/application/payments/usecase/create"
	"github.com/shortlink-org/billing/payments/internal/application/payments/usecase/refund"
)

type PaymentService struct {
	Log        logger.Logger
	Config     *config.Config
	AutoMaxPro autoMaxPro.AutoMaxPro
	Context    context.Context

	Tracer        trace.TracerProvider
	Metrics       *metrics.Monitoring
	PprofEndpoint profiling.PprofEndpoint

	CreatePayment *create.Handler
	RefundPayment *refund.Handler
}

var InfrastructureSet = wire.NewSet(
	ProvidePaymentRepository,
	ProvidePaymentProvider,
)

var UsecaseSet = wire.NewSet(
	ProvideCreateHandler,
	ProvideRefundHandler,
)

var PaymentSet = wire.NewSet(
	di.DefaultSet,
	InfrastructureSet,
	UsecaseSet,
	NewPaymentService,
)

func NewPaymentService(
	ctx context.Context,
	log logger.Logger,
	cfg *config.Config,
	auto autoMaxPro.AutoMaxPro,
	tr trace.TracerProvider,
	mon *metrics.Monitoring,
	pprof profiling.PprofEndpoint,
	createUC *create.Handler,
	refundUC *refund.Handler,
) (*PaymentService, error) {
	return &PaymentService{
		Context:       ctx,
		Log:           log,
		Config:        cfg,
		AutoMaxPro:    auto,
		Tracer:        tr,
		Metrics:       mon,
		PprofEndpoint: pprof,
		CreatePayment: createUC,
		RefundPayment: refundUC,
	}, nil
}

func InitializePaymentService() (*PaymentService, func(), error) {
	panic(wire.Build(PaymentSet))
}
