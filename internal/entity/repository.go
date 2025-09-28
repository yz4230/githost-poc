package entity

import (
	"time"
)

type Repository struct {
	ID           ID        `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	DeployBranch string    `json:"deploy_branch"`
	LatestSHA    string    `json:"latest_sha"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (r *Repository) FillDefaults() {
	if r.DeployBranch == "" {
		r.DeployBranch = "main"
	}
	if r.LatestSHA == "" {
		r.LatestSHA = "0000000000000000000000000000000000000000" // 40 zeros
	}
}
