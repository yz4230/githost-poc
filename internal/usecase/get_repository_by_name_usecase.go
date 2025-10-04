package usecase

import (
	"context"

	"github.com/samber/do"
	"github.com/yz4230/githost-poc/internal/entity"
	"github.com/yz4230/githost-poc/internal/repository"
)

type GetRepositoryByNameUsecase interface {
	Execute(ctx context.Context, name string) (*entity.Repository, error)
}

type getRepositoryByNameUsecaseImpl struct {
	repositoryRepository repository.RepositoryRepository
}

// Execute implements GetRepositoryByNameUsecase.
func (g *getRepositoryByNameUsecaseImpl) Execute(ctx context.Context, name string) (*entity.Repository, error) {
	return g.repositoryRepository.GetByName(ctx, name)
}

func NewGetRepositoryByNameUsecase(injector *do.Injector) (GetRepositoryByNameUsecase, error) {
	return &getRepositoryByNameUsecaseImpl{
		repositoryRepository: do.MustInvoke[repository.RepositoryRepository](injector),
	}, nil
}
