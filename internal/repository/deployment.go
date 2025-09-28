package repository

import (
	"context"

	"github.com/yz4230/githost-poc/internal/entity"
	"gorm.io/gorm"
)

type DeploymentRepository interface {
	Create(ctx context.Context, dep *entity.Deployment) (*entity.Deployment, error)
	GetByID(ctx context.Context, id entity.ID) (*entity.Deployment, error)
	List(ctx context.Context) ([]*entity.Deployment, error)
	ListByRepo(ctx context.Context, repoID entity.ID) ([]*entity.Deployment, error)
	Update(ctx context.Context, dep *entity.Deployment) (*entity.Deployment, error)
	Delete(ctx context.Context, id entity.ID) error
}

type deploymentRepositoryImpl struct {
	db *gorm.DB
}

func NewDeploymentRepository(db *gorm.DB) DeploymentRepository {
	return &deploymentRepositoryImpl{db: db}
}

// Create a new deployment record.
func (r *deploymentRepositoryImpl) Create(ctx context.Context, dep *entity.Deployment) (*entity.Deployment, error) {
	var model Deployment
	model.FromEntity(dep)
	if err := gorm.G[Deployment](r.db).Create(ctx, &model); err != nil {
		return nil, err
	}
	return model.ToEntity(), nil
}

// GetByID finds deployment by id.
func (r *deploymentRepositoryImpl) GetByID(ctx context.Context, id entity.ID) (*entity.Deployment, error) {
	found, err := gorm.G[Deployment](r.db).Where("id = ?", id.Uint()).First(ctx)
	if err != nil {
		return nil, err
	}
	return found.ToEntity(), nil
}

// List returns all deployments.
func (r *deploymentRepositoryImpl) List(ctx context.Context) ([]*entity.Deployment, error) {
	founds, err := gorm.G[Deployment](r.db).Find(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*entity.Deployment, len(founds))
	for i, f := range founds {
		res[i] = f.ToEntity()
	}
	return res, nil
}

// ListByRepo lists deployments belonging to a repository.
func (r *deploymentRepositoryImpl) ListByRepo(ctx context.Context, repoID entity.ID) ([]*entity.Deployment, error) {
	founds, err := gorm.G[Deployment](r.db).Where("repo_id = ?", repoID.Uint()).Find(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*entity.Deployment, len(founds))
	for i, f := range founds {
		res[i] = f.ToEntity()
	}
	return res, nil
}

// Update deployment record (status, active flag, commit sha etc.).
func (r *deploymentRepositoryImpl) Update(ctx context.Context, dep *entity.Deployment) (*entity.Deployment, error) {
	var model Deployment
	model.FromEntity(dep)
	_, err := gorm.G[Deployment](r.db).Where("id = ?", dep.ID.Uint()).Updates(ctx, model)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, dep.ID)
}

// Delete deployment by id.
func (r *deploymentRepositoryImpl) Delete(ctx context.Context, id entity.ID) error {
	_, err := gorm.G[Deployment](r.db).Where("id = ?", id.Uint()).Delete(ctx)
	return err
}
