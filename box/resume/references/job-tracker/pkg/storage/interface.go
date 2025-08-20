package storage

import "github.com/younsl/box/resume/references/job-tracker/pkg/models"

// StorageInterface defines the common interface for all storage implementations
type StorageInterface interface {
	GetApplications() []models.JobApplication
	AddApplication(app models.JobApplication) error
	UpdateApplication(app models.JobApplication) error
	DeleteApplication(id string) error
	Save() error
	Load() error
}