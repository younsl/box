package storage

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/younsl/box/resume/references/job-tracker/pkg/crypto"
	"github.com/younsl/box/resume/references/job-tracker/pkg/logging"
	"github.com/younsl/box/resume/references/job-tracker/pkg/models"
)

type DatabaseStorage struct {
	db           *sql.DB
	encryptedFile string
	tempDBFile   string
	crypto       *crypto.GPGCrypto
}

func NewDatabaseStorage(encryptedFile, gpgRecipient string) *DatabaseStorage {
	tempDir := os.TempDir()
	tempDBFile := filepath.Join(tempDir, "job_tracker_temp.db")
	
	return &DatabaseStorage{
		encryptedFile: encryptedFile,
		tempDBFile:   tempDBFile,
		crypto:       crypto.NewGPGCrypto(gpgRecipient),
	}
}

// getFileSize returns formatted file size for logging
func (s *DatabaseStorage) getFileSize() string {
	if stat, err := os.Stat(s.encryptedFile); err == nil {
		size := stat.Size()
		return formatFileSize(size)
	}
	return "unknown"
}

// formatFileSize formats bytes to human readable format
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (s *DatabaseStorage) Initialize() error {
	// Load encrypted database if exists
	if err := s.Load(); err != nil {
		logging.Logger.WithError(err).Warn("Could not load existing encrypted database")
	}

	// Open temporary database
	db, err := sql.Open("sqlite3", s.tempDBFile)
	if err != nil {
		return err
	}
	s.db = db

	// Create tables if not exists
	createApplicationsTableSQL := `
	CREATE TABLE IF NOT EXISTS job_applications (
		id TEXT PRIMARY KEY,
		company TEXT NOT NULL,
		position TEXT NOT NULL,
		status TEXT NOT NULL,
		final_result TEXT,
		applied_date TEXT NOT NULL,
		platform TEXT,
		url TEXT,
		notes TEXT,
		last_modified DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	createFilesTableSQL := `
	CREATE TABLE IF NOT EXISTS application_files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		application_id TEXT NOT NULL,
		filename TEXT NOT NULL,
		description TEXT,
		content_type TEXT,
		file_data BLOB NOT NULL,
		file_size INTEGER NOT NULL,
		uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (application_id) REFERENCES job_applications(id) ON DELETE CASCADE
	);
	`

	if _, err = s.db.Exec(createApplicationsTableSQL); err != nil {
		return err
	}

	if _, err = s.db.Exec(createFilesTableSQL); err != nil {
		return err
	}

	// Add description column if it doesn't exist (for existing databases)
	_, err = s.db.Exec("ALTER TABLE application_files ADD COLUMN description TEXT")
	if err != nil {
		// Column already exists or other error - continue
		logging.Logger.WithError(err).Debug("Note: Could not add description column (may already exist)")
	}

	logging.Logger.WithFields(map[string]interface{}{"file": s.encryptedFile, "size": s.getFileSize()}).Info("Database initialized")
	return nil
}

func (s *DatabaseStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *DatabaseStorage) GetApplications() []models.JobApplication {
	rows, err := s.db.Query(`
		SELECT id, company, position, status, final_result, applied_date, 
			   platform, url, notes, last_modified 
		FROM job_applications 
		ORDER BY last_modified DESC
	`)
	if err != nil {
		logging.Logger.WithError(err).Error("Error querying applications")
		return []models.JobApplication{}
	}
	defer rows.Close()

	var applications []models.JobApplication
	for rows.Next() {
		var app models.JobApplication
		var finalResult sql.NullString
		var platform sql.NullString
		var url sql.NullString
		var notes sql.NullString

		err := rows.Scan(
			&app.ID, &app.Company, &app.Position, &app.Status,
			&finalResult, &app.AppliedDate, &platform, &url, &notes,
			&app.LastModified,
		)
		if err != nil {
			logging.Logger.WithError(err).Error("Error scanning row")
			continue
		}

		// Handle nullable fields
		app.FinalResult = finalResult.String
		app.Platform = platform.String
		app.URL = url.String
		app.Notes = notes.String

		// Get file info for this application
		app.Files = s.getFileInfo(app.ID)

		applications = append(applications, app)
	}

	return applications
}

func (s *DatabaseStorage) getFileInfo(appID string) []models.FileInfo {
	rows, err := s.db.Query("SELECT filename, COALESCE(description, ''), file_size FROM application_files WHERE application_id = ? ORDER BY uploaded_at", appID)
	if err != nil {
		logging.Logger.WithError(err).WithField("app_id", appID).Error("Error querying files")
		return []models.FileInfo{}
	}
	defer rows.Close()

	var files []models.FileInfo
	for rows.Next() {
		var fileInfo models.FileInfo
		if err := rows.Scan(&fileInfo.Filename, &fileInfo.Description, &fileInfo.Size); err != nil {
			logging.Logger.WithError(err).Error("Error scanning file info")
			continue
		}
		files = append(files, fileInfo)
	}
	return files
}

func (s *DatabaseStorage) AddApplication(app models.JobApplication) error {
	_, err := s.db.Exec(`
		INSERT INTO job_applications 
		(id, company, position, status, final_result, applied_date, platform, url, notes, last_modified)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, app.ID, app.Company, app.Position, app.Status, app.FinalResult, app.AppliedDate,
		app.Platform, app.URL, app.Notes, time.Now())

	if err != nil {
		logging.Logger.WithError(err).WithField("app_id", app.ID).Error("Error adding application")
	}
	return err
}

