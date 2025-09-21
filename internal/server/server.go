package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/yz4230/githost-poc/internal/gitopt"
)

type Config struct {
	Root   string
	Port   int
	Logger zerolog.Logger
}

type Server struct {
	e   *echo.Echo
	cfg *Config
}

func New(cfg *Config) *Server {
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
			cfg.Logger.Info().
				Str("remote_ip", v.RemoteIP).
				Str("host", v.Host).
				Str("method", v.Method).
				Str("uri", v.URI).
				Str("user_agent", v.UserAgent).
				Int("status", v.Status).
				Int64("latency_ms", v.Latency.Milliseconds()).
				Msg("handled request")
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := cfg.Logger.WithContext(req.Context())
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	})

	s := &Server{e: e, cfg: cfg}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	g := s.e.Group("/:username/:reponame")

	g.GET("/info/refs", func(c echo.Context) error {
		req, res := c.Request(), c.Response()
		username, reponame := c.Param("username"), c.Param("reponame")
		repodir, err := gitopt.EnsureBareRepo(req.Context(), s.cfg.Root, username, reponame)
		if err != nil {
			s.cfg.Logger.Error().Err(err).Msg("ensure repo failed")
			return c.NoContent(http.StatusInternalServerError)
		}
		service := gitopt.Service(c.QueryParam("service"))
		res.Header().Set("Content-Type", "application/x-"+string(service)+"-advertisement")
		res.Header().Set("Cache-Control", "no-cache")
		if err := gitopt.AdvertiseRefs(req.Context(), service, repodir, res.Writer); err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return nil
	})

	smartHandler := func(service gitopt.Service) echo.HandlerFunc {
		return func(c echo.Context) error {
			req, res := c.Request(), c.Response()
			username, reponame := c.Param("username"), c.Param("reponame")
			repodir, err := gitopt.EnsureBareRepo(req.Context(), s.cfg.Root, username, reponame)
			if err != nil {
				s.cfg.Logger.Error().Err(err).Msg("ensure repo failed")
				return c.NoContent(http.StatusInternalServerError)
			}
			res.Header().Set("Content-Type", "application/x-"+string(service)+"-result")
			res.Header().Set("Cache-Control", "no-cache")
			if err := gitopt.ExecStatelessRPC(req.Context(), service, repodir, req.Body, res.Writer); err != nil {
				return c.NoContent(http.StatusInternalServerError)
			}
			return nil
		}
	}

	g.POST("/git-upload-pack", smartHandler(gitopt.ServiceUploadPack))
	g.POST("/git-receive-pack", smartHandler(gitopt.ServiceReceivePack))
}

func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	s.cfg.Logger.Info().Str("addr", addr).Msg("starting server")

	errCh := make(chan error, 1)
	go func() {
		if err := s.e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.e.Shutdown(shutdownCtx); err != nil {
			s.cfg.Logger.Error().Err(err).Msg("graceful shutdown failed")
			return err
		}
		s.cfg.Logger.Info().Msg("server shutdown complete")
		return ctx.Err() // propagate context cancellation (root may choose to ignore)
	case err := <-errCh:
		if err != nil {
			s.cfg.Logger.Error().Err(err).Msg("server exited with error")
			return err
		}
		return nil
	}
}
