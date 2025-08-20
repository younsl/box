package storage

import (
	"encoding/json"
	"os"
	"time"

	"github.com/younsl/box/resume/references/job-tracker/pkg/crypto"
	"github.com/younsl/box/resume/references/job-tracker/pkg/models"
)

type Storage struct {
	dataFile string
	crypto   *crypto.GPGCrypto
	store    *models.DataStore
}

func NewStorage(dataFile, gpgRecipient string) *Storage {
	return &Storage{
		dataFile: dataFile,
		crypto:   crypto.NewGPGCrypto(gpgRecipient),
		store:    &models.DataStore{Applications: []models.JobApplication{}},
	}
}

func (s *Storage) Load() error {
	if _, err := os.Stat(s.dataFile); os.IsNotExist(err) {
		return nil
	}

	data, err := s.crypto.Decrypt(s.dataFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, s.store)
}

func (s *Storage) Save() error {
	s.store.LastUpdated = time.Now()
	
	data, err := json.MarshalIndent(s.store, "", "  ")
	if err != nil {
		return err
	}

	return s.crypto.EncryptData(data, s.dataFile)
}

func (s *Storage) GetApplications() []models.JobApplication {
	return s.store.Applications
}

func (s *Storage) AddApplication(app models.JobApplication) {
	app.LastModified = time.Now()
	s.store.Applications = append(s.store.Applications, app)
}

func (s *Storage) UpdateApplication(app models.JobApplication) {
	app.LastModified = time.Now()
	for i, a := range s.store.Applications {
		if a.ID == app.ID {
			s.store.Applications[i] = app
			break
		}
	}
}

func (s *Storage) DeleteApplication(id string) {
	for i, app := range s.store.Applications {
		if app.ID == id {
			s.store.Applications = append(s.store.Applications[:i], s.store.Applications[i+1:]...)
			break
		}
	}
}