package entity

import "time"

type DeploymentStatus string

const (
	DeploymentStatusPending DeploymentStatus = "pending"
	DeploymentStatusRunning DeploymentStatus = "running"
	DeploymentStatusSuccess DeploymentStatus = "success"
	DeploymentStatusFailed  DeploymentStatus = "failed"
)

type Deployment struct {
	ID        ID               `json:"id"`
	RepoID    ID               `json:"repo_id"`
	CommitSHA string           `json:"commit_sha"`
	Status    DeploymentStatus `json:"status"`
	IsActive  bool             `json:"is_active"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}
