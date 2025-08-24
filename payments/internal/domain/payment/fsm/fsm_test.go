package fsm_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	flowv1 "github.com/shortlink-org/billing/payments/internal/domain/flow/v1"
	"github.com/shortlink-org/billing/payments/internal/domain/payment/fsm"
)

// setEq compares two slices as sets (order-agnostic).
func setEq(t *testing.T, got, want []string, msg string) {
	t.Helper()
	g := map[string]struct{}{}
	for _, s := range got {
		g[s] = struct{}{}
	}
	w := map[string]struct{}{}
	for _, s := range want {
		w[s] = struct{}{}
	}
	require.Equalf(t, len(w), len(g), "%s: size mismatch\n got: %v\nwant: %v", msg, got, want)
	for s := range w {
		_, ok := g[s]
		require.Truef(t, ok, "%s: missing %q in got=%v", msg, s, got)
	}
}

func TestFSM_AllowedTransitionsMatrix(t *testing.T) {
	cases := []struct {
		state  flowv1.PaymentFlow
		expect []string
	}{
		{
			state: flowv1.PaymentFlow_PAYMENT_FLOW_CREATED,
			expect: []string{
				fsm.EventSCARequired,
				fsm.EventAuthorize,
				fsm.EventCapture, // FSM allows immediate capture; domain Policy may forbid
				fsm.EventCancel,
				fsm.EventFail,
			},
		},
		{
			state:  flowv1.PaymentFlow_PAYMENT_FLOW_WAITING_FOR_CONFIRMATION,
			expect: []string{fsm.EventConfirm, fsm.EventCancel, fsm.EventFail},
		},
		{
			state:  flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED,
			expect: []string{fsm.EventCapture, fsm.EventCancel, fsm.EventFail},
		},
		{
			// IMPORTANT: partial refunds do NOT trigger FSM (state stays PAID),
			// so only full refund is listed as a transition.
			state:  flowv1.PaymentFlow_PAYMENT_FLOW_PAID,
			expect: []string{fsm.EventRefundFull},
		},
		{state: flowv1.PaymentFlow_PAYMENT_FLOW_REFUNDED, expect: []string{}},
		{state: flowv1.PaymentFlow_PAYMENT_FLOW_CANCELED, expect: []string{}},
		{state: flowv1.PaymentFlow_PAYMENT_FLOW_FAILED, expect: []string{}},
	}

	for _, tc := range cases {
		g := fsm.New(tc.state)
		got := g.AllowedTransitions()
		setEq(t, got, tc.expect, "state="+tc.state.String())

		for _, ev := range tc.expect {
			require.Truef(t, g.IsAllowed(ev), "state=%s should allow %q (got %v)",
				tc.state, ev, got)
		}
		require.False(t, g.IsAllowed("nonexistent_event"))
	}
}

func TestFSM_HappyPaths(t *testing.T) {
	ctx := context.Background()

	// CREATED -> WAITING -> AUTHORIZED -> PAID -> REFUNDED (full)
	{
		g := fsm.New(flowv1.PaymentFlow_PAYMENT_FLOW_CREATED)

		require.Contains(t, g.AllowedTransitions(), fsm.EventSCARequired)
		require.NoError(t, g.Trigger(ctx, fsm.EventSCARequired))
		require.Equal(t, flowv1.PaymentFlow_PAYMENT_FLOW_WAITING_FOR_CONFIRMATION.String(), g.Current())

		require.Contains(t, g.AllowedTransitions(), fsm.EventConfirm)
		require.NoError(t, g.Trigger(ctx, fsm.EventConfirm))
		require.Equal(t, flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED.String(), g.Current())

		require.Contains(t, g.AllowedTransitions(), fsm.EventCapture)
		require.NoError(t, g.Trigger(ctx, fsm.EventCapture))
		require.Equal(t, flowv1.PaymentFlow_PAYMENT_FLOW_PAID.String(), g.Current())

		require.Contains(t, g.AllowedTransitions(), fsm.EventRefundFull)
		require.NoError(t, g.Trigger(ctx, fsm.EventRefundFull))
		require.Equal(t, flowv1.PaymentFlow_PAYMENT_FLOW_REFUNDED.String(), g.Current())
	}

	// CREATED -> AUTHORIZED -> PAID (skip SCA)
	{
		g := fsm.New(flowv1.PaymentFlow_PAYMENT_FLOW_CREATED)

		require.NoError(t, g.Trigger(ctx, fsm.EventAuthorize))
		require.Equal(t, flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED.String(), g.Current())

		require.NoError(t, g.Trigger(ctx, fsm.EventCapture))
		require.Equal(t, flowv1.PaymentFlow_PAYMENT_FLOW_PAID.String(), g.Current())
	}

	// CREATED -> PAID (immediate capture) — allowed by FSM; domain Policy decides if permitted.
	{
		g := fsm.New(flowv1.PaymentFlow_PAYMENT_FLOW_CREATED)

		require.NoError(t, g.Trigger(ctx, fsm.EventCapture))
		require.Equal(t, flowv1.PaymentFlow_PAYMENT_FLOW_PAID.String(), g.Current())
	}
}

