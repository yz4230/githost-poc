package usecase

import (
	"context"

	"github.com/samber/do"
	"github.com/yz4230/githost-poc/internal/entity"
	"github.com/yz4230/githost-poc/internal/repository"
)

type GetRepositoryByIdUsecase interface {
	Execute(ctx context.Context, id entity.ID) (*entity.Repository, error)
}

type getRepositoryByIdUsecaseImpl struct {
	repositoryRepository repository.RepositoryRepository
}

// Execute implements GetRepositoryByIdUsecase.
func (g *getRepositoryByIdUsecaseImpl) Execute(ctx context.Context, id entity.ID) (*entity.Repository, error) {
	return g.repositoryRepository.GetByID(ctx, id)
}

func NewGetRepositoryByIdUsecase(injector *do.Injector) (GetRepositoryByIdUsecase, error) {
	return &getRepositoryByIdUsecaseImpl{
		repositoryRepository: do.MustInvoke[repository.RepositoryRepository](injector),
	}, nil
}
