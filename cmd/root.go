package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var rootFlags struct {
	verbose bool
	port    int
	root    string
}

var rootCmd = &cobra.Command{
	Use: "githost",
	Run: func(cmd *cobra.Command, args []string) {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		if rootFlags.verbose {
			log.Logger.Level(zerolog.DebugLevel)
		}

		e := echo.New()
		e.HidePort = true
		e.HideBanner = true
		e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogRemoteIP:  true,
			LogHost:      true,
			LogMethod:    true,
			LogURI:       true,
			LogUserAgent: true,
			LogStatus:    true,
			LogLatency:   true,
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				log.Info().
					Str("remote_ip", v.RemoteIP).
					Str("host", v.Host).
					Str("method", v.Method).
					Str("uri", v.URI).
					Str("user_agent", v.UserAgent).
					Int("status", v.Status).
					Int64("latency_ms", v.Latency.Milliseconds()).
					Msg("Handled request")
				return nil
			},
		}))
		e.Use(middleware.Recover())

		g := e.Group("/:username/:reponame")
		g.GET("/info/refs", func(c echo.Context) error {
			username := c.Param("username")
			reponame := c.Param("reponame")
			repodir := lo.Must(filepath.Abs(filepath.Join(rootFlags.root, username, reponame)))

			if _, err := os.Stat(repodir); os.IsNotExist(err) {
				if err := os.MkdirAll(repodir, 0744); err != nil {
					log.Error().Err(err).Str("dir", repodir).Msg("Failed to create repository directory")
					return c.NoContent(http.StatusInternalServerError)
				}
				if exec.Command("git", "init", "--bare", repodir).Run() != nil {
					log.Error().Err(err).Str("dir", repodir).Msg("Failed to init bare git repo")
					return c.NoContent(http.StatusInternalServerError)
				}
				log.Info().Str("dir", repodir).Msg("Initialized bare git repository")
			}

			service := c.QueryParam("service")

			res := c.Response()
			header := res.Header()
			header.Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", service))
			header.Set("Cache-Control", "no-cache")

			res.Writer.WriteHeader(http.StatusOK)

			{
				var pktline bytes.Buffer
				header := fmt.Sprintf("# service=%s\n", service)
				s := strconv.FormatInt(int64(len(header)+4), 16)
				if len(s)%4 > 0 {
					s = strings.Repeat("0", 4-len(s)%4) + s
				}
				pktline.WriteString(s)
				pktline.WriteString(header)
				pktline.WriteString("0000")
				res.Writer.Write(pktline.Bytes())
			}

			cmd := exec.Command("git", strings.TrimPrefix(service, "git-"), "--stateless-rpc", "--advertise-refs", repodir)
			var stderr bytes.Buffer
			cmd.Stdin = c.Request().Body
			cmd.Stdout = res.Writer
			cmd.Stderr = &stderr

			log.Debug().Strs("command", cmd.Args).Msg("Executing git command")
			if err := cmd.Run(); err != nil {
				// log error from command execution
				log.Error().Err(err).Str("stderr", stderr.String()).Msg("git command failed")
			}

			return nil
		})

		handleSmartService := func(service string) echo.HandlerFunc {
			return func(c echo.Context) error {
				username := c.Param("username")
				reponame := c.Param("reponame")
				repodir := lo.Must(filepath.Abs(filepath.Join(rootFlags.root, username, reponame)))

				res := c.Response()
				header := res.Header()
				header.Set("Content-Type", fmt.Sprintf("application/x-%s-result", service))
				header.Set("Cache-Control", "no-cache")

				res.Writer.WriteHeader(http.StatusOK)

				cmd := exec.Command("git", strings.TrimPrefix(service, "git-"), "--stateless-rpc", repodir)
				var stderr bytes.Buffer
				cmd.Stdin = c.Request().Body
				cmd.Stdout = res.Writer
				cmd.Stderr = &stderr

				log.Debug().Strs("command", cmd.Args).Msg("Executing git command")
				if err := cmd.Run(); err != nil {
					log.Error().Err(err).Str("stderr", stderr.String()).Msg("git command failed")
				}

				return nil
			}
		}

		g.POST("/git-upload-pack", handleSmartService("git-upload-pack"))
		g.POST("/git-receive-pack", handleSmartService("git-receive-pack"))

		if err := e.Start(fmt.Sprintf(":%d", rootFlags.port)); err != nil {
			log.Fatal().Err(err).Msg("shutting down the server")
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
	rootCmd.PersistentFlags().BoolVarP(&rootFlags.verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().IntVarP(&rootFlags.port, "port", "p", 8080, "Port to listen on")
	rootCmd.Flags().StringVarP(&rootFlags.root, "root", "r", "./repos", "Root directory to store repositories")
}
