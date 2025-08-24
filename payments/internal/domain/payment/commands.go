package payment

import (
	"context"

	eventv1 "github.com/shortlink-org/billing/payments/internal/domain/event/v1"
	flowv1 "github.com/shortlink-org/billing/payments/internal/domain/flow/v1"
	"github.com/shortlink-org/billing/payments/internal/domain/payment/ledger"

	"google.golang.org/genproto/googleapis/type/money"
)

// RequireSCA: CREATED -> WAITING_FOR_CONFIRMATION
func (p *Payment) RequireSCA(ctx context.Context) error {
	if p.isTerminal() {
		return ErrTerminalState
	}
	if err := p.guard.Trigger(ctx, EventSCARequired); err != nil {
		return err
	}
	ev := &eventv1.PaymentWaitingForConfirmation{Meta: p.metaNext()}
	p.apply(ev)
	p.record(ev)
	return nil
}

// Authorize: CREATED -> AUTHORIZED
func (p *Payment) Authorize(ctx context.Context, amt *money.Money) error {
	if p.isTerminal() {
		return ErrTerminalState
	}
	if err := p.guard.Trigger(ctx, EventAuthorize); err != nil {
		return err
	}
	if err := p.ledger.Authorize(amt); err != nil {
		return err
	}
	ev := &eventv1.PaymentAuthorized{
		Meta:             p.metaNext(),
		AuthorizedAmount: ledger.Clone(amt),
	}
	p.apply(ev)
	p.record(ev)
	return nil
}

// Confirm: WAITING_FOR_CONFIRMATION -> AUTHORIZED
func (p *Payment) Confirm(ctx context.Context, amt *money.Money) error {
	if p.isTerminal() {
		return ErrTerminalState
	}
	if err := p.guard.Trigger(ctx, EventConfirm); err != nil {
		return err
	}
	if err := p.ledger.Authorize(amt); err != nil {
		return err
	}
	ev := &eventv1.PaymentAuthorized{
		Meta:             p.metaNext(),
		AuthorizedAmount: ledger.Clone(amt),
	}
	p.apply(ev)
	p.record(ev)
	return nil
}

// Capture: AUTHORIZED|CREATED -> PAID
func (p *Payment) Capture(ctx context.Context, amt *money.Money) error {
	if p.isTerminal() {
		return ErrTerminalState
	}

	if p.state == flowv1.PaymentFlow_PAYMENT_FLOW_CREATED && p.captureMode == eventv1.CaptureMode_CAPTURE_MODE_MANUAL {
		return ErrPolicyCaptureMode
	}

	if err := p.guard.Trigger(ctx, EventCapture); err != nil {
		return err
	}
	if err := p.ledger.Capture(amt); err != nil {
		return err
	}
	ev := &eventv1.PaymentPaid{
		Meta:           p.metaNext(),
		CapturedAmount: ledger.Clone(amt),
	}
	p.apply(ev)
	p.record(ev)
	return nil
}

// Refund: PAID -> REFUNDED (на первую успешную операцию),
// затем остаёмся в REFUNDED и накапливаем total_refunded.
func (p *Payment) Refund(ctx context.Context, amt *money.Money) (bool, error) {
	if p.isTerminal() {
		return false, ErrTerminalState
	}
	if err := p.guard.Trigger(ctx, EventRefund); err != nil {
		return false, err
	}
	full, err := p.ledger.Refund(amt)
	if err != nil {
		return false, err
	}
	ev := &eventv1.PaymentRefunded{
		Meta:          p.metaNext(),
		RefundAmount:  ledger.Clone(amt),
		TotalRefunded: ledger.Clone(p.ledger.TotalRefunded),
		Full:          full,
	}
	p.apply(ev)
	p.record(ev)
	return full, nil
}

// RefundFailed: остаёмся в PAID; только событие и версия (без перехода FSM).
func (p *Payment) RefundFailed(ctx context.Context, code, msg string) {
	_ = ctx // резерв на будущее (трассировка и т.п.)
	ev := &eventv1.PaymentRefundFailed{
		Meta:          p.metaNext(),
		ReasonCode:    code,
		ReasonMessage: msg,
	}
	p.apply(ev) // version++
	p.record(ev)
}

// Cancel: CREATED|WAITING|AUTHORIZED -> CANCELED
func (p *Payment) Cancel(ctx context.Context, reason string) error {
	if p.isTerminal() {
		return ErrTerminalState
	}
	if err := p.guard.Trigger(ctx, EventCancel); err != nil {
		return err
	}
	ev := &eventv1.PaymentCanceled{Meta: p.metaNext(), Reason: reason}
	p.apply(ev)
	p.record(ev)
	return nil
}

// Fail: CREATED|WAITING|AUTHORIZED -> FAILED
func (p *Payment) Fail(ctx context.Context, code, msg string) error {
	if p.isTerminal() {
		return ErrTerminalState
	}
	if err := p.guard.Trigger(ctx, EventFail); err != nil {
		return err
	}
	ev := &eventv1.PaymentFailed{
		Meta:          p.metaNext(),
		ReasonCode:    code,
		ReasonMessage: msg,
	}
	p.apply(ev)
	p.record(ev)
	return nil
}
