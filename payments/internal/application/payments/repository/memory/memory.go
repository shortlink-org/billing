package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/shortlink-org/billing/payments/internal/application/payments/repository"
	"github.com/shortlink-org/billing/payments/internal/domain/payment"
)

// InMemory implements repository.PaymentRepository using an in-proc event store.
// Concurrency-safe; suitable for tests/dev.
type InMemory struct {
	mu       sync.RWMutex
	streams  map[uuid.UUID][]proto.Message // append-only event stream per aggregate
	versions map[uuid.UUID]uint64          // last persisted version per aggregate
}

// New returns a fresh in-memory repository.
func New() *InMemory {
	return &InMemory{
		streams:  make(map[uuid.UUID][]proto.Message),
		versions: make(map[uuid.UUID]uint64),
	}
}

var _ repository.PaymentRepository = (*InMemory)(nil)

func (r *InMemory) Save(_ context.Context, p *payment.Payment, expectedVersion uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := p.ID()
	cur := r.versions[id]

	// Optimistic concurrency
	if cur != expectedVersion {
		return payment.ErrVersionConflict
	}

	evts := p.UncommittedEvents()
	if len(evts) == 0 {
		return nil
	}

	// Append deep copies (defensive)
	dst := r.streams[id]
	for _, e := range evts {
		dst = append(dst, proto.Clone(e))
	}
	r.streams[id] = dst
	r.versions[id] = cur + uint64(len(evts))

	// Clear aggregate buffer after successful commit
	p.ClearUncommitted()
	return nil
}

func (r *InMemory) Load(_ context.Context, id uuid.UUID) (*payment.Payment, error) {
	r.mu.RLock()
	events, ok := r.streams[id]
	r.mu.RUnlock()

	if !ok {
		return nil, repository.ErrNotFound
	}
	// Rebuild aggregate.
	return payment.Rehydrate(events), nil
}
