package payment

import (
	"fmt"

	eventv1 "github.com/shortlink-org/billing/payments/internal/domain/event/v1"
	flowv1 "github.com/shortlink-org/billing/payments/internal/domain/flow/v1"
	"github.com/shortlink-org/billing/payments/internal/domain/payment/ledger"

	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/proto"
)

// Payment — aggregate root.
type Payment struct {
	id        string
	invoiceID string

	kind        eventv1.PaymentKind
	captureMode eventv1.CaptureMode

	state   flowv1.PaymentFlow
	ledger  ledger.Ledger
	version uint64

	uncommitted []proto.Message
	guard       *guardFSM
}

// New: CREATED + PaymentCreated.
func New(id, invoiceID string, amount *money.Money, kind eventv1.PaymentKind, mode eventv1.CaptureMode) (*Payment, error) {
	if id == "" || invoiceID == "" || amount == nil {
		return nil, ErrInvalidArgs
	}
	p := &Payment{
		id:          id,
		invoiceID:   invoiceID,
		kind:        kind,
		captureMode: mode,
		state:       flowv1.PaymentFlow_PAYMENT_FLOW_CREATED,
		ledger:      ledger.Ledger{Amount: ledger.Clone(amount)},
		guard:       newGuard(flowv1.PaymentFlow_PAYMENT_FLOW_CREATED),
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

// Event-sourcing rehydration
func (p *Payment) Apply(e proto.Message) { p.apply(e) }

func (p *Payment) apply(e proto.Message) {
	switch ev := e.(type) {
	case *eventv1.PaymentCreated:
		p.id = ev.GetMeta().GetPaymentId()
		p.invoiceID = ev.GetInvoiceId()
		p.kind = ev.GetKind()
		p.captureMode = ev.GetCaptureMode()
		p.ledger.Amount = ledger.Clone(ev.GetAmount())
		p.state = flowv1.PaymentFlow_PAYMENT_FLOW_CREATED
		p.version = ev.GetMeta().GetVersion()
	case *eventv1.PaymentWaitingForConfirmation:
		p.state = flowv1.PaymentFlow_PAYMENT_FLOW_WAITING_FOR_CONFIRMATION
		p.version = ev.GetMeta().GetVersion()
	case *eventv1.PaymentAuthorized:
		p.ledger.Authorized = ledger.Clone(ev.GetAuthorizedAmount())
		p.state = flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED
		p.version = ev.GetMeta().GetVersion()
	case *eventv1.PaymentPaid:
		p.ledger.Captured = ledger.Clone(ev.GetCapturedAmount())
		p.state = flowv1.PaymentFlow_PAYMENT_FLOW_PAID
		p.version = ev.GetMeta().GetVersion()
	case *eventv1.PaymentRefunded:
		p.ledger.TotalRefunded = ledger.Clone(ev.GetTotalRefunded())
		p.state = flowv1.PaymentFlow_PAYMENT_FLOW_REFUNDED
		p.version = ev.GetMeta().GetVersion()
	case *eventv1.PaymentRefundFailed:
		p.version = ev.GetMeta().GetVersion() // no state change
	case *eventv1.PaymentCanceled:
		p.state = flowv1.PaymentFlow_PAYMENT_FLOW_CANCELED
		p.version = ev.GetMeta().GetVersion()
	case *eventv1.PaymentFailed:
		p.state = flowv1.PaymentFlow_PAYMENT_FLOW_FAILED
		p.version = ev.GetMeta().GetVersion()
	}
	p.guard = newGuard(p.state)
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
		// EventId full on level of outbox/publisher.
	}
}

func (p *Payment) record(evt proto.Message) { p.uncommitted = append(p.uncommitted, evt) }

func (p *Payment) String() string {
	return fmt.Sprintf("Payment{id=%s, state=%s, version=%d}", p.id, p.state.String(), p.version)
}
