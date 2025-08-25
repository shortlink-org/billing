package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/shortlink-org/billing/payments/internal/domain/payment"
)

// PaymentRepository is the abstraction used by use cases.
// Implementations live under internal/repository/payment/{memory,postgres,...}.
type PaymentRepository interface {
	// Save appends the aggregate's uncommitted events if the current persisted
	// version equals expectedVersion (optimistic concurrency). On success the
	// aggregate must clear its uncommitted buffer.
	Save(ctx context.Context, p *payment.Payment, expectedVersion uint64) error

	// Load reconstructs a payment aggregate by its ID or returns ErrNotFound.
	Load(ctx context.Context, id uuid.UUID) (*payment.Payment, error)
}
