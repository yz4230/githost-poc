package usecase

import (
	"context"

	"github.com/samber/do"
	"github.com/yz4230/githost-poc/internal/entity"
	"github.com/yz4230/githost-poc/internal/repository"
	"github.com/yz4230/githost-poc/internal/storage"
)

type CreateRepositoryUsecase interface {
	Execute(ctx context.Context, repo *entity.Repository) (*entity.Repository, error)
}

type createRepositoryUsecaseImpl struct {
	gitStorage           storage.GitStorage
	repositoryRepository repository.RepositoryRepository
}

// Execute implements CreateRepositoryUsecase.
func (c *createRepositoryUsecaseImpl) Execute(ctx context.Context, repo *entity.Repository) (*entity.Repository, error) {
	repo.FillDefaults()
	if exists := c.gitStorage.IsRepoExist(repo.Name); exists {
		return nil, entity.ErrConflict
	}
	if err := c.gitStorage.InitBareRepo(ctx, repo.Name); err != nil {
		return nil, entity.ErrInternal
	}
	return c.repositoryRepository.Create(ctx, repo)
}

func NewCreateRepositoryUsecase(injector *do.Injector) (CreateRepositoryUsecase, error) {
	return &createRepositoryUsecaseImpl{
		gitStorage:           do.MustInvoke[storage.GitStorage](injector),
		repositoryRepository: do.MustInvoke[repository.RepositoryRepository](injector),
	}, nil
}
