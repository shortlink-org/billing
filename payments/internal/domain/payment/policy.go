// internal/domain/payment/policy.go
package payment

import (
	eventv1 "github.com/shortlink-org/billing/payments/internal/domain/event/v1"
)

// Policy defines business rules independent from process state.
type Policy interface {
	// AllowImmediateCapture indicates whether capture is allowed directly from CREATED (without auth).
	AllowImmediateCapture(kind eventv1.PaymentKind, mode eventv1.CaptureMode) bool
	// IsCurrencySupported tells if currency is allowed.
	IsCurrencySupported(code string) bool
	// ShouldRequireSCA allows forcing SCA at creation time (can be extended by amount/region/etc).
	ShouldRequireSCA(kind eventv1.PaymentKind, mode eventv1.CaptureMode) bool
}

// StaticPolicy is a simple default implementation.
type StaticPolicy struct {
	SupportedCurrencies map[string]struct{} // nil => allow all
	ForceSCA            bool                // if true — always require SCA
}

func (p *StaticPolicy) AllowImmediateCapture(_ eventv1.PaymentKind, mode eventv1.CaptureMode) bool {
	return mode == eventv1.CaptureMode_CAPTURE_MODE_IMMEDIATE
}

func (p *StaticPolicy) IsCurrencySupported(code string) bool {
	if p.SupportedCurrencies == nil {
		return true
	}
	_, ok := p.SupportedCurrencies[code]
	return ok
}

func (p *StaticPolicy) ShouldRequireSCA(_ eventv1.PaymentKind, _ eventv1.CaptureMode) bool {
	return p.ForceSCA
}

var defaultPolicy = &StaticPolicy{}
