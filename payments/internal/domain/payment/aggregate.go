package payment

import (
	"fmt"

	eventv1 "github.com/shortlink-org/billing/payments/internal/domain/event/v1"
	flowv1 "github.com/shortlink-org/billing/payments/internal/domain/flow/v1"
	"github.com/shortlink-org/billing/payments/internal/domain/payment/fsm"
	"github.com/shortlink-org/billing/payments/internal/domain/payment/ledger"

	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/proto"
)

// Payment is the aggregate root of the Payments domain.
type Payment struct {
	id        string
	invoiceID string

	kind        eventv1.PaymentKind
	captureMode eventv1.CaptureMode

	state   flowv1.PaymentFlow
	Ledger  ledger.Ledger
	version uint64

	uncommitted []proto.Message
	guard       *fsm.Guard
	policy      Policy
}

// New constructs a Payment in CREATED state and emits PaymentCreated.
func New(id, invoiceID string, amount *money.Money, kind eventv1.PaymentKind, mode eventv1.CaptureMode, opts ...Option) (*Payment, error) {
	if id == "" || invoiceID == "" || amount == nil {
		return nil, ErrInvalidArgs
	}
	p := &Payment{
		id:          id,
		invoiceID:   invoiceID,
		kind:        kind,
		captureMode: mode,
		state:       flowv1.PaymentFlow_PAYMENT_FLOW_CREATED,
		Ledger:      ledger.Ledger{Amount: ledger.Clone(amount)},
		guard:       fsm.New(flowv1.PaymentFlow_PAYMENT_FLOW_CREATED),
		policy:      defaultPolicy,
	}
	for _, opt := range opts {
		opt(p)
	}
	// Currency policy gate
	if !p.policy.IsCurrencySupported(p.Ledger.Amount.GetCurrencyCode()) {
		return nil, ErrUnsupportedCurrency
	}

	ev := &eventv1.PaymentCreated{
		Meta:        p.metaNext(),
		InvoiceId:   invoiceID,
		Amount:      ledger.Clone(amount),
		Kind:        kind,
		CaptureMode: mode,
	}
	p.apply(ev)
	p.record(ev)
	return p, nil
}

// Queries / accessors
func (p *Payment) ID() string                         { return p.id }
func (p *Payment) State() flowv1.PaymentFlow          { return p.state }
func (p *Payment) Version() uint64                    { return p.version }
func (p *Payment) UncommittedEvents() []proto.Message { return p.uncommitted }
func (p *Payment) ClearUncommitted()                  { p.uncommitted = nil }

// Event-sourcing rehydration (no side effects except state/ledger mutation)
func (p *Payment) Apply(e proto.Message) { p.apply(e) }

func (p *Payment) apply(e proto.Message) {
	switch ev := e.(type) {
	case *eventv1.PaymentCreated:
		p.id = ev.GetMeta().GetPaymentId()
		p.invoiceID = ev.GetInvoiceId()
		p.kind = ev.GetKind()
		p.captureMode = ev.GetCaptureMode()
		p.Ledger.Amount = ledger.Clone(ev.GetAmount())
		p.state = flowv1.PaymentFlow_PAYMENT_FLOW_CREATED
		p.version = ev.GetMeta().GetVersion()

	case *eventv1.PaymentWaitingForConfirmation:
		p.state = flowv1.PaymentFlow_PAYMENT_FLOW_WAITING_FOR_CONFIRMATION
		p.version = ev.GetMeta().GetVersion()

	case *eventv1.PaymentAuthorized:
		// INCREMENTAL add — do not overwrite
		if p.Ledger.Authorized == nil {
			p.Ledger.Authorized = ledger.Clone(ev.GetAuthorizedAmount())
		} else {
			sum, _ := ledger.Add(p.Ledger.Authorized, ev.GetAuthorizedAmount())
			p.Ledger.Authorized = sum
		}
		p.state = flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED
		p.version = ev.GetMeta().GetVersion()

	case *eventv1.PaymentPaid:
		// INCREMENTAL add — do not overwrite
		if p.Ledger.Captured == nil {
			p.Ledger.Captured = ledger.Clone(ev.GetCapturedAmount())
		} else {
			sum, _ := ledger.Add(p.Ledger.Captured, ev.GetCapturedAmount())
			p.Ledger.Captured = sum
		}
		p.state = flowv1.PaymentFlow_PAYMENT_FLOW_PAID
		p.version = ev.GetMeta().GetVersion()

	case *eventv1.PaymentRefunded:
		// Use event-carried total for deterministic rehydration.
		p.Ledger.TotalRefunded = ledger.Clone(ev.GetTotalRefunded())
		if ev.GetFull() {
			p.state = flowv1.PaymentFlow_PAYMENT_FLOW_REFUNDED
		} else {
			p.state = flowv1.PaymentFlow_PAYMENT_FLOW_PAID
		}
		p.version = ev.GetMeta().GetVersion()

	case *eventv1.PaymentRefundFailed:
		// state unchanged; version++
		p.version = ev.GetMeta().GetVersion()

	case *eventv1.PaymentCanceled:
		p.state = flowv1.PaymentFlow_PAYMENT_FLOW_CANCELED
		p.version = ev.GetMeta().GetVersion()

	case *eventv1.PaymentFailed:
		p.state = flowv1.PaymentFlow_PAYMENT_FLOW_FAILED
		p.version = ev.GetMeta().GetVersion()
	}

	// keep FSM in sync with the latest state
	p.guard = fsm.New(p.state)
}

