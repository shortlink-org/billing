package specifications

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shortlink-org/go-sdk/specification"
	"google.golang.org/genproto/googleapis/type/money"

	eventv1 "github.com/shortlink-org/billing/payments/internal/domain/event/v1"
	flowv1 "github.com/shortlink-org/billing/payments/internal/domain/flow/v1"
	"github.com/shortlink-org/billing/payments/internal/domain/payment"
)

// Domain-specific errors
var (
	ErrPaymentIdRequired         = errors.New("payment ID is required")
	ErrInvoiceIdRequired         = errors.New("invoice ID is required")
	ErrAmountRequired           = errors.New("payment amount is required")
	ErrInvalidAmount            = errors.New("payment amount must be positive")
	ErrUnsupportedCurrency      = errors.New("currency is not supported")
	ErrPaymentNotInCreatedState = errors.New("payment is not in created state")
	ErrPaymentNotAuthorized     = errors.New("payment is not authorized")
	ErrPaymentNotCapturable     = errors.New("payment is not in a capturable state")
	ErrPaymentNotRefundable     = errors.New("payment is not in a refundable state")
	ErrInvalidCaptureMode       = errors.New("invalid capture mode")
	ErrInvalidPaymentKind       = errors.New("invalid payment kind")
)

// PaymentIdRequired validates that the payment has a valid ID
type PaymentIdRequired struct{}

func NewPaymentIdRequired() *PaymentIdRequired {
	return &PaymentIdRequired{}
}

func (s *PaymentIdRequired) IsSatisfiedBy(p *payment.Payment) error {
	if p.ID() != uuid.Nil {
		return nil
	}
	return ErrPaymentIdRequired
}

// InvoiceIdRequired validates that the payment has a valid invoice ID
type InvoiceIdRequired struct{}

func NewInvoiceIdRequired() *InvoiceIdRequired {
	return &InvoiceIdRequired{}
}

func (s *InvoiceIdRequired) IsSatisfiedBy(p *payment.Payment) error {
	if p.InvoiceID() != uuid.Nil {
		return nil
	}
	return ErrInvoiceIdRequired
}

// AmountRequired validates that the payment has a valid amount
type AmountRequired struct{}

func NewAmountRequired() *AmountRequired {
	return &AmountRequired{}
}

func (s *AmountRequired) IsSatisfiedBy(p *payment.Payment) error {
	if p.Ledger.Amount == nil {
		return ErrAmountRequired
	}
	return nil
}

// AmountPositive validates that the payment amount is positive
type AmountPositive struct{}

func NewAmountPositive() *AmountPositive {
	return &AmountPositive{}
}

func (s *AmountPositive) IsSatisfiedBy(p *payment.Payment) error {
	if p.Ledger.Amount == nil {
		return ErrAmountRequired
	}
	
	if p.Ledger.Amount.Units <= 0 && p.Ledger.Amount.Nanos <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

// CurrencySupported validates that the payment currency is supported
type CurrencySupported struct {
	SupportedCurrencies map[string]struct{}
}

func NewCurrencySupported(currencies ...string) *CurrencySupported {
	supported := make(map[string]struct{})
	for _, currency := range currencies {
		supported[currency] = struct{}{}
	}
	return &CurrencySupported{SupportedCurrencies: supported}
}

func (s *CurrencySupported) IsSatisfiedBy(p *payment.Payment) error {
	if p.Ledger.Amount == nil {
		return ErrAmountRequired
	}
	
	if len(s.SupportedCurrencies) == 0 {
		return nil // All currencies supported if none specified
	}
	
	_, ok := s.SupportedCurrencies[p.Ledger.Amount.CurrencyCode]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnsupportedCurrency, p.Ledger.Amount.CurrencyCode)
	}
	return nil
}

// PaymentInState validates that the payment is in a specific state
type PaymentInState struct {
	RequiredState flowv1.PaymentFlow
}

func NewPaymentInState(state flowv1.PaymentFlow) *PaymentInState {
	return &PaymentInState{RequiredState: state}
}

func (s *PaymentInState) IsSatisfiedBy(p *payment.Payment) error {
	if p.State() == s.RequiredState {
		return nil
	}
	return fmt.Errorf("payment is in state %v, expected %v", p.State(), s.RequiredState)
}

// PaymentKindAllowed validates that the payment kind is allowed
type PaymentKindAllowed struct {
	AllowedKinds []eventv1.PaymentKind
}

func NewPaymentKindAllowed(kinds ...eventv1.PaymentKind) *PaymentKindAllowed {
	return &PaymentKindAllowed{AllowedKinds: kinds}
}

