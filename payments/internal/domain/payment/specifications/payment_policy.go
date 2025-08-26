package specifications

import (
	"github.com/shortlink-org/go-sdk/specification"
	"google.golang.org/genproto/googleapis/type/money"

	eventv1 "github.com/shortlink-org/billing/payments/internal/domain/event/v1"
	"github.com/shortlink-org/billing/payments/internal/domain/payment"
)

// SpecificationBasedPolicy implements the Policy interface using specifications
type SpecificationBasedPolicy struct {
	supportedCurrencies     map[string]struct{}
	forceSCA               bool
	minAmountForSCA        *money.Money
	highRiskCurrencies     []string
	allowImmediateCapture  map[eventv1.CaptureMode]bool
}

// NewSpecificationBasedPolicy creates a new policy that uses specifications
func NewSpecificationBasedPolicy(options ...PolicyOption) *SpecificationBasedPolicy {
	policy := &SpecificationBasedPolicy{
		allowImmediateCapture: map[eventv1.CaptureMode]bool{
			eventv1.CaptureMode_CAPTURE_MODE_IMMEDIATE: true,
		},
	}
	
	for _, opt := range options {
		opt(policy)
	}
	
	return policy
}

// PolicyOption allows configuring the specification-based policy
type PolicyOption func(*SpecificationBasedPolicy)

// WithSupportedCurrencies sets the supported currencies
func WithSupportedCurrencies(currencies ...string) PolicyOption {
	return func(p *SpecificationBasedPolicy) {
		p.supportedCurrencies = make(map[string]struct{})
		for _, currency := range currencies {
			p.supportedCurrencies[currency] = struct{}{}
		}
	}
}

// WithForceSCA enables forced SCA for all payments
func WithForceSCA(force bool) PolicyOption {
	return func(p *SpecificationBasedPolicy) {
		p.forceSCA = force
	}
}

// WithMinAmountForSCA sets the minimum amount that requires SCA
func WithMinAmountForSCA(amount *money.Money) PolicyOption {
	return func(p *SpecificationBasedPolicy) {
		p.minAmountForSCA = amount
	}
}

// WithHighRiskCurrencies sets currencies that require fraud checking
func WithHighRiskCurrencies(currencies ...string) PolicyOption {
	return func(p *SpecificationBasedPolicy) {
		p.highRiskCurrencies = currencies
	}
}

// AllowImmediateCapture uses specifications to determine if immediate capture is allowed
func (p *SpecificationBasedPolicy) AllowImmediateCapture(kind eventv1.PaymentKind, mode eventv1.CaptureMode) bool {
	return p.allowImmediateCapture[mode]
}

// IsCurrencySupported uses a specification to check currency support
func (p *SpecificationBasedPolicy) IsCurrencySupported(code string) bool {
	if p.supportedCurrencies == nil {
		return true // All currencies supported if none specified
	}
	_, ok := p.supportedCurrencies[code]
	return ok
}

// ShouldRequireSCA uses specifications to determine if SCA is required
func (p *SpecificationBasedPolicy) ShouldRequireSCA(kind eventv1.PaymentKind, mode eventv1.CaptureMode) bool {
	if p.forceSCA {
		return true
	}
	
	// This method would typically need access to the payment amount to make a decision
	// In a real implementation, you might need to restructure this to take a payment parameter
	return false
}

// ValidatePaymentForCreation validates a payment using specifications
func (p *SpecificationBasedPolicy) ValidatePaymentForCreation(payment *payment.Payment) error {
	// Build currency list for validation
	var currencies []string
	if p.supportedCurrencies != nil {
		currencies = make([]string, 0, len(p.supportedCurrencies))
		for currency := range p.supportedCurrencies {
			currencies = append(currencies, currency)
		}
	}
	
	// Use composite specification for validation
	spec := NewValidPaymentForCreation(currencies...)
	return spec.IsSatisfiedBy(payment)
}

// ValidatePaymentForAuthorization validates a payment can be authorized
func (p *SpecificationBasedPolicy) ValidatePaymentForAuthorization(payment *payment.Payment) error {
	spec := NewPaymentCanBeAuthorized()
	return spec.IsSatisfiedBy(payment)
}

// ValidatePaymentForCapture validates a payment can be captured
func (p *SpecificationBasedPolicy) ValidatePaymentForCapture(payment *payment.Payment) error {
	spec := NewPaymentCanBeCaptured()
	return spec.IsSatisfiedBy(payment)
}

// ValidatePaymentForRefund validates a payment can be refunded
func (p *SpecificationBasedPolicy) ValidatePaymentForRefund(payment *payment.Payment) error {
	spec := NewPaymentCanBeRefunded()
	return spec.IsSatisfiedBy(payment)
}

// RequiresFraudCheck determines if a payment requires fraud checking
func (p *SpecificationBasedPolicy) RequiresFraudCheck(payment *payment.Payment) bool {
	spec := NewPaymentForFraudCheck(p.minAmountForSCA, p.highRiskCurrencies...)
	err := spec.IsSatisfiedBy(payment)
	return err == nil // No error means fraud check is required
}

// FilterPaymentsBySpecification filters a slice of payments using a specification
func FilterPaymentsBySpecification(payments []*payment.Payment, spec specification.Specification[payment.Payment]) ([]*payment.Payment, error) {
	return specification.Filter(payments, spec)
}

// Example usage functions

// GetPaymentsReadyForCapture returns payments that can be captured
func GetPaymentsReadyForCapture(payments []*payment.Payment) ([]*payment.Payment, error) {
	spec := NewPaymentCanBeCaptured()
	return FilterPaymentsBySpecification(payments, spec)
}

// GetPaymentsRequiringFraudCheck returns payments that need fraud checking
func GetPaymentsRequiringFraudCheck(payments []*payment.Payment, policy *SpecificationBasedPolicy) ([]*payment.Payment, error) {
	spec := NewPaymentForFraudCheck(policy.minAmountForSCA, policy.highRiskCurrencies...)
	return FilterPaymentsBySpecification(payments, spec)
}

// GetValidPaymentsForRefund returns payments that can be refunded
func GetValidPaymentsForRefund(payments []*payment.Payment) ([]*payment.Payment, error) {
	spec := NewPaymentCanBeRefunded()
	return FilterPaymentsBySpecification(payments, spec)
}