func (p *Payment) isTerminal() bool {
	switch p.state {
	case flowv1.PaymentFlow_PAYMENT_FLOW_REFUNDED,
		flowv1.PaymentFlow_PAYMENT_FLOW_CANCELED,
		flowv1.PaymentFlow_PAYMENT_FLOW_FAILED:
		return true
	default:
		return false
	}
}

func (p *Payment) metaNext() *eventv1.EventMeta {
	p.version++
	return &eventv1.EventMeta{
		PaymentId: p.id,
		Version:   p.version,
		// EventId is filled at outbox/publisher layer.
	}
}

func (p *Payment) record(evt proto.Message) { p.uncommitted = append(p.uncommitted, evt) }

func (p *Payment) String() string {
	return fmt.Sprintf("Payment{id=%s, state=%s, version=%d}", p.id, p.state.String(), p.version)
}

// Invariants validates currency consistency and ledger bounds.
// Repository implementations MUST call this before storing the aggregate.
func (p *Payment) Invariants() error {
	// Currency policy gate
	cur := p.Ledger.Amount.GetCurrencyCode()
	if !p.policy.IsCurrencySupported(cur) {
		return ErrUnsupportedCurrency
	}
	for _, m := range []*money.Money{p.Ledger.Authorized, p.Ledger.Captured, p.Ledger.TotalRefunded} {
		if m == nil {
			continue
		}
		if ledger.Currency(m) != cur {
			return ErrInvariantViolation
		}
	}

	// Authorized ≤ Amount
	if p.Ledger.Authorized != nil && ledger.Compare(p.Ledger.Authorized, p.Ledger.Amount) > 0 {
		return ErrInvariantViolation
	}

	// Captured ≤ (Authorized if present, else Amount)
	lim := p.Ledger.Amount
	if p.Ledger.Authorized != nil {
		lim = p.Ledger.Authorized
	}
	if p.Ledger.Captured != nil && ledger.Compare(p.Ledger.Captured, lim) > 0 {
		return ErrInvariantViolation
	}

	// TotalRefunded ≤ Captured
	if p.Ledger.TotalRefunded != nil && p.Ledger.Captured != nil &&
		ledger.Compare(p.Ledger.TotalRefunded, p.Ledger.Captured) > 0 {
		return ErrInvariantViolation
	}

	// Policy: CREATED->PAID immediate capture not allowed for MANUAL mode (i.e., no auth recorded).
	if p.state == flowv1.PaymentFlow_PAYMENT_FLOW_PAID &&
		p.captureMode == eventv1.CaptureMode_CAPTURE_MODE_MANUAL &&
		p.Ledger.Authorized == nil {
		return ErrPolicyCaptureMode
	}

	return nil
}
