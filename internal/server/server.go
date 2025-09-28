package server

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/samber/do"
	"github.com/yz4230/githost-poc/internal/repository"
	"github.com/yz4230/githost-poc/internal/server/routes"
	"github.com/yz4230/githost-poc/internal/storage"
	"github.com/yz4230/githost-poc/internal/usecase"
	"gorm.io/gorm"
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
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			config.Logger.Error().Err(err).Bytes("stack", stack).Send()
			return err
		},
	}))
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
	do.Provide(injector, func(i *do.Injector) (*gorm.DB, error) {
		return repository.NewSQLiteDB(s.config.Root)
	})
	do.Provide(injector, func(i *do.Injector) (storage.GitStorage, error) {
		root := filepath.Join(s.config.Root, "repositories")
		return storage.NewGitStorage(root, s.config.Logger), nil
	})
	do.Provide(injector, func(i *do.Injector) (repository.RepositoryRepository, error) {
		db := do.MustInvoke[*gorm.DB](i)
		return repository.NewRepositoryRepository(db), nil
	})
	do.Provide(injector, usecase.NewCreateRepositoryUsecase)
	do.Provide(injector, usecase.NewListRepositoryUsecase)
}

func (s *Server) registerRoutes(injector *do.Injector) {
	routes.RegisterRestAPI(injector, s.e)
	routes.RegisterGitSmartHTTP(injector, s.e)
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	s.config.Logger.Info().Str("addr", addr).Msg("starting server")
	return s.e.Start(addr)
}

func (s *Server) Stop(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}
