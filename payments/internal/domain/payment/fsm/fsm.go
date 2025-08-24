// Package fsm contains a side-effect-free guard state machine for the Payments domain.
//
// Design notes:
//   - This FSM is a *validator only*: it encodes which transitions are allowed.
//     It does not emit domain events and does not mutate aggregates.
//   - Business policies (e.g., disallow immediate capture in MANUAL mode) are enforced
//     in the domain layer, *not* here. The FSM remains provider- and policy-agnostic.
//   - Refund semantics are explicit:
//   - partial refund keeps the state in PAID (self-loop),
//   - full refund moves to REFUNDED (terminal).
package fsm

import (
	"context"

	"github.com/looplab/fsm"

	flowv1 "github.com/shortlink-org/billing/payments/internal/domain/flow/v1"
)

// Domain-level event names (triggers) for the guard FSM.
//
// Refund semantics:
//   - EventRefundPartial: PAID -> PAID (partial refund; state stays PAID)
//   - EventRefundFull:    PAID -> REFUNDED (full refund; terminal)
//
// NOTE: The FSM allows CREATED -> PAID (immediate capture). Whether this is permitted
// in a particular scenario must be decided by the domain Policy (e.g., disallow for MANUAL).
const (
	EventSCARequired   = "sca_required"   // CREATED -> WAITING_FOR_CONFIRMATION
	EventConfirm       = "confirm"        // WAITING_FOR_CONFIRMATION -> AUTHORIZED
	EventAuthorize     = "authorize"      // CREATED -> AUTHORIZED (skip SCA)
	EventCapture       = "capture"        // AUTHORIZED|CREATED -> PAID (Policy may restrict CREATED)
	EventRefundPartial = "refund_partial" // PAID -> PAID (self-loop) — NOTE: not triggered in FSM
	EventRefundFull    = "refund_full"    // PAID -> REFUNDED
	EventCancel        = "cancel"         // CREATED|WAITING|AUTHORIZED -> CANCELED
	EventFail          = "fail"           // CREATED|WAITING|AUTHORIZED -> FAILED
)

// guardTransitions — canonical list of allowed transitions for the guard FSM.
// No callbacks or side effects here — domain logic owns event emission and mutations.
//
// IMPORTANT:
//
//	looplab/fsm does not support self-transitions. Therefore, partial refunds do not
//	trigger the FSM at all (state remains PAID); only full refunds are modeled here.
var guardTransitions = fsm.Events{
	{
		Name: EventSCARequired,
		Src:  []string{flowv1.PaymentFlow_PAYMENT_FLOW_CREATED.String()},
		Dst:  flowv1.PaymentFlow_PAYMENT_FLOW_WAITING_FOR_CONFIRMATION.String(),
	},
	{
		Name: EventConfirm,
		Src:  []string{flowv1.PaymentFlow_PAYMENT_FLOW_WAITING_FOR_CONFIRMATION.String()},
		Dst:  flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED.String(),
	},
	{
		Name: EventAuthorize,
		Src:  []string{flowv1.PaymentFlow_PAYMENT_FLOW_CREATED.String()},
		Dst:  flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED.String(),
	},
	{
		Name: EventCapture,
		Src: []string{
			flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED.String(),
			flowv1.PaymentFlow_PAYMENT_FLOW_CREATED.String(), // immediate capture (domain Policy may forbid)
		},
		Dst: flowv1.PaymentFlow_PAYMENT_FLOW_PAID.String(),
	},

	// Refunds:
	// Full — transition to REFUNDED (terminal)
	{
		Name: EventRefundFull,
		Src:  []string{flowv1.PaymentFlow_PAYMENT_FLOW_PAID.String()},
		Dst:  flowv1.PaymentFlow_PAYMENT_FLOW_REFUNDED.String(),
	},

	// Problem exits:
	{
		Name: EventCancel,
		Src: []string{
			flowv1.PaymentFlow_PAYMENT_FLOW_CREATED.String(),
			flowv1.PaymentFlow_PAYMENT_FLOW_WAITING_FOR_CONFIRMATION.String(),
			flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED.String(),
		},
		Dst: flowv1.PaymentFlow_PAYMENT_FLOW_CANCELED.String(),
	},
	{
		Name: EventFail,
		Src: []string{
			flowv1.PaymentFlow_PAYMENT_FLOW_CREATED.String(),
			flowv1.PaymentFlow_PAYMENT_FLOW_WAITING_FOR_CONFIRMATION.String(),
			flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED.String(),
		},
		Dst: flowv1.PaymentFlow_PAYMENT_FLOW_FAILED.String(),
	},
}

// Guard validates transitions. It is intentionally thin over looplab/fsm.
type Guard struct{ m *fsm.FSM }

// New builds a new FSM with the given initial state.
//
// Important:
//   - The caller (domain layer) is responsible for providing a valid initial state.
//   - This function does not validate business policies; it only encodes graph edges.
func New(initial flowv1.PaymentFlow) *Guard {
	return &Guard{
		m: fsm.NewFSM(initial.String(), guardTransitions, nil),
	}
}

// Trigger attempts to move the FSM along the given event name.
//
// It returns a non-nil error if the transition is not allowed from the current state.
// The domain layer should wrap/map this error to its own sentinel (e.g., ErrInvalidTransition).
func (g *Guard) Trigger(ctx context.Context, name string, args ...interface{}) error {
	return g.m.Event(ctx, name, args...)
}

// AllowedTransitions lists all triggers currently permitted from the FSM's state.
// Useful for diagnostics and tests (order is not guaranteed).
func (g *Guard) AllowedTransitions() []string {
	return g.m.AvailableTransitions()
}

// IsAllowed returns true if the given trigger is currently permitted.
// This is a convenience wrapper over AvailableTransitions.
func (g *Guard) IsAllowed(name string) bool {
	for _, n := range g.m.AvailableTransitions() {
		if n == name {
			return true
		}
	}
	return false
}

// Current returns the current state name as a string (for logging/diagnostics).
// Example: "PAYMENT_FLOW_PAID".
func (g *Guard) Current() string { return g.m.Current() }