func TestFSM_InvalidTransitions(t *testing.T) {
	ctx := context.Background()

	// Cannot refund from CREATED
	{
		g := fsm.New(flowv1.PaymentFlow_PAYMENT_FLOW_CREATED)
		require.Error(t, g.Trigger(ctx, fsm.EventRefundFull))
	}

	// Terminal states: no transitions allowed
	terminals := []flowv1.PaymentFlow{
		flowv1.PaymentFlow_PAYMENT_FLOW_REFUNDED,
		flowv1.PaymentFlow_PAYMENT_FLOW_CANCELED,
		flowv1.PaymentFlow_PAYMENT_FLOW_FAILED,
	}
	allEvents := []string{
		fsm.EventSCARequired, fsm.EventAuthorize, fsm.EventConfirm,
		fsm.EventCapture, fsm.EventRefundFull,
		fsm.EventCancel, fsm.EventFail,
	}

	for _, s := range terminals {
		g := fsm.New(s)
		require.Empty(t, g.AllowedTransitions(), "terminal state must expose no transitions: %s", s)
		for _, ev := range allEvents {
			require.Errorf(t, g.Trigger(ctx, ev), "expected error for event=%s from state=%s", ev, s)
		}
	}
}

func TestFSM_ProblemExits(t *testing.T) {
	ctx := context.Background()

	// CREATED -> CANCELED / FAILED
	{
		g := fsm.New(flowv1.PaymentFlow_PAYMENT_FLOW_CREATED)
		require.NoError(t, g.Trigger(ctx, fsm.EventCancel))
		require.Equal(t, flowv1.PaymentFlow_PAYMENT_FLOW_CANCELED.String(), g.Current())
	}
	{
		g := fsm.New(flowv1.PaymentFlow_PAYMENT_FLOW_CREATED)
		require.NoError(t, g.Trigger(ctx, fsm.EventFail))
		require.Equal(t, flowv1.PaymentFlow_PAYMENT_FLOW_FAILED.String(), g.Current())
	}

	// WAITING -> CANCELED / FAILED
	{
		g := fsm.New(flowv1.PaymentFlow_PAYMENT_FLOW_WAITING_FOR_CONFIRMATION)
		require.NoError(t, g.Trigger(ctx, fsm.EventCancel))
		require.Equal(t, flowv1.PaymentFlow_PAYMENT_FLOW_CANCELED.String(), g.Current())
	}
	{
		g := fsm.New(flowv1.PaymentFlow_PAYMENT_FLOW_WAITING_FOR_CONFIRMATION)
		require.NoError(t, g.Trigger(ctx, fsm.EventFail))
		require.Equal(t, flowv1.PaymentFlow_PAYMENT_FLOW_FAILED.String(), g.Current())
	}

	// AUTHORIZED -> CANCELED / FAILED
	{
		g := fsm.New(flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED)
		require.NoError(t, g.Trigger(ctx, fsm.EventCancel))
		require.Equal(t, flowv1.PaymentFlow_PAYMENT_FLOW_CANCELED.String(), g.Current())
	}
	{
		g := fsm.New(flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED)
		require.NoError(t, g.Trigger(ctx, fsm.EventFail))
		require.Equal(t, flowv1.PaymentFlow_PAYMENT_FLOW_FAILED.String(), g.Current())
	}
}
