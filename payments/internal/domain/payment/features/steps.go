package payment

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
	"github.com/google/uuid"

	eventv1 "github.com/shortlink-org/billing/payments/internal/domain/event/v1"
	flowv1 "github.com/shortlink-org/billing/payments/internal/domain/flow/v1"
	"github.com/shortlink-org/billing/payments/internal/domain/payment"

	"google.golang.org/genproto/googleapis/type/money"
)

// ---- world ----

type paymentWorld struct {
	ctx      context.Context
	id       uuid.UUID
	invoice  uuid.UUID
	amount   *money.Money
	kind     eventv1.PaymentKind
	mode     eventv1.CaptureMode
	p        *payment.Payment
	lastErr  error
	lastFull *bool
}

func (w *paymentWorld) reset(ctx context.Context) {
	w.ctx = ctx
	w.id = uuid.Nil
	w.invoice = uuid.Nil
	w.amount = nil
	w.kind = eventv1.PaymentKind_PAYMENT_KIND_UNSPECIFIED
	w.mode = eventv1.CaptureMode_CAPTURE_MODE_UNSPECIFIED
	w.p = nil
	w.lastErr = nil
	w.lastFull = nil
}

func (w *paymentWorld) ensureCreated() error {
	if w.p != nil {
		return nil
	}
	if w.id == uuid.Nil || w.invoice == uuid.Nil || w.amount == nil ||
		w.kind == eventv1.PaymentKind_PAYMENT_KIND_UNSPECIFIED ||
		w.mode == eventv1.CaptureMode_CAPTURE_MODE_UNSPECIFIED {
		return nil
	}
	var err error
	w.p, err = payment.New(w.id, w.invoice, w.amount, w.kind, w.mode)
	return err
}

// ---- helpers ----

func parseMoney(input string) (*money.Money, error) {
	parts := strings.Fields(strings.TrimSpace(input))
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid money: %q", input)
	}
	cur := strings.ToUpper(parts[0])
	dec := parts[1]
	re := regexp.MustCompile(`^([0-9]+)(?:\.([0-9]{1,9}))?$`)
	m := re.FindStringSubmatch(dec)
	if m == nil {
		return nil, fmt.Errorf("invalid decimal: %q", dec)
	}
	intPart := m[1]
	frac := ""
	if len(m) > 2 {
		frac = m[2]
	}
	for len(frac) < 9 {
		frac += "0"
	}
	units, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil {
		return nil, err
	}
	nanos, err := strconv.ParseInt(frac, 10, 32)
	if err != nil {
		return nil, err
	}
	return &money.Money{
		CurrencyCode: cur,
		Units:        units,
		Nanos:        int32(nanos),
	}, nil
}

func moneyEq(a, b *money.Money) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.GetCurrencyCode() == b.GetCurrencyCode() &&
		a.GetUnits() == b.GetUnits() &&
		a.GetNanos() == b.GetNanos()
}

func enumState(flow flowv1.PaymentFlow) string {
	return strings.TrimPrefix(flow.String(), "PAYMENT_FLOW_")
}

func kindFrom(s string) (eventv1.PaymentKind, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "ONE_TIME":
		return eventv1.PaymentKind_PAYMENT_KIND_ONE_TIME, nil
	case "SUBSCRIPTION", "RECURRING":
		return eventv1.PaymentKind_PAYMENT_KIND_RECURRING, nil
	default:
		return eventv1.PaymentKind_PAYMENT_KIND_UNSPECIFIED, fmt.Errorf("unknown kind %q", s)
	}
}

func modeFrom(s string) (eventv1.CaptureMode, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "IMMEDIATE":
		return eventv1.CaptureMode_CAPTURE_MODE_IMMEDIATE, nil
	case "MANUAL":
		return eventv1.CaptureMode_CAPTURE_MODE_MANUAL, nil
	default:
		return eventv1.CaptureMode_CAPTURE_MODE_UNSPECIFIED, fmt.Errorf("unknown capture mode %q", s)
	}
}

