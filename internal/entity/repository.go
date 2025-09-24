package entity

import "time"

type Repository struct {
	ID           ID
	Name         string
	Description  string
	DeployBranch string
	LatestSHA    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
