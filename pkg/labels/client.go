package labels

import (
	"encoding/json"
	"fmt"

	gh "github.com/cli/go-gh/v2"
	repo "github.com/cli/go-gh/v2/pkg/repository"
)

// Label represents a GitHub label
type Label struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

func urlForRepoLabels(repo repo.Repository) string {
	return fmt.Sprintf("repos/%s/%s/labels", repo.Owner, repo.Name)
}

func urlForRepoLabel(repo repo.Repository, labelName string) string {
	return fmt.Sprintf("repos/%s/%s/labels/%s", repo.Owner, repo.Name, labelName)
}

// FindLabelByName fetches a label by name from a GitHub repository
func FindLabelByName(repo repo.Repository, labelName string) (*Label, error) {
	// Use GitHub CLI to make API request
	args := []string{"api", urlForRepoLabel(repo, labelName)}
	stdout, stderr, err := gh.Exec(args...)
	if err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("failed to fetch label: %s", stderr.String())
		}
		return nil, fmt.Errorf("failed to fetch label: %w", err)
	}

	// Parse the JSON response
	var label Label
	if err := json.Unmarshal(stdout.Bytes(), &label); err != nil {
		return nil, fmt.Errorf("failed to parse label response: %w", err)
	}

	return &label, nil
}

func AddLabelToRepo(repo repo.Repository, label Label, update bool) (bool, error) {
	// Check if the label already exists
	existingLabel, err := FindLabelByName(repo, label.Name)
	if err == nil {
		// Label already exists, update it if necessary
		if existingLabel.Color != label.Color || existingLabel.Description != label.Description {
			if update {
				updated, err := UpdateLabelInRepo(repo, label)
				return updated, err
			}
			return false, nil // No update needed
		}
		return false, nil // No update needed
	}
	// Label does not exist, create it
	args := []string{"api", urlForRepoLabels(repo), "--method", "POST", "-f", fmt.Sprintf("name=%s", label.Name), "-f", fmt.Sprintf("color=%s", label.Color), "-f", fmt.Sprintf("description=%s", label.Description)}
	_, stderr, err := gh.Exec(args...)
	if err != nil {
		if stderr.Len() > 0 {
			return false, fmt.Errorf("failed to create label: %s", stderr.String())
		}
		return false, fmt.Errorf("failed to create label: %w", err)
	}
	return true, nil
}

func UpdateLabelInRepo(repo repo.Repository, label Label) (bool, error) {
	args := []string{"api", urlForRepoLabel(repo, label.Name), "--method", "PATCH", "-f", fmt.Sprintf("color=%s", label.Color), "-f", fmt.Sprintf("description=%s", label.Description)}
	_, stderr, err := gh.Exec(args...)
	if err != nil {
		if stderr.Len() > 0 {
			return false, fmt.Errorf("failed to update label: %s", stderr.String())
		}
		return false, fmt.Errorf("failed to update label: %w", err)
	}
	return true, nil
}
