package tinkoffadp

import (
	"crypto/tls"
	"errors"
	"net/http"

	"github.com/spf13/viper"
)

var (
	// ErrMissingAPIKey is returned when TINKOFF_API_KEY is not set.
	ErrMissingAPIKey = errors.New("tinkoff: missing TINKOFF_API_KEY")
	// ErrMissingCertFile is returned when TINKOFF_CERT_FILE is not set.
	ErrMissingCertFile = errors.New("tinkoff: missing TINKOFF_CERT_FILE")
	// ErrMissingKeyFile is returned when TINKOFF_KEY_FILE is not set.
	ErrMissingKeyFile = errors.New("tinkoff: missing TINKOFF_KEY_FILE")
)

// Provider implements PaymentProvider interface for Tinkoff.
type Provider struct {
	client   *http.Client
	apiKey   string
	baseURL  string
	certFile string
	keyFile  string
}

// New creates a Tinkoff client using environment variables.
// Required environment variables:
// - TINKOFF_API_KEY: API key for authentication
// - TINKOFF_CERT_FILE: Path to client certificate file (.pem)
// - TINKOFF_KEY_FILE: Path to client private key file (.key)
// Optional:
// - TINKOFF_BASE_URL: Base URL (defaults to https://secured-openapi.tbank.ru)
func New() (*Provider, error) {
	viper.AutomaticEnv()

	apiKey := viper.GetString("TINKOFF_API_KEY")
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	certFile := viper.GetString("TINKOFF_CERT_FILE")
	if certFile == "" {
		return nil, ErrMissingCertFile
	}

	keyFile := viper.GetString("TINKOFF_KEY_FILE")
	if keyFile == "" {
		return nil, ErrMissingKeyFile
	}

	baseURL := viper.GetString("TINKOFF_BASE_URL")
	if baseURL == "" {
		baseURL = "https://secured-openapi.tbank.ru"
	}

	// Load client certificate
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// Configure TLS transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	client := &http.Client{
		Transport: transport,
	}

	return &Provider{
		client:   client,
		apiKey:   apiKey,
		baseURL:  baseURL,
		certFile: certFile,
		keyFile:  keyFile,
	}, nil
}
