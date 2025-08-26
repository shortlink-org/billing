package rules

import (
	"errors"

	"github.com/google/uuid"
	"github.com/shortlink-org/go-sdk/specification"

	v1 "github.com/shortlink-org/billing/billing/internal/domain/account/v1"
)

var ErrTariffIdRequired = errors.New("tariffId is required")

// TariffId implements the specification pattern for validating tariff ID requirements
type TariffId struct{}

// NewTariffId creates a new tariff ID specification
func NewTariffId() *TariffId {
	return &TariffId{}
}

// IsSatisfiedBy validates that the account has a valid tariff ID
func (t *TariffId) IsSatisfiedBy(account *v1.Account) error {
	if account.GetTariffId() != uuid.Nil {
		return nil
	}

	return ErrTariffIdRequired
}

// Ensure TariffId implements the Specification interface
var _ specification.Specification[v1.Account] = (*TariffId)(nil)
