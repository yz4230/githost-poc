package repository

import (
	"github.com/yz4230/githost-poc/internal/entity"
	"gorm.io/gorm"
)

type Repository struct {
	gorm.Model
	Name         string
	Description  string
	DeployBranch string
	LatestSHA    string
}

func (r *Repository) ToEntity() *entity.Repository {
	return &entity.Repository{
		ID:           entity.NewID(r.ID),
		Name:         r.Name,
		Description:  r.Description,
		DeployBranch: r.DeployBranch,
		LatestSHA:    r.LatestSHA,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

func (r *Repository) FromEntity(e *entity.Repository) {
	r.ID = e.ID.Uint()
	r.Name = e.Name
	r.Description = e.Description
	r.DeployBranch = e.DeployBranch
	r.LatestSHA = e.LatestSHA
}

type Deployment struct {
	gorm.Model
	RepoID    uint
	Repo      Repository
	Branch    string
	CommitSHA string
	Status    string
	IsActive  bool
}

func (d *Deployment) ToEntity() *entity.Deployment {
	return &entity.Deployment{
		ID:        entity.NewID(d.ID),
		RepoID:    entity.NewID(d.RepoID),
		CommitSHA: d.CommitSHA,
		Status:    entity.DeploymentStatus(d.Status),
		IsActive:  d.IsActive,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func (d *Deployment) FromEntity(e *entity.Deployment) {
	d.ID = e.ID.Uint()
	d.RepoID = e.RepoID.Uint()
	d.CommitSHA = e.CommitSHA
	d.Status = string(e.Status)
	d.IsActive = e.IsActive
}
