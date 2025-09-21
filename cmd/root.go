package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
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
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		if rootPersistentFlags.verbose {
			log.Logger.Level(zerolog.DebugLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		cfg := &server.Config{Root: rootFlags.root, Port: rootFlags.port, Logger: log.Logger}
		srv := server.New(cfg)
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		if err := srv.Start(ctx); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				log.Info().Msg("server stopped (context canceled)")
				return
			}
			log.Fatal().Err(err).Msg("server exited")
		}
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
