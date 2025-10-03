package routes

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/samber/do"
	"github.com/yz4230/githost-poc/internal/git"
	"github.com/yz4230/githost-poc/internal/storage"
)

func RegisterGitSmartHTTP(injector *do.Injector, e *echo.Echo) {
	g := e.Group("/repos/:reponame")

	// Validate reponame
	g.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		reReponame := regexp.MustCompile(`^[a-zA-Z0-9_-]+\.git$`)
		return func(c echo.Context) error {
			ua := c.Request().UserAgent()
			if !strings.HasPrefix(ua, "git/") {
				return c.NoContent(http.StatusBadRequest)
			}
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
