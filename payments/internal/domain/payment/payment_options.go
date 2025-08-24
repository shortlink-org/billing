package payment

type Option func(*Payment)

func WithPolicy(pol Policy) Option {
	return func(p *Payment) { p.policy = pol }
}
