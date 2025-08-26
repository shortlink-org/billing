package rules

import (
	"errors"

	"github.com/google/uuid"
	"github.com/shortlink-org/go-sdk/specification"

	v1 "github.com/shortlink-org/billing/billing/internal/domain/account/v1"
)

var ErrUserIdRequired = errors.New("userId is required")

// UserId implements the specification pattern for validating user ID requirements
type UserId struct{}

// NewUserId creates a new user ID specification
func NewUserId() *UserId {
	return &UserId{}
}

// IsSatisfiedBy validates that the account has a valid user ID
func (u *UserId) IsSatisfiedBy(account *v1.Account) error {
	if account.GetUserId() != uuid.Nil {
		return nil
	}

	return ErrUserIdRequired
}

// Ensure UserId implements the Specification interface
var _ specification.Specification[v1.Account] = (*UserId)(nil)
