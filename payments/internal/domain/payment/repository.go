package payment

import (
	"context"
)

// Repository abstracts persistence for Payment aggregates.
// Save must persist the aggregate and its uncommitted events atomically (with outbox) in real impl.
type Repository interface {
	// Get loads a payment by id. Returns ErrNotFound if there is no such aggregate.
	Get(ctx context.Context, id string) (*Payment, error)

	// Save persists the aggregate. expectedVersion is the version the caller believes is stored.
	// Implementations must enforce optimistic concurrency: if stored version != expectedVersion, return ErrVersionConflict.
	// On success, implementations SHOULD clear Uncommitted events (commit point).
	Save(ctx context.Context, p *Payment, expectedVersion uint64) error
}