func eventTypeName(m any) string {
	t := reflect.TypeOf(m)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// --- reason parsers (align with internal event FailureReason) ---

func parseFailureReason(s string) (eventv1.FailureReason, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DECLINED":
		return eventv1.FailureReason_FAILURE_REASON_DECLINED, nil
	case "REVERSED":
		return eventv1.FailureReason_FAILURE_REASON_REVERSED, nil
	case "AUTH_EXPIRED", "EXPIRED":
		return eventv1.FailureReason_FAILURE_REASON_AUTH_EXPIRED, nil
	case "NETWORK_ERROR", "NETWORK":
		return eventv1.FailureReason_FAILURE_REASON_NETWORK_ERROR, nil
	default:
		return eventv1.FailureReason_FAILURE_REASON_UNSPECIFIED,
			fmt.Errorf("unknown failure reason %q", s)
	}
}

func parseCancelReason(s string) (eventv1.CancelReason, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "USER":
		return eventv1.CancelReason_CANCEL_REASON_USER, nil
	case "SYSTEM":
		return eventv1.CancelReason_CANCEL_REASON_SYSTEM, nil
	case "AUTH_VOID", "VOID":
		return eventv1.CancelReason_CANCEL_REASON_AUTH_VOID, nil
	case "DUPLICATE":
		return eventv1.CancelReason_CANCEL_REASON_DUPLICATE, nil
	default:
		return eventv1.CancelReason_CANCEL_REASON_UNSPECIFIED, fmt.Errorf("unknown cancel reason %q", s)
	}
}

// ---- steps (Given/When/Then) ----

func (w *paymentWorld) givenPaymentCreatedForInvoice(id, invoice string) error {
	pid, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return fmt.Errorf("bad payment id: %w", err)
	}
	inv, err := uuid.Parse(strings.TrimSpace(invoice))
	if err != nil {
		return fmt.Errorf("bad invoice id: %w", err)
	}
	w.id, w.invoice = pid, inv
	return w.ensureCreated()
}

func (w *paymentWorld) andAmountIs(s string) error {
	m, err := parseMoney(s)
	if err != nil {
		return err
	}
	w.amount = m
	return w.ensureCreated()
}

func (w *paymentWorld) andKindIs(s string) error {
	k, err := kindFrom(s)
	if err != nil {
		return err
	}
	w.kind = k
	return w.ensureCreated()
}

func (w *paymentWorld) andCaptureModeIs(s string) error {
	m, err := modeFrom(s)
	if err != nil {
		return err
	}
	w.mode = m
	return w.ensureCreated()
}

func (w *paymentWorld) whenRequireSCA() error {
	if err := w.ensureCreated(); err != nil {
		return err
	}
	w.lastErr = w.p.RequireSCA(w.ctx)
	return w.lastErr
}

func (w *paymentWorld) whenConfirmAuthorizationOf(s string) error {
	if err := w.ensureCreated(); err != nil {
		return err
	}
	m, err := parseMoney(s)
	if err != nil {
		return err
	}
	w.lastErr = w.p.Confirm(w.ctx, m)
	return w.lastErr
}

func (w *paymentWorld) whenAuthorize(s string) error {
	if err := w.ensureCreated(); err != nil {
		return err
	}
	m, err := parseMoney(s)
	if err != nil {
		return err
	}
	w.lastErr = w.p.Authorize(w.ctx, m)
	return w.lastErr
}

func (w *paymentWorld) whenTryAuthorize(s string) error {
	err := w.whenAuthorize(s)
	if err == nil {
		return fmt.Errorf("expected error, got nil")
	}
	return nil
}

func (w *paymentWorld) whenCapture(s string) error {
	if err := w.ensureCreated(); err != nil {
		return err
	}
	m, err := parseMoney(s)
	if err != nil {
		return err
	}
	w.lastErr = w.p.Capture(w.ctx, m)
	return w.lastErr
}

func (w *paymentWorld) whenTryCapture(s string) error {
	err := w.whenCapture(s)
	if err == nil {
		return fmt.Errorf("expected error, got nil")
	}
	return nil
}

func (w *paymentWorld) whenRefund(s string) error {
	if err := w.ensureCreated(); err != nil {
		return err
	}
	m, err := parseMoney(s)
	if err != nil {
		return err
	}
	full, e := w.p.Refund(w.ctx, m)
	w.lastErr = e
	w.lastFull = &full
	return e
}

// Refund failed (enum)
func (w *paymentWorld) whenRefundFailedWithReason(s string) error {
	if err := w.ensureCreated(); err != nil {
		return err
	}
	r, err := parseFailureReason(s)
	if err != nil {
		return err
	}
	w.p.RefundFailed(w.ctx, r)
	return nil
}

