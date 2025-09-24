package server

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/samber/do"
	"github.com/yz4230/githost-poc/internal/git"
	"github.com/yz4230/githost-poc/internal/storage"
)

type Config struct {
	Root   string
	Port   int
	Logger zerolog.Logger
}

type Server struct {
	e      *echo.Echo
	config *Config
}

func New(config *Config) *Server {
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
			config.Logger.Info().
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
			ctx := config.Logger.WithContext(req.Context())
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	})

	s := &Server{e: e, config: config}
	s.init()
	return s
}

func (s *Server) init() {
	injector := do.New()
	s.injectDependencies(injector)
	s.registerRoutes(injector)
}

func (s *Server) injectDependencies(injector *do.Injector) {
	do.Provide(injector, func(i *do.Injector) (storage.GitStorage, error) {
		root := filepath.Join(s.config.Root, "repositories")
		return storage.NewGitStorage(root, s.config.Logger), nil
	})
}

func (s *Server) registerRoutes(injector *do.Injector) {
	s.registerRestAPI(injector)
	s.registerGitSmartHTTP(injector)
}

func (s *Server) registerRestAPI(injector *do.Injector) {
	g := s.e.Group("/api")

	g.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
}

func (s *Server) registerGitSmartHTTP(injector *do.Injector) {
	g := s.e.Group("/:reponame")

	// Validate reponame
	g.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		reReponame := regexp.MustCompile(`^[a-zA-Z0-9_-]+\.git$`)
		return func(c echo.Context) error {
			reponame := c.Param("reponame")
			if !reReponame.MatchString(reponame) {
				return c.NoContent(http.StatusNotFound)
			}
			return next(c)
		}
	})

	g.GET("/info/refs", func(c echo.Context) error {
		storage := do.MustInvoke[storage.GitStorage](injector)

		req, res := c.Request(), c.Response()
		reponame := strings.TrimSuffix(c.Param("reponame"), ".git")
		repodir := storage.GetRepoDir(reponame)

		if err := storage.EnsureBareRepo(req.Context(), reponame); err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		service := c.QueryParam("service")
		res.Header().Set("Content-Type", "application/x-"+service+"-advertisement")
		res.Header().Set("Cache-Control", "no-cache")
		if err := git.AdvertiseRefs(req.Context(), service, repodir, res.Writer); err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return nil
	})

	smartHandler := func(service string) echo.HandlerFunc {
		return func(c echo.Context) error {
			storage := do.MustInvoke[storage.GitStorage](injector)

			req, res := c.Request(), c.Response()
			reponame := strings.TrimSuffix(c.Param("reponame"), ".git")
			repodir := storage.GetRepoDir(reponame)

			res.Header().Set("Content-Type", "application/x-"+service+"-result")
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
	addr := fmt.Sprintf(":%d", s.config.Port)
	s.config.Logger.Info().Str("addr", addr).Msg("starting server")
	return s.e.Start(addr)
}

func (s *Server) Stop(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}
