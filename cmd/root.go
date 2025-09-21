package cmd

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/yz4230/githost-poc/internal/server"
)

var rootPersistentFlags struct {
	verbose bool
}

var rootFlags struct {
	port int
	root string
}

var rootCmd = &cobra.Command{
	Use: "githost",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
		if rootPersistentFlags.verbose {
			log.Logger.Level(zerolog.DebugLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		cfg := &server.Config{Root: rootFlags.root, Port: rootFlags.port, Logger: log.Logger}
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

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootPersistentFlags.verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().IntVarP(&rootFlags.port, "port", "p", 8080, "Port to listen on")
	rootCmd.Flags().StringVarP(&rootFlags.root, "root", "r", "./repos", "Root directory to store repositories")
}
