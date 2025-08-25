package models

import "time"

type FileInfo struct {
	Filename    string `json:"filename"`
	Description string `json:"description"`
	Size        int64  `json:"size"`
}

type JobApplication struct {
	ID           string    `json:"id"`
	Company      string    `json:"company"`
	Position     string    `json:"position"`
	Status       string    `json:"status"`
	FinalResult  string    `json:"final_result"`
	AppliedDate  string    `json:"applied_date"`
	Platform     string    `json:"platform"`
	URL          string    `json:"url"`
	Notes        string     `json:"notes"`
	Files        []FileInfo `json:"files"`
	LastModified time.Time  `json:"last_modified"`
}

type DataStore struct {
	Applications []JobApplication `json:"applications"`
	LastUpdated  time.Time        `json:"last_updated"`
}