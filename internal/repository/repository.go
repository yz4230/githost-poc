package repository

import (
	"context"

	"github.com/yz4230/githost-poc/internal/entity"
	"gorm.io/gorm"
)

type RepositoryRepository interface {
	Create(ctx context.Context, repo *entity.Repository) (*entity.Repository, error)
	GetByID(ctx context.Context, id entity.ID) (*entity.Repository, error)
	GetByName(ctx context.Context, name string) (*entity.Repository, error)
	List(ctx context.Context) ([]*entity.Repository, error)
	Update(ctx context.Context, repo *entity.Repository) (*entity.Repository, error)
	Delete(ctx context.Context, id entity.ID) error
}

type RepositoryRepositoryImpl struct {
	db *gorm.DB
}

// Create implements RepoRepository.
func (r *RepositoryRepositoryImpl) Create(ctx context.Context, repo *entity.Repository) (*entity.Repository, error) {
	var model Repository
	model.FromEntity(repo)
	err := gorm.G[Repository](r.db).Create(ctx, &model)
	if err != nil {
		return nil, err
	}
	return model.ToEntity(), nil
}

// GetByID implements RepoRepository.
func (r *RepositoryRepositoryImpl) GetByID(ctx context.Context, id entity.ID) (*entity.Repository, error) {
	found, err := gorm.G[Repository](r.db).Where("id = ?", id.Uint()).First(ctx)
	if err != nil {
		return nil, err
	}
	return found.ToEntity(), nil
}

// GetByName implements RepoRepository.
func (r *RepositoryRepositoryImpl) GetByName(ctx context.Context, name string) (*entity.Repository, error) {
	found, err := gorm.G[Repository](r.db).Where("name = ?", name).First(ctx)
	if err != nil {
		return nil, err
	}
	return found.ToEntity(), nil
}

// List implements RepoRepository.
func (r *RepositoryRepositoryImpl) List(ctx context.Context) ([]*entity.Repository, error) {
	founds, err := gorm.G[Repository](r.db).Find(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Repository, len(founds))
	for i, f := range founds {
		result[i] = f.ToEntity()
	}
	return result, nil
}

// Update implements RepoRepository.
func (r *RepositoryRepositoryImpl) Update(ctx context.Context, repo *entity.Repository) (*entity.Repository, error) {
	var model Repository
	model.FromEntity(repo)
	_, err := gorm.G[Repository](r.db).Where("id = ?", repo.ID.Uint()).Updates(ctx, model)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, repo.ID)
}

// Delete implements RepoRepository.
func (r *RepositoryRepositoryImpl) Delete(ctx context.Context, id entity.ID) error {
	_, err := gorm.G[Repository](r.db).Where("id = ?", id.Uint()).Delete(ctx)
	return err
}

func NewRepositoryRepository(db *gorm.DB) RepositoryRepository {
	return &RepositoryRepositoryImpl{db: db}
}
