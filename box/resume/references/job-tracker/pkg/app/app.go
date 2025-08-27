package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/younsl/box/resume/references/job-tracker/pkg/config"
	"github.com/younsl/box/resume/references/job-tracker/pkg/git"
	"github.com/younsl/box/resume/references/job-tracker/pkg/logging"
	"github.com/younsl/box/resume/references/job-tracker/pkg/models"
	"github.com/younsl/box/resume/references/job-tracker/pkg/storage"
)

type App struct {
	config  *config.Config
	storage storage.StorageInterface
	git     *git.GitSync
}

func New(cfg *config.Config) *App {
	// Use SQLite database storage with GPG encryption
	dbFile := cfg.DataFile
	if !strings.HasSuffix(dbFile, ".db.gpg") {
		// Convert .json.gpg to .db.gpg
		if strings.HasSuffix(dbFile, ".json.gpg") {
			dbFile = strings.Replace(dbFile, ".json.gpg", ".db.gpg", 1)
		} else {
			dbFile = dbFile + ".db.gpg"
		}
	}
	
	store := storage.NewDatabaseStorage(dbFile, cfg.GPGRecipient)
	
	if err := store.Initialize(); err != nil {
		logging.Logger.WithError(err).Fatal("Failed to initialize database")
	}
	
	// Migrate from old JSON file if it exists
	oldJSONFile := strings.Replace(dbFile, ".db.gpg", ".json.gpg", 1)
	if oldJSONFile != dbFile {
		if _, err := os.Stat(oldJSONFile); err == nil {
			logging.Logger.WithField("file", oldJSONFile).Info("Found existing JSON file")
			
			if err := store.MigrateFromJSON(oldJSONFile); err != nil {
				logging.Logger.WithError(err).Error("Migration from JSON failed")
			} else {
				logging.Logger.Info("Migration successful!")
				
				// Save the migrated data
				if err := store.Save(); err != nil {
					logging.Logger.WithError(err).Error("Failed to save migrated data")
				} else {
					logging.Logger.WithField("file", dbFile).Info("Migrated data saved")
					
					// Backup old JSON file
					backupFile := oldJSONFile + ".backup"
					if err := os.Rename(oldJSONFile, backupFile); err != nil {
						logging.Logger.WithError(err).Error("Failed to backup old JSON file")
					} else {
						logging.Logger.WithField("backup_file", backupFile).Info("Old JSON file backed up")
					}
				}
			}
		}
	}

	return &App{
		config:  cfg,
		storage: store,
		git:     git.NewGitSync(),
	}
}

func (a *App) HandleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	tmpl.Execute(w, nil)
}

func (a *App) HandleApplications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a.storage.GetApplications())
}

