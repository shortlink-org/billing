package payment

import (
	"context"
	"fmt"

	"github.com/looplab/fsm"
	flowv1 "github.com/shortlink-org/billing/payments/internal/domain/flow/v1"
)

// Domain-level event names (triggers) for the guard FSM.
const (
	EventSCARequired = "sca_required" // CREATED -> WAITING_FOR_CONFIRMATION
	EventConfirm     = "confirm"      // WAITING_FOR_CONFIRMATION -> AUTHORIZED
	EventAuthorize   = "authorize"    // CREATED -> AUTHORIZED (skip SCA)
	EventCapture     = "capture"      // AUTHORIZED|CREATED -> PAID
	EventRefund      = "refund"       // PAID -> REFUNDED
	EventCancel      = "cancel"       // CREATED|WAITING|AUTHORIZED -> CANCELED
	EventFail        = "fail"         // CREATED|WAITING|AUTHORIZED -> FAILED
)

// guardTransitions — common list of allowed transitions for the guard FSM.
var guardTransitions = fsm.Events{
	{
		Name: EventSCARequired,
		Src: []string{
			flowv1.PaymentFlow_PAYMENT_FLOW_CREATED.String(),
		},
		Dst: flowv1.PaymentFlow_PAYMENT_FLOW_WAITING_FOR_CONFIRMATION.String(),
	},
	{
		Name: EventConfirm,
		Src: []string{
			flowv1.PaymentFlow_PAYMENT_FLOW_WAITING_FOR_CONFIRMATION.String(),
		},
		Dst: flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED.String(),
	},
	{
		Name: EventAuthorize,
		Src: []string{
			flowv1.PaymentFlow_PAYMENT_FLOW_CREATED.String(),
		},
		Dst: flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED.String(),
	},
	{
		Name: EventCapture,
		Src: []string{
			flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED.String(),
			flowv1.PaymentFlow_PAYMENT_FLOW_CREATED.String(), // immediate capture
		},
		Dst: flowv1.PaymentFlow_PAYMENT_FLOW_PAID.String(),
	},
	{
		Name: EventRefund,
		Src: []string{
			flowv1.PaymentFlow_PAYMENT_FLOW_PAID.String(),
		},
		Dst: flowv1.PaymentFlow_PAYMENT_FLOW_REFUNDED.String(),
	},
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

// guardFSM is a thin validator for transitions. No side effects or event emission.
type guardFSM struct{ m *fsm.FSM }

// newGuard create FSM with initial state.
func newGuard(initial flowv1.PaymentFlow) *guardFSM {
	init := initial.String()
	if init == "" {
		panic(fmt.Errorf("%w: %v", ErrInvalidArgs, initial))
	}
	return &guardFSM{
		m: fsm.NewFSM(
			init,
			guardTransitions,
			nil, // no callbacks — domain emits events itself
		),
	}
}

// Trigger tries to move the FSM along the given event name.
func (g *guardFSM) Trigger(ctx context.Context, name string, args ...interface{}) error {
	if err := g.m.Event(ctx, name, args...); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidTransition, err)
	}

	return nil
}
