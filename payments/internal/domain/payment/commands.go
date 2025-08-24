package payment

import (
	"context"
	"fmt"

	eventv1 "github.com/shortlink-org/billing/payments/internal/domain/event/v1"
	flowv1 "github.com/shortlink-org/billing/payments/internal/domain/flow/v1"
	"github.com/shortlink-org/billing/payments/internal/domain/payment/fsm"
	"github.com/shortlink-org/billing/payments/internal/domain/payment/ledger"

	"google.golang.org/genproto/googleapis/type/money"
)

// RequireSCA: CREATED -> WAITING_FOR_CONFIRMATION
func (p *Payment) RequireSCA(ctx context.Context) error {
	if p.isTerminal() {
		return ErrTerminalState
	}
	if err := p.guard.Trigger(ctx, fsm.EventSCARequired); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidTransition, err)
	}
	ev := &eventv1.PaymentWaitingForConfirmation{Meta: p.metaNext()}
	p.apply(ev)
	p.record(ev)
	return nil
}

// Authorize:
//   - CREATED    -> AUTHORIZED (via FSM)    + emit incremental PaymentAuthorized
//   - AUTHORIZED -> AUTHORIZED (no FSM)     + emit incremental PaymentAuthorized
func (p *Payment) Authorize(ctx context.Context, amt *money.Money) error {
	if p.isTerminal() {
		return ErrTerminalState
	}

	switch p.state {
	case flowv1.PaymentFlow_PAYMENT_FLOW_CREATED:
		if err := p.guard.Trigger(ctx, fsm.EventAuthorize); err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidTransition, err)
		}
	case flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED:
		// incremental top-up — no FSM trigger
	default:
		return ErrInvalidTransition
	}

	// Validate next authorized ≤ amount (no mutation here)
	cur := ledger.Clone(p.Ledger.Authorized)
	if cur == nil {
		cur = ledger.Zero(p.Ledger.Amount.GetCurrencyCode())
	}
	next, err := ledger.Add(cur, amt)
	if err != nil {
		return err
	}
	if ledger.Compare(next, p.Ledger.Amount) > 0 {
		return ledger.ErrAuthorizeExceeds
	}

	// Emit incremental event; apply() will accumulate.
	ev := &eventv1.PaymentAuthorized{
		Meta:             p.metaNext(),
		AuthorizedAmount: ledger.Clone(amt),
	}
	p.apply(ev)
	p.record(ev)
	return nil
}

// Confirm: WAITING_FOR_CONFIRMATION -> AUTHORIZED (incremental authorize)
func (p *Payment) Confirm(ctx context.Context, amt *money.Money) error {
	if p.isTerminal() {
		return ErrTerminalState
	}
	if p.state != flowv1.PaymentFlow_PAYMENT_FLOW_WAITING_FOR_CONFIRMATION {
		return ErrInvalidTransition
	}
	if err := p.guard.Trigger(ctx, fsm.EventConfirm); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidTransition, err)
	}

	// Same validation as Authorize (no mutation)
	cur := ledger.Clone(p.Ledger.Authorized)
	if cur == nil {
		cur = ledger.Zero(p.Ledger.Amount.GetCurrencyCode())
	}
	next, err := ledger.Add(cur, amt)
	if err != nil {
		return err
	}
	if ledger.Compare(next, p.Ledger.Amount) > 0 {
		return ledger.ErrAuthorizeExceeds
	}

	ev := &eventv1.PaymentAuthorized{
		Meta:             p.metaNext(),
		AuthorizedAmount: ledger.Clone(amt),
	}
	p.apply(ev)
	p.record(ev)
	return nil
}

