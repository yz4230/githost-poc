package usecase

import (
	"context"

	"github.com/samber/do"
	"github.com/yz4230/githost-poc/internal/entity"
	"github.com/yz4230/githost-poc/internal/repository"
)

type CheckRepositoryNameUsecase interface {
	Execute(ctx context.Context, name string) (bool, error)
}

type checkRepositoryNameUsecaseImpl struct {
	repositoryRepository repository.RepositoryRepository
}

func (c *checkRepositoryNameUsecaseImpl) Execute(ctx context.Context, name string) (bool, error) {
	_, err := c.repositoryRepository.GetByName(ctx, name)
	if err == entity.ErrNotFound {
		return true, nil
	}
	return false, err
}

func NewCheckRepositoryNameUsecase(i *do.Injector) (CheckRepositoryNameUsecase, error) {
	return &checkRepositoryNameUsecaseImpl{
		repositoryRepository: do.MustInvoke[repository.RepositoryRepository](i),
	}, nil
}