// Cancel (enum)
func (w *paymentWorld) whenCancel(reason string) error {
	if err := w.ensureCreated(); err != nil {
		return err
	}
	r, err := parseCancelReason(reason)
	if err != nil {
		return err
	}
	return w.p.Cancel(w.ctx, r)
}

// Fail (enum)
func (w *paymentWorld) whenFail(reason string) error {
	if err := w.ensureCreated(); err != nil {
		return err
	}
	r, err := parseFailureReason(reason)
	if err != nil {
		return err
	}
	return w.p.Fail(w.ctx, r)
}

func (w *paymentWorld) whenTryRefund(s string) error {
	if err := w.ensureCreated(); err != nil {
		return err
	}
	m, err := parseMoney(s)
	if err != nil {
		return err
	}
	full, e := w.p.Refund(w.ctx, m)
	if e == nil {
		return fmt.Errorf("expected error, got nil")
	}
	w.lastErr = e
	w.lastFull = &full
	return nil
}

// Try-cancel (enum)
func (w *paymentWorld) whenTryCancel(reason string) error {
	if err := w.ensureCreated(); err != nil {
		return err
	}
	r, err := parseCancelReason(reason)
	if err != nil {
		return err
	}
	err = w.p.Cancel(w.ctx, r)
	if err == nil {
		return fmt.Errorf("expected error, got nil")
	}
	w.lastErr = err
	return nil
}

// ---- assertions ----

func (w *paymentWorld) thenStateMustBe(expected string) error {
	if err := w.ensureCreated(); err != nil {
		return err
	}
	got := enumState(w.p.State())
	if got != strings.ToUpper(expected) {
		return fmt.Errorf("state mismatch: got %s, want %s", got, expected)
	}
	return nil
}

func (w *paymentWorld) thenCapturedTotalEquals(s string) error {
	want, err := parseMoney(s)
	if err != nil {
		return err
	}
	got := w.p.Ledger.Captured
	if !moneyEq(want, got) {
		return fmt.Errorf("captured mismatch: got %s %d.%09d, want %s %d.%09d",
			got.GetCurrencyCode(), got.GetUnits(), got.GetNanos(),
			want.GetCurrencyCode(), want.GetUnits(), want.GetNanos())
	}
	return nil
}

func (w *paymentWorld) thenAuthorizedTotalEquals(s string) error {
	want, err := parseMoney(s)
	if err != nil {
		return err
	}
	got := w.p.Ledger.Authorized
	if !moneyEq(want, got) {
		return fmt.Errorf("authorized mismatch: got %s %d.%09d, want %s %d.%09d",
			got.GetCurrencyCode(), got.GetUnits(), got.GetNanos(),
			want.GetCurrencyCode(), want.GetUnits(), want.GetNanos())
	}
	return nil
}

func (w *paymentWorld) thenTotalRefundedEquals(s string) error {
	want, err := parseMoney(s)
	if err != nil {
		return err
	}
	got := w.p.Ledger.TotalRefunded
	if !moneyEq(want, got) {
		return fmt.Errorf("total_refunded mismatch: got %s %d.%09d, want %s %d.%09d",
			got.GetCurrencyCode(), got.GetUnits(), got.GetNanos(),
			want.GetCurrencyCode(), want.GetUnits(), want.GetNanos())
	}
	return nil
}

func (w *paymentWorld) thenUncommittedEventsIncludeInOrder(table *godog.Table) error {
	evs := w.p.UncommittedEvents()
	if len(table.Rows) == 0 {
		return fmt.Errorf("empty table")
	}
	want := make([]string, 0, len(table.Rows))
	for _, r := range table.Rows {
		if len(r.Cells) != 1 {
			return fmt.Errorf("table must be one column with type names")
		}
		want = append(want, strings.TrimSpace(r.Cells[0].Value))
	}
	if len(evs) != len(want) {
		return fmt.Errorf("events length mismatch: got %d, want %d", len(evs), len(want))
	}
	for i, e := range evs {
		got := eventTypeName(e)
		if got != want[i] {
			return fmt.Errorf("event[%d] mismatch: got %s, want %s", i, got, want[i])
		}
	}
	return nil
}

func (w *paymentWorld) thenLastUncommittedEventIs(name string) error {
	evs := w.p.UncommittedEvents()
	if len(evs) == 0 {
		return fmt.Errorf("no events")
	}
	got := eventTypeName(evs[len(evs)-1])
	if got != name {
		return fmt.Errorf("last event mismatch: got %s, want %s", got, name)
	}
	return nil
}

