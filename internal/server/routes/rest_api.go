package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/samber/do"
	"github.com/yz4230/githost-poc/internal/entity"
	"github.com/yz4230/githost-poc/internal/usecase"
)

func RegisterRestAPI(injector *do.Injector, e *echo.Echo) {
	g := e.Group("/api")

	g.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	g.POST("/repositories", func(c echo.Context) error {
		type request struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		var req request
		if err := c.Bind(&req); err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		usecase := do.MustInvoke[usecase.CreateRepositoryUsecase](injector)
		repo, err := usecase.Execute(c.Request().Context(), &entity.Repository{
			Name:        req.Name,
			Description: req.Description,
		})
		if err != nil {
			if err == entity.ErrInvalid {
				return c.NoContent(http.StatusBadRequest)
			}
			if err == entity.ErrConflict {
				return c.NoContent(http.StatusConflict)
			}
			return c.NoContent(http.StatusInternalServerError)
		}

		return c.JSON(http.StatusCreated, repo)
	})
	g.GET("/repositories", func(c echo.Context) error {
		usecase := do.MustInvoke[usecase.ListRepositoryUsecase](injector)
		repos, err := usecase.Execute(c.Request().Context())
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		type response struct {
			Repositories []*entity.Repository `json:"repositories"`
		}

		result := &response{Repositories: make([]*entity.Repository, len(repos))}
		copy(result.Repositories, repos)

		return c.JSON(http.StatusOK, result)
	})
}
