package main

import (
	"fmt"
	"net/http"

	"github.com/younsl/box/resume/references/job-tracker/pkg/app"
	"github.com/younsl/box/resume/references/job-tracker/pkg/config"
	"github.com/younsl/box/resume/references/job-tracker/pkg/logging"
)

func main() {
	cfg := config.Load()
	
	application := app.New(cfg)
	
	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))
	
	// Routes
	http.HandleFunc("/", application.HandleIndex)
	http.HandleFunc("/api/applications", application.HandleApplications)
	http.HandleFunc("/api/application", application.HandleApplication)
	http.HandleFunc("/api/save", application.HandleSave)
	http.HandleFunc("/api/sync", application.HandleGitSync)
	http.HandleFunc("/api/status", application.HandleStatus)
	http.HandleFunc("/api/upload", application.HandleFileUpload)
	http.HandleFunc("/api/download", application.HandleFileDownload)
	http.HandleFunc("/api/file/delete", application.HandleFileDelete)
	http.HandleFunc("/api/file/description", application.HandleFileDescription)

	logging.Logger.WithField("port", cfg.Port).WithField("url", fmt.Sprintf("http://localhost:%s", cfg.Port)).Info("Server starting")
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		logging.Logger.WithError(err).Fatal("Failed to start HTTP server")
	}
}