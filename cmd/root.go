package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	hookcmd "github.com/yz4230/githost-poc/cmd/hook"
)

var rootPersistentFlags struct {
	verbose bool
}

var rootCmd = &cobra.Command{
	Use: "githost",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
		if rootPersistentFlags.verbose {
			log.Logger.Level(zerolog.DebugLevel)
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
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(hookcmd.HookCmd)
}
