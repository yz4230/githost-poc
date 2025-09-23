package hook

import (
	"bufio"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var postReceiveCmd = &cobra.Command{
	Use: "post-receive",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("post-receive is triggered")

		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			line := s.Text()
			parts := strings.Fields(line)
			if len(parts) != 3 {
				log.Error().Str("line", line).Msg("invalid input line")
			}
			old, new, ref := parts[0], parts[1], parts[2]
			log.Info().Str("ref", ref).
				Str("old", old).
				Str("new", new).
				Msg("received update")

			if ref == "refs/heads/main" {
				dir, err := os.MkdirTemp("", "githost-deploy-*")
				if err != nil {
					log.Error().Err(err).Msg("create temp dir")
					continue
				}
				defer os.RemoveAll(dir)
			}
		}
	},
}