func (s *PaymentKindAllowed) IsSatisfiedBy(p *payment.Payment) error {
	// Note: Payment aggregate doesn't expose Kind() method publicly
	// This specification would need to be used with payment creation data
	// or the Payment struct would need to expose the kind field
	for _, kind := range s.AllowedKinds {
		// For now, we'll assume any kind is allowed if not specifically restricted
		_ = kind
	}
	return nil // TODO: Implement when Payment exposes Kind() method
}

// CaptureModeAllowed validates that the capture mode is allowed
type CaptureModeAllowed struct {
	AllowedModes []eventv1.CaptureMode
}

func NewCaptureModeAllowed(modes ...eventv1.CaptureMode) *CaptureModeAllowed {
	return &CaptureModeAllowed{AllowedModes: modes}
}

func (s *CaptureModeAllowed) IsSatisfiedBy(p *payment.Payment) error {
	// Note: Payment aggregate doesn't expose CaptureMode() method publicly
	// This specification would need to be used with payment creation data
	// or the Payment struct would need to expose the captureMode field
	for _, mode := range s.AllowedModes {
		// For now, we'll assume any mode is allowed if not specifically restricted
		_ = mode
	}
	return nil // TODO: Implement when Payment exposes CaptureMode() method
}

// AmountWithinRange validates that the payment amount is within a specific range
type AmountWithinRange struct {
	MinAmount *money.Money
	MaxAmount *money.Money
}

func NewAmountWithinRange(min, max *money.Money) *AmountWithinRange {
	return &AmountWithinRange{MinAmount: min, MaxAmount: max}
}

func (s *AmountWithinRange) IsSatisfiedBy(p *payment.Payment) error {
	if p.Ledger.Amount == nil {
		return ErrAmountRequired
	}
	
	amount := p.Ledger.Amount
	
	// Check minimum amount
	if s.MinAmount != nil && amount.CurrencyCode == s.MinAmount.CurrencyCode {
		if amount.Units < s.MinAmount.Units || 
		   (amount.Units == s.MinAmount.Units && amount.Nanos < s.MinAmount.Nanos) {
			return fmt.Errorf("amount %v %s is below minimum %v %s", 
				amount.Units, amount.CurrencyCode, s.MinAmount.Units, s.MinAmount.CurrencyCode)
		}
	}
	
	// Check maximum amount
	if s.MaxAmount != nil && amount.CurrencyCode == s.MaxAmount.CurrencyCode {
		if amount.Units > s.MaxAmount.Units || 
		   (amount.Units == s.MaxAmount.Units && amount.Nanos > s.MaxAmount.Nanos) {
			return fmt.Errorf("amount %v %s exceeds maximum %v %s", 
				amount.Units, amount.CurrencyCode, s.MaxAmount.Units, s.MaxAmount.CurrencyCode)
		}
	}
	
	return nil
}

// Composite Specifications

// ValidPaymentForCreation combines multiple validations for payment creation
type ValidPaymentForCreation struct {
	*specification.AndSpecification[payment.Payment]
}

func NewValidPaymentForCreation(supportedCurrencies ...string) *ValidPaymentForCreation {
	andSpec := specification.NewAndSpecification[payment.Payment](
		NewPaymentIdRequired(),
		NewInvoiceIdRequired(),
		NewAmountRequired(),
		NewAmountPositive(),
		NewCurrencySupported(supportedCurrencies...),
		NewPaymentInState(flowv1.PaymentFlow_PAYMENT_FLOW_CREATED),
	)
	
	return &ValidPaymentForCreation{
		AndSpecification: andSpec,
	}
}

// PaymentCanBeAuthorized validates that a payment can be authorized
type PaymentCanBeAuthorized struct {
	*specification.AndSpecification[payment.Payment]
}

func NewPaymentCanBeAuthorized() *PaymentCanBeAuthorized {
	andSpec := specification.NewAndSpecification[payment.Payment](
		NewPaymentInState(flowv1.PaymentFlow_PAYMENT_FLOW_CREATED),
		NewAmountPositive(),
	)
	
	return &PaymentCanBeAuthorized{
		AndSpecification: andSpec,
	}
}

// PaymentCanBeCaptured validates that a payment can be captured
type PaymentCanBeCaptured struct {
	*specification.OrSpecification[payment.Payment]
}

func NewPaymentCanBeCaptured() *PaymentCanBeCaptured {
	// Payment can be captured if it's authorized OR if it allows immediate capture
	authorizedSpec := NewPaymentInState(flowv1.PaymentFlow_PAYMENT_FLOW_AUTHORIZED)
	createdSpec := NewPaymentInState(flowv1.PaymentFlow_PAYMENT_FLOW_CREATED)
	
	orSpec := specification.NewOrSpecification[payment.Payment](
		authorizedSpec,
		createdSpec, // Allow capture from created state (policy will validate capture mode)
	)
	
	return &PaymentCanBeCaptured{
		OrSpecification: orSpec,
	}
}