// Capture: AUTHORIZED|CREATED -> PAID (Policy may forbid CREATED immediate)
func (p *Payment) Capture(ctx context.Context, amt *money.Money) error {
	if p.isTerminal() {
		return ErrTerminalState
	}
	// Policy gate for immediate capture from CREATED
	if p.state == flowv1.PaymentFlow_PAYMENT_FLOW_CREATED &&
		!p.policy.AllowImmediateCapture(p.kind, p.captureMode) {
		return ErrPolicyCaptureMode
	}

	// Validate next captured ≤ limit (no mutation)
	cur := ledger.Clone(p.Ledger.Captured)
	if cur == nil {
		cur = ledger.Zero(p.Ledger.Amount.GetCurrencyCode())
	}
	next, err := ledger.Add(cur, amt)
	if err != nil {
		return err
	}
	limit := p.Ledger.Amount
	if p.Ledger.Authorized != nil {
		limit = p.Ledger.Authorized
	}
	if ledger.Compare(next, limit) > 0 {
		return ledger.ErrCaptureExceedsLimit
	}

	// FSM trigger per current state
	if err := p.guard.Trigger(ctx, fsm.EventCapture); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidTransition, err)
	}

	// Emit incremental captured
	ev := &eventv1.PaymentPaid{
		Meta:           p.metaNext(),
		CapturedAmount: ledger.Clone(amt),
	}
	p.apply(ev)
	p.record(ev)
	return nil
}

// Refund: partial -> stay PAID; full -> FSM refund_full -> REFUNDED.
// Validate next total_refunded ≤ captured (no mutation).
func (p *Payment) Refund(ctx context.Context, amt *money.Money) (bool, error) {
	if p.isTerminal() {
		return false, ErrTerminalState
	}
	if p.Ledger.Captured == nil {
		return false, ledger.ErrRefundWithoutCapture
	}
	cur := ledger.Clone(p.Ledger.TotalRefunded)
	if cur == nil {
		cur = ledger.Zero(p.Ledger.Captured.GetCurrencyCode())
	}
	next, err := ledger.Add(cur, amt)
	if err != nil {
		return false, err
	}
	if ledger.Compare(next, p.Ledger.Captured) > 0 {
		return false, ledger.ErrRefundExceeds
	}
	full := ledger.Compare(next, p.Ledger.Captured) == 0

	if full {
		if err := p.guard.Trigger(ctx, fsm.EventRefundFull); err != nil {
			return false, fmt.Errorf("%w: %s", ErrInvalidTransition, err)
		}
	}

	ev := &eventv1.PaymentRefunded{
		Meta:          p.metaNext(),
		RefundAmount:  ledger.Clone(amt),
		TotalRefunded: ledger.Clone(next), // carry the new total for deterministic rehydration
		Full:          full,
	}
	p.apply(ev)
	p.record(ev)
	return full, nil
}

// RefundFailed: stays in PAID; version++ only (enum reason).
func (p *Payment) RefundFailed(ctx context.Context, reason eventv1.FailureReason) {
	_ = ctx
	ev := &eventv1.PaymentRefundFailed{
		Meta:   p.metaNext(),
		Reason: reason,
	}
	p.apply(ev)
	p.record(ev)
}

// Cancel: CREATED|WAITING|AUTHORIZED -> CANCELED (enum reason)
func (p *Payment) Cancel(ctx context.Context, reason eventv1.CancelReason) error {
	if p.isTerminal() {
		return ErrTerminalState
	}
	if err := p.guard.Trigger(ctx, fsm.EventCancel); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidTransition, err)
	}
	ev := &eventv1.PaymentCanceled{Meta: p.metaNext(), Reason: reason}
	p.apply(ev)
	p.record(ev)
	return nil
}

// Fail: CREATED|WAITING|AUTHORIZED -> FAILED (enum reason)
func (p *Payment) Fail(ctx context.Context, reason eventv1.FailureReason) error {
	if p.isTerminal() {
		return ErrTerminalState
	}
	if err := p.guard.Trigger(ctx, fsm.EventFail); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidTransition, err)
	}
	ev := &eventv1.PaymentFailed{Meta: p.metaNext(), Reason: reason}
	p.apply(ev)
	p.record(ev)
	return nil
}
