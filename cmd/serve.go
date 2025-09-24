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
	port int
	root string
}

var serveCmd = &cobra.Command{
	Use: "serve",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := &server.Config{Root: serveFlags.root, Port: serveFlags.port, Logger: log.Logger}
		srv := server.New(cfg)
		chSignal := make(chan os.Signal, 1)
		signal.Notify(chSignal, os.Interrupt, syscall.SIGTERM)

		wg := &sync.WaitGroup{}
		wg.Go(func() {
			if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				cfg.Logger.Fatal().Err(err).Msg("server error")
			}
		})

		sig := <-chSignal
		cfg.Logger.Info().Str("signal", sig.String()).Msg("shutting down server...")
		if err := srv.Stop(context.Background()); err != nil {
			cfg.Logger.Error().Err(err).Msg("error during server shutdown")
		}

		wg.Wait()
		cfg.Logger.Info().Msg("server stopped")
	},
}

func init() {
	serveCmd.Flags().IntVarP(&serveFlags.port, "port", "p", 8080, "Port to listen on")
	serveCmd.Flags().StringVarP(&serveFlags.root, "root", "r", "./repos", "Root directory to store repositories")
}
