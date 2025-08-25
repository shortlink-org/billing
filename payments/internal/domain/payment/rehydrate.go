package payment

import (
	"google.golang.org/protobuf/proto"
)

// Rehydrate reconstructs a Payment aggregate from a stream of past events.
// It sets sane defaults (policy, FSM) and applies events in order.
//
// Note: events MUST be in causal order starting from PaymentCreated.
func Rehydrate(events []proto.Message) *Payment {
	p := &Payment{
		policy: defaultPolicy, // keep domain rules available after load
	}
	for _, e := range events {
		// apply() updates state/ledger/version and refreshes guard
		p.apply(proto.Clone(e))
	}
	return p
}