func (s *DatabaseStorage) UpdateApplication(app models.JobApplication) error {
	_, err := s.db.Exec(`
		UPDATE job_applications 
		SET company = ?, position = ?, status = ?, final_result = ?, applied_date = ?,
			platform = ?, url = ?, notes = ?, last_modified = ?
		WHERE id = ?
	`, app.Company, app.Position, app.Status, app.FinalResult, app.AppliedDate,
		app.Platform, app.URL, app.Notes, time.Now(), app.ID)

	if err != nil {
		logging.Logger.WithError(err).WithField("app_id", app.ID).Error("Error updating application")
	}
	return err
}

func (s *DatabaseStorage) DeleteApplication(id string) error {
	_, err := s.db.Exec("DELETE FROM job_applications WHERE id = ?", id)
	if err != nil {
		logging.Logger.WithError(err).WithField("app_id", id).Error("Error deleting application")
	}
	return err
}

func (s *DatabaseStorage) Save() error {
	// Close current connection
	if s.db != nil {
		s.db.Close()
	}

	// Read temporary database file
	dbData, err := ioutil.ReadFile(s.tempDBFile)
	if err != nil {
		return err
	}

	// Encrypt and save to encrypted file
	if err := s.crypto.EncryptData(dbData, s.encryptedFile); err != nil {
		return err
	}

	// Reopen database
	db, err := sql.Open("sqlite3", s.tempDBFile)
	if err != nil {
		return err
	}
	s.db = db

	logging.Logger.WithFields(map[string]interface{}{"file": s.encryptedFile, "size": s.getFileSize()}).Info("Database encrypted and saved")
	return nil
}

func (s *DatabaseStorage) Load() error {
	// Check if encrypted file exists
	if _, err := os.Stat(s.encryptedFile); os.IsNotExist(err) {
		return nil // No existing file to load
	}

	// Decrypt the database file
	decryptedData, err := s.crypto.Decrypt(s.encryptedFile)
	if err != nil {
		return err
	}

	// Write decrypted data to temporary file
	if err := ioutil.WriteFile(s.tempDBFile, decryptedData, 0600); err != nil {
		return err
	}

	logging.Logger.WithFields(map[string]interface{}{"file": s.encryptedFile, "size": s.getFileSize()}).Info("Database decrypted and loaded")
	return nil
}

// Migration function to convert from JSON file to SQLite
func (s *DatabaseStorage) MigrateFromJSON(jsonFile string) error {
	// Read existing JSON data if it exists
	oldStorage := NewStorage(jsonFile, "")
	if err := oldStorage.Load(); err != nil {
		logging.Logger.WithError(err).Debug("No existing JSON file to migrate or error loading")
		return nil
	}

	applications := oldStorage.GetApplications()
	logging.Logger.WithField("count", len(applications)).Info("Migrating applications from JSON to SQLite")

	for _, app := range applications {
		if err := s.AddApplication(app); err != nil {
			logging.Logger.WithError(err).WithField("app_id", app.ID).Error("Error migrating application")
		}
	}

	logging.Logger.Info("Migration completed successfully")
	return nil
}

// File management methods
func (s *DatabaseStorage) AddFile(appID, filename, description, contentType string, fileData []byte) error {
	_, err := s.db.Exec(`
		INSERT INTO application_files (application_id, filename, description, content_type, file_data, file_size, uploaded_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, appID, filename, description, contentType, fileData, len(fileData), time.Now())

	if err != nil {
		logging.Logger.WithError(err).WithFields(map[string]interface{}{"app_id": appID, "filename": filename}).Error("Error adding file")
	}
	return err
}

func (s *DatabaseStorage) GetFile(appID, filename string) ([]byte, string, error) {
	var fileData []byte
	var contentType string

	err := s.db.QueryRow(`
		SELECT file_data, content_type 
		FROM application_files 
		WHERE application_id = ? AND filename = ?
	`, appID, filename).Scan(&fileData, &contentType)

	if err != nil {
		logging.Logger.WithError(err).WithFields(map[string]interface{}{"app_id": appID, "filename": filename}).Error("Error getting file")
		return nil, "", err
	}

	return fileData, contentType, nil
}

func (s *DatabaseStorage) DeleteFile(appID, filename string) error {
	_, err := s.db.Exec(`
		DELETE FROM application_files 
		WHERE application_id = ? AND filename = ?
	`, appID, filename)

	if err != nil {
		logging.Logger.WithError(err).WithFields(map[string]interface{}{"app_id": appID, "filename": filename}).Error("Error deleting file")
	}
	return err
}

func (s *DatabaseStorage) UpdateFileDescription(appID, filename, description string) error {
	_, err := s.db.Exec(`
		UPDATE application_files 
		SET description = ?
		WHERE application_id = ? AND filename = ?
	`, description, appID, filename)

	if err != nil {
		logging.Logger.WithError(err).WithFields(map[string]interface{}{"app_id": appID, "filename": filename}).Error("Error updating file description")
	}
	return err
}