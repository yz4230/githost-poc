package usecase

import (
	"context"

	"github.com/samber/do"
	"github.com/yz4230/githost-poc/internal/entity"
	"github.com/yz4230/githost-poc/internal/repository"
)

type GetRepositoryUsecase interface {
	Execute(ctx context.Context, name string) (*entity.Repository, error)
}

type getRepositoryUsecaseImpl struct {
	repositoryRepository repository.RepositoryRepository
}

// Execute implements GetRepositoryUsecase.
func (g *getRepositoryUsecaseImpl) Execute(ctx context.Context, name string) (*entity.Repository, error) {
	return g.repositoryRepository.GetByName(ctx, name)
}

func NewGetRepositoryUsecase(injector *do.Injector) (GetRepositoryUsecase, error) {
	return &getRepositoryUsecaseImpl{
		repositoryRepository: do.MustInvoke[repository.RepositoryRepository](injector),
	}, nil
}
