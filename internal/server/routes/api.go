package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/samber/do"
	"github.com/yz4230/githost-poc/internal/entity"
	"github.com/yz4230/githost-poc/internal/usecase"
	"github.com/yz4230/githost-poc/internal/utils"
)

func RegisterAPI(injector *do.Injector, e *echo.Echo) {
	api := e.Group("/api")

	api.POST("/check-name", func(c echo.Context) error {
		type request struct {
			Name string `json:"name"`
		}
		var req request
		if err := c.Bind(&req); err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		name := utils.SanitizeName(req.Name)
		usecase := do.MustInvoke[usecase.CheckRepositoryNameUsecase](injector)
		available, err := usecase.Execute(c.Request().Context(), name)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		type response struct {
			Name      string `json:"name"`
			Available bool   `json:"available"`
		}
		return c.JSON(http.StatusOK, &response{
			Name:      name,
			Available: available,
		})

	})
	api.POST("/repositories", func(c echo.Context) error {
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
	api.GET("/repositories", func(c echo.Context) error {
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
