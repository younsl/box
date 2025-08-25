package git

import (
	"fmt"
	"os/exec"
	"time"
)

type GitSync struct{}

func NewGitSync() *GitSync {
	return &GitSync{}
}

func (g *GitSync) SyncToGit(dataFile string) error {
	commands := [][]string{
		{"git", "add", dataFile},
		{"git", "commit", "-m", fmt.Sprintf("Update job applications - %s", time.Now().Format("2006-01-02 15:04"))},
		{"git", "push"},
	}

	for _, cmd := range commands {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			if cmd[1] == "commit" {
				continue
			}
			return fmt.Errorf("git sync failed at %s: %w", cmd[1], err)
		}
	}

	return nil
}