func (a *App) HandleApplication(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST", "PUT":
		var app models.JobApplication
		if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
			logging.Logger.WithError(err).Error("JSON decode error")
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		
		logging.Logger.WithField("app_id", app.ID).WithField("company", app.Company).Debug("Received application")
		app.LastModified = time.Now()
		
		var operation string
		if r.Method == "PUT" {
			operation = "updated"
			if err := a.storage.UpdateApplication(app); err != nil {
				logging.Logger.WithError(err).WithField("app_id", app.ID).Error("Update error")
				http.Error(w, "Update failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			operation = "created"
			if err := a.storage.AddApplication(app); err != nil {
				logging.Logger.WithError(err).WithField("app_id", app.ID).Error("Add error")
				http.Error(w, "Add failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
		
		if err := a.storage.Save(); err != nil {
			logging.Logger.WithError(err).Error("Save error")
			http.Error(w, "Save failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		logging.Logger.WithFields(map[string]interface{}{
			"operation":    operation,
			"app_id":       app.ID,
			"company":      app.Company,
			"position":     app.Position,
			"status":       app.Status,
			"final_result": app.FinalResult,
			"applied_date": app.AppliedDate,
			"platform":     app.Platform,
			"url":          app.URL != "",
			"files_count":  len(app.Files),
			"has_notes":    app.Notes != "",
			"last_modified": app.LastModified.Format("2006-01-02 15:04:05"),
		}).Info("Application saved successfully")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(app)
		
	case "DELETE":
		id := r.URL.Query().Get("id")
		if err := a.storage.DeleteApplication(id); err != nil {
			logging.Logger.WithError(err).WithField("app_id", id).Error("Delete error")
			http.Error(w, "Delete failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		if err := a.storage.Save(); err != nil {
			logging.Logger.WithError(err).Error("Save error after delete")
			http.Error(w, "Save failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusOK)
	}
}

func (a *App) HandleSave(w http.ResponseWriter, r *http.Request) {
	if err := a.storage.Save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Data saved successfully"})
}

func (a *App) HandleGitSync(w http.ResponseWriter, r *http.Request) {
	if err := a.git.SyncToGit(a.config.DataFile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully synced with Git"})
}

func (a *App) HandleStatus(w http.ResponseWriter, r *http.Request) {
	applications := a.storage.GetApplications()
	
	var lastUpdated *time.Time
	if len(applications) > 0 {
		latest := applications[0].LastModified
		for _, app := range applications {
			if app.LastModified.After(latest) {
				latest = app.LastModified
			}
		}
		lastUpdated = &latest
	}
	
	// Get file size
	var fileSize string
	if stat, err := os.Stat(a.config.DataFile); err == nil {
		size := stat.Size()
		fileSize = formatFileSize(size)
	} else {
		fileSize = "Unknown"
	}
	
	status := map[string]interface{}{
		"gpg_recipient": a.config.GPGRecipient,
		"encrypted":     true,
		"data_file":     a.config.DataFile,
		"file_size":     fileSize,
		"last_updated":  lastUpdated,
		"total_apps":    len(applications),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

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

func (a *App) HandleFileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	appID := r.FormValue("app_id")
	if appID == "" {
		http.Error(w, "Application ID required", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Generate safe filename
	filename := strings.ReplaceAll(header.Filename, " ", "_")
	filename = filepath.Clean(filename)

	// Detect content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Store file in database
	dbStorage, ok := a.storage.(*storage.DatabaseStorage)
	if !ok {
		http.Error(w, "Database storage not available", http.StatusInternalServerError)
		return
	}

	if err := dbStorage.AddFile(appID, filename, description, contentType, fileData); err != nil {
		http.Error(w, "Error storing file in database", http.StatusInternalServerError)
		return
	}

	if err := a.storage.Save(); err != nil {
		http.Error(w, "Error saving data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "File uploaded successfully",
		"filename": filename,
	})
}

func (a *App) HandleFileDownload(w http.ResponseWriter, r *http.Request) {
	appID := r.URL.Query().Get("app_id")
	filename := r.URL.Query().Get("filename")

	if appID == "" || filename == "" {
		http.Error(w, "Application ID and filename required", http.StatusBadRequest)
		return
	}

	// Get file from database
	dbStorage, ok := a.storage.(*storage.DatabaseStorage)
	if !ok {
		http.Error(w, "Database storage not available", http.StatusInternalServerError)
		return
	}

	fileData, contentType, err := dbStorage.GetFile(appID, filename)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error retrieving file", http.StatusInternalServerError)
		}
		return
	}

	// Set headers and serve file
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileData)))
	
	w.Write(fileData)
}

func (a *App) HandleFileDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	appID := r.URL.Query().Get("app_id")
	filename := r.URL.Query().Get("filename")

	if appID == "" || filename == "" {
		http.Error(w, "Application ID and filename required", http.StatusBadRequest)
		return
	}

	// Delete file from database
	dbStorage, ok := a.storage.(*storage.DatabaseStorage)
	if !ok {
		http.Error(w, "Database storage not available", http.StatusInternalServerError)
		return
	}

	if err := dbStorage.DeleteFile(appID, filename); err != nil {
		http.Error(w, "Error deleting file", http.StatusInternalServerError)
		return
	}

	if err := a.storage.Save(); err != nil {
		http.Error(w, "Error saving data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "File deleted successfully"})
}

func (a *App) HandleFileDescription(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	appID := r.URL.Query().Get("app_id")
	filename := r.URL.Query().Get("filename")
	
	logging.Logger.WithFields(map[string]interface{}{"app_id": appID, "filename": filename}).Debug("HandleFileDescription called")

	if appID == "" || filename == "" {
		logging.Logger.WithFields(map[string]interface{}{"app_id": appID, "filename": filename}).Error("Missing parameters")
		http.Error(w, "Application ID and filename required", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		logging.Logger.WithError(err).Error("JSON decode error")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	logging.Logger.WithField("description", requestBody.Description).Debug("Updating description")

	// Update file description in database
	dbStorage, ok := a.storage.(*storage.DatabaseStorage)
	if !ok {
		logging.Logger.Error("Database storage not available")
		http.Error(w, "Database storage not available", http.StatusInternalServerError)
		return
	}

	if err := dbStorage.UpdateFileDescription(appID, filename, requestBody.Description); err != nil {
		logging.Logger.WithError(err).WithFields(map[string]interface{}{"app_id": appID, "filename": filename}).Error("Error updating file description")
		http.Error(w, "Error updating file description", http.StatusInternalServerError)
		return
	}

	if err := a.storage.Save(); err != nil {
		logging.Logger.WithError(err).Error("Error saving data")
		http.Error(w, "Error saving data", http.StatusInternalServerError)
		return
	}

	logging.Logger.WithFields(map[string]interface{}{"app_id": appID, "filename": filename}).Info("Description updated successfully")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Description updated successfully"})
}