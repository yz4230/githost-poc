package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/yz4230/githost-poc/internal/git"
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
	g := s.e.Group("/:reponame")

	g.GET("/info/refs", func(c echo.Context) error {
		req, res := c.Request(), c.Response()
		reponame := c.Param("reponame")
		repodir, err := git.EnsureBareRepo(req.Context(), s.cfg.Root, reponame)
		if err != nil {
			s.cfg.Logger.Error().Err(err).Msg("ensure repo failed")
			return c.NoContent(http.StatusInternalServerError)
		}
		service := git.Service(c.QueryParam("service"))
		res.Header().Set("Content-Type", "application/x-"+string(service)+"-advertisement")
		res.Header().Set("Cache-Control", "no-cache")
		if err := git.AdvertiseRefs(req.Context(), service, repodir, res.Writer); err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return nil
	})

	smartHandler := func(service git.Service) echo.HandlerFunc {
		return func(c echo.Context) error {
			req, res := c.Request(), c.Response()
			reponame := c.Param("reponame")
			repodir, err := git.EnsureBareRepo(req.Context(), s.cfg.Root, reponame)
			if err != nil {
				s.cfg.Logger.Error().Err(err).Msg("ensure repo failed")
				return c.NoContent(http.StatusInternalServerError)
			}
			res.Header().Set("Content-Type", "application/x-"+string(service)+"-result")
			res.Header().Set("Cache-Control", "no-cache")
			if err := git.ExecStatelessRPC(req.Context(), service, repodir, req.Body, res.Writer); err != nil {
				return c.NoContent(http.StatusInternalServerError)
			}
			return nil
		}
	}

	g.POST("/git-upload-pack", smartHandler(git.ServiceUploadPack))
	g.POST("/git-receive-pack", smartHandler(git.ServiceReceivePack))
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	s.cfg.Logger.Info().Str("addr", addr).Msg("starting server")
	return s.e.Start(addr)
}

func (s *Server) Stop(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}
