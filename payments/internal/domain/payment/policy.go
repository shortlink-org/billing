package payment

import (
	eventv1 "github.com/shortlink-org/billing/payments/internal/domain/event/v1"
)

// Policy определяет бизнес-правила, не зависящие от состояния процесса.
type Policy interface {
	// Разрешена ли мгновенная оплата из CREATED (без AUTHORIZED).
	AllowImmediateCapture(kind eventv1.PaymentKind, mode eventv1.CaptureMode) bool
	// Поддерживается ли валюта.
	IsCurrencySupported(code string) bool
	// Нужна ли SCA при создании (можно расширять по сумме/стране и т.п.).
	ShouldRequireSCA(kind eventv1.PaymentKind, mode eventv1.CaptureMode) bool
}

// StaticPolicy — простой вариант (по умолчанию).
type StaticPolicy struct {
	SupportedCurrencies map[string]struct{} // nil => разрешить все
	ForceSCA            bool                // если true — всегда требовать SCA
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
