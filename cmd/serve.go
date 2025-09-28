package cmd

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/yz4230/githost-poc/internal/server"
)

var serveFlags struct {
	port    int
	dataDir string
}

var serveCmd = &cobra.Command{
	Use: "serve",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := &server.Config{Root: serveFlags.dataDir, Port: serveFlags.port, Logger: log.Logger}
		srv := server.New(config)
		chSignal := make(chan os.Signal, 1)
		signal.Notify(chSignal, os.Interrupt, syscall.SIGTERM)

		wg := &sync.WaitGroup{}
		wg.Go(func() {
			if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				config.Logger.Fatal().Err(err).Msg("server error")
			}
		})

		sig := <-chSignal
		config.Logger.Info().Str("signal", sig.String()).Msg("shutting down server...")
		if err := srv.Stop(context.Background()); err != nil {
			config.Logger.Error().Err(err).Msg("error during server shutdown")
		}

		wg.Wait()
		config.Logger.Info().Msg("server stopped")

		return nil
	},
}

func init() {
	serveCmd.Flags().IntVarP(&serveFlags.port, "port", "p", 8080, "Port to listen on")
	serveCmd.Flags().StringVarP(&serveFlags.dataDir, "data", "d", "./data", "Directory to store server data")
}
