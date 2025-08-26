package stripeadp

import (
	"errors"

	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v82"
)

var (
	// ErrMissingAPIKey is returned when STRIPE_API_KEY is not set.
	ErrMissingAPIKey = errors.New("stripe: missing STRIPE_API_KEY")
)

// Provider implements PaymentProvider interface for Stripe.
type Provider struct {
	client *stripe.Client
}

// New creates a Stripe client using STRIPE_API_KEY from env.
// Example: export STRIPE_API_KEY=sk_test_123...
func New() (*Provider, error) {
	viper.AutomaticEnv()

	apiKey := viper.GetString("STRIPE_API_KEY")
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	client := stripe.NewClient(apiKey)

	return &Provider{client: client}, nil
}