// PaymentCanBeRefunded validates that a payment can be refunded
type PaymentCanBeRefunded struct {
	*specification.AndSpecification[payment.Payment]
}

func NewPaymentCanBeRefunded() *PaymentCanBeRefunded {
	andSpec := specification.NewAndSpecification[payment.Payment](
		NewPaymentInState(flowv1.PaymentFlow_PAYMENT_FLOW_PAID), // Use PAID state instead of CAPTURED
		NewAmountPositive(),
	)
	
	return &PaymentCanBeRefunded{
		AndSpecification: andSpec,
	}
}

// PaymentForFraudCheck validates payment conditions that require fraud checking
type PaymentForFraudCheck struct {
	*specification.OrSpecification[payment.Payment]
}

func NewPaymentForFraudCheck(highRiskAmount *money.Money, highRiskCurrencies ...string) *PaymentForFraudCheck {
	// Create specifications for high-risk conditions
	var specs []specification.Specification[payment.Payment]
	
	// High amount
	if highRiskAmount != nil {
		specs = append(specs, &AmountAboveThreshold{Threshold: highRiskAmount})
	}
	
	// High-risk currencies
	if len(highRiskCurrencies) > 0 {
		highRiskCurrencyMap := make(map[string]struct{})
		for _, currency := range highRiskCurrencies {
			highRiskCurrencyMap[currency] = struct{}{}
		}
		specs = append(specs, &CurrencyInList{Currencies: highRiskCurrencyMap})
	}
	
	orSpec := specification.NewOrSpecification[payment.Payment](specs...)
	
	return &PaymentForFraudCheck{
		OrSpecification: orSpec,
	}
}

// Helper specifications for fraud checking

// AmountAboveThreshold checks if payment amount is above a threshold
type AmountAboveThreshold struct {
	Threshold *money.Money
}

func (s *AmountAboveThreshold) IsSatisfiedBy(p *payment.Payment) error {
	if p.Ledger.Amount == nil {
		return ErrAmountRequired
	}
	
	amount := p.Ledger.Amount
	if amount.CurrencyCode == s.Threshold.CurrencyCode {
		if amount.Units > s.Threshold.Units || 
		   (amount.Units == s.Threshold.Units && amount.Nanos > s.Threshold.Nanos) {
			return nil // Amount is above threshold, fraud check needed
		}
	}
	
	return errors.New("amount is below threshold")
}

// CurrencyInList checks if payment currency is in a specific list
type CurrencyInList struct {
	Currencies map[string]struct{}
}

func (s *CurrencyInList) IsSatisfiedBy(p *payment.Payment) error {
	if p.Ledger.Amount == nil {
		return ErrAmountRequired
	}
	
	_, ok := s.Currencies[p.Ledger.Amount.CurrencyCode]
	if ok {
		return nil // Currency is in the list
	}
	
	return errors.New("currency is not in the specified list")
}

// Ensure all specifications implement the Specification interface
var (
	_ specification.Specification[payment.Payment] = (*PaymentIdRequired)(nil)
	_ specification.Specification[payment.Payment] = (*InvoiceIdRequired)(nil)
	_ specification.Specification[payment.Payment] = (*AmountRequired)(nil)
	_ specification.Specification[payment.Payment] = (*AmountPositive)(nil)
	_ specification.Specification[payment.Payment] = (*CurrencySupported)(nil)
	_ specification.Specification[payment.Payment] = (*PaymentInState)(nil)
	_ specification.Specification[payment.Payment] = (*PaymentKindAllowed)(nil)
	_ specification.Specification[payment.Payment] = (*CaptureModeAllowed)(nil)
	_ specification.Specification[payment.Payment] = (*AmountWithinRange)(nil)
	_ specification.Specification[payment.Payment] = (*ValidPaymentForCreation)(nil)
	_ specification.Specification[payment.Payment] = (*PaymentCanBeAuthorized)(nil)
	_ specification.Specification[payment.Payment] = (*PaymentCanBeCaptured)(nil)
	_ specification.Specification[payment.Payment] = (*PaymentCanBeRefunded)(nil)
	_ specification.Specification[payment.Payment] = (*PaymentForFraudCheck)(nil)
	_ specification.Specification[payment.Payment] = (*AmountAboveThreshold)(nil)
	_ specification.Specification[payment.Payment] = (*CurrencyInList)(nil)
)