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
	ID        ID
	RepoID    ID
	CommitSHA string
	Status    DeploymentStatus
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
