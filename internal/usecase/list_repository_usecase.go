package usecase

import (
	"context"

	"github.com/samber/do"
	"github.com/yz4230/githost-poc/internal/entity"
	"github.com/yz4230/githost-poc/internal/repository"
)

type ListRepositoryUsecase interface {
	Execute(ctx context.Context) ([]*entity.Repository, error)
}

type listRepositoryUsecaseImpl struct {
	repositoryRepository repository.RepositoryRepository
}

// Execute implements ListRepositoryUsecase.
func (l *listRepositoryUsecaseImpl) Execute(ctx context.Context) ([]*entity.Repository, error) {
	return l.repositoryRepository.List(ctx)
}

func NewListRepositoryUsecase(injector *do.Injector) (ListRepositoryUsecase, error) {
	return &listRepositoryUsecaseImpl{
		repositoryRepository: do.MustInvoke[repository.RepositoryRepository](injector),
	}, nil
}
