/*
Wallet application

Wallet-service
*/
package main

import (
	"log/slog"
	"os"

	"github.com/spf13/viper"

	"github.com/shortlink-org/shortlink/pkg/graceful_shutdown"

	wallet_di "github.com/shortlink-org/shortlink/boundaries/billing/wallet/internal/di"
)

func main() {
	viper.SetDefault("SERVICE_NAME", "shortlink-wallet")

	// Init a new service
	service, cleanup, err := wallet_di.InitializeWalletService()
	if err != nil { // TODO: use as helpers
		panic(err)
	}

	defer func() {
		if r := recover(); r != nil {
			service.Log.Error(r.(string)) //nolint:forcetypeassert // simple type assertion
		}
	}()

	// Handle SIGINT, SIGQUIT and SIGTERM.
	signal := graceful_shutdown.GracefulShutdown()

	cleanup()

	service.Log.Info("Service stopped", slog.String("signal", signal.String()))

	// Exit Code 143: Graceful Termination (SIGTERM)
	os.Exit(143) //nolint:gocritic // exit code 143 is used to indicate graceful termination
}
