package stripeadp

import "github.com/stripe/stripe-go/v82"

// Config holds configuration for the Stripe provider.
type Config struct{ 
	APIKey string 
}

// Provider implements PaymentProvider interface for Stripe.
type Provider struct{ 
	c *stripe.Client 
}

// New creates a new Stripe provider with the given configuration.
func New(cfg Config) *Provider {
	return &Provider{c: stripe.NewClient(cfg.APIKey)}
}