func (w *paymentWorld) thenOperationMustBeRejected() error {
	if w.lastErr == nil {
		return fmt.Errorf("expected operation error, got nil")
	}
	return nil
}

func (w *paymentWorld) thenFullRefundFlagIs(val string) error {
	want := strings.EqualFold(val, "true") || strings.EqualFold(val, "yes")
	if w.lastFull != nil {
		if *w.lastFull != want {
			return fmt.Errorf("full-refund flag mismatch: got %v, want %v", *w.lastFull, want)
		}
		return nil
	}
	evs := w.p.UncommittedEvents()
	for i := len(evs) - 1; i >= 0; i-- {
		if rr, ok := evs[i].(*eventv1.PaymentRefunded); ok {
			if rr.GetFull() != want {
				return fmt.Errorf("full-refund flag mismatch: got %v, want %v", rr.GetFull(), want)
			}
			return nil
		}
	}
	return fmt.Errorf("no PaymentRefunded event found")
}

// Alias: "the payment state must still be" → same as "the payment state must be"
func (w *paymentWorld) thenStateMustStillBe(expected string) error {
	return w.thenStateMustBe(expected)
}

// ---- godog wiring ----

func InitializeScenario(sc *godog.ScenarioContext) {
	var w paymentWorld

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		w.reset(ctx)
		return ctx, nil
	})

	// Given / And (setup)
	sc.Step(`^a payment "([^"]+)" is created for invoice "([^"]+)"$`, w.givenPaymentCreatedForInvoice)
	sc.Step(`^the amount is "([^"]+)"$`, w.andAmountIs)
	sc.Step(`^the payment kind is "([^"]+)"$`, w.andKindIs)
	sc.Step(`^the capture mode is "([^"]+)"$`, w.andCaptureModeIs)

	// When (commands)
	sc.Step(`^I require SCA$`, w.whenRequireSCA)
	sc.Step(`^I confirm authorization of "([^"]+)"$`, w.whenConfirmAuthorizationOf)
	sc.Step(`^I authorize "([^"]+)"$`, w.whenAuthorize)
	sc.Step(`^I try to authorize "([^"]+)"$`, w.whenTryAuthorize)
	sc.Step(`^I capture "([^"]+)"$`, w.whenCapture)
	sc.Step(`^I try to capture "([^"]+)"$`, w.whenTryCapture)
	sc.Step(`^I refund "([^"]+)"$`, w.whenRefund)
	sc.Step(`^a refund attempt fails with reason "([^"]+)"$`, w.whenRefundFailedWithReason)
	sc.Step(`^I cancel the payment with reason "([^"]+)"$`, w.whenCancel)
	sc.Step(`^I fail the payment with reason "([^"]+)"$`, w.whenFail)
	// message parameter (if present in features) is ignored by aggregate;
	// keep signature for compatibility:
	sc.Step(`^I fail the payment with reason "([^"]+)" and message "([^"]+)"$`, func(reason, _ string) error {
		return w.whenFail(reason)
	})
	sc.Step(`^I try to refund "([^"]+)"$`, w.whenTryRefund)
	sc.Step(`^I try to cancel the payment with reason "([^"]+)"$`, w.whenTryCancel)

	// Then (assertions)
	sc.Step(`^the payment state must be "([^"]+)"$`, w.thenStateMustBe)
	sc.Step(`^the captured total equals "([^"]+)"$`, w.thenCapturedTotalEquals)
	sc.Step(`^the authorized total equals "([^"]+)"$`, w.thenAuthorizedTotalEquals)
	sc.Step(`^the total refunded equals "([^"]+)"$`, w.thenTotalRefundedEquals)
	sc.Step(`^the uncommitted events include, in order:$`, w.thenUncommittedEventsIncludeInOrder)
	sc.Step(`^the last uncommitted event is "([^"]+)"$`, w.thenLastUncommittedEventIs)
	sc.Step(`^the operation must be rejected$`, w.thenOperationMustBeRejected)
	sc.Step(`^full refund flag is "([^"]+)"$`, w.thenFullRefundFlagIs)
	sc.Step(`^the payment state must still be "([^"]+)"$`, w.thenStateMustStillBe)
}
