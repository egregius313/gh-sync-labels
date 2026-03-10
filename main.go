package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cli/go-gh/v2/pkg/repository"
	lbl "github.com/egregius313/gh-sync-labels/pkg/labels"
	repo "github.com/egregius313/gh-sync-labels/pkg/repos"
)

var (
	fromRepo       string
	labels         []string
	toRepos        []string
	verbose        bool
	updateIfExists bool
)

func verbosePrintf(format string, v ...interface{}) {
	if verbose {
		fmt.Fprintf(os.Stderr, format, v...)
	}
}

type Response struct {
	Repository string `json:"repository"`
	Label      string `json:"label"`
	Success    bool   `json:"success"`
	Message    string `json:"message,omitempty"`
}

var rootCmd = &cobra.Command{
	Use:   "gh sync-labels",
	Short: "Sync GitHub labels across repositories",
	Long:  "Synchronize labels from one GitHub repository to one or more target repositories",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate required flags
		if len(labels) == 0 {
			return fmt.Errorf("at least one --label flag is required")
		}
		if len(toRepos) == 0 {
			return fmt.Errorf("at least one --to flag is required")
		}

		// Get source repository (use current if not specified)
		var sourceRepo repository.Repository
		var err error
		if fromRepo == "" {
			sourceRepo, err = repository.Current()
			if err != nil {
				return fmt.Errorf("failed to get current repository: %w", err)
			}
			verbosePrintf("Using current repository: %s/%s\n", sourceRepo.Owner, sourceRepo.Name)
		} else {
			sourceRepo, err = repository.Parse(fromRepo)
			if err != nil {
				return fmt.Errorf("failed to parse source repository: %w", err)
			}
		}

		verbosePrintf("Syncing labels from %s/%s to %v\n", sourceRepo.Owner, sourceRepo.Name, toRepos)
		verbosePrintf("Labels to sync: %v\n", labels)

		targetRepos := make([]*repository.Repository, 0)

		for _, toRepo := range toRepos {
			repos, err := repo.GetRepositories(toRepo)
			if err != nil {
				verbosePrintf("Error fetching target repositories for '%s': %v\n", toRepo, err)
				continue
			}
			targetRepos = append(targetRepos, repos...)
		}

		if responses, err := SynchronizeLabels(&sourceRepo, labels, targetRepos); err != nil {
			return err
		} else {
			if jsonData, err := json.Marshal(responses); err != nil {
				return fmt.Errorf("failed to marshal responses: %w", err)
			} else {
				fmt.Println(string(jsonData))
			}
		}
		return nil
	},
}

func SynchronizeLabels(sourceRepo *repository.Repository, labels []string, repositories []*repository.Repository) ([]Response, error) {
	responses := make([]Response, 0)

	for _, label := range labels {
		verbosePrintf("Processing label: %s\n", label)

		sourceLabel, err := lbl.FindLabelByName(*sourceRepo, label)
		if err != nil {
			verbosePrintf("Error fetching label '%s' from source repository '%s/%s': %v\n", label, sourceRepo.Owner, sourceRepo.Name, err)
			for _, targetRepo := range repositories {
				responses = append(responses, Response{
					Repository: fmt.Sprintf("%s/%s", targetRepo.Owner, targetRepo.Name),
					Label:      label,
					Success:    false,
					Message:    fmt.Sprintf("failed to fetch source label: %v", err),
				})
			}
			continue
		}

		for _, targetRepo := range repositories {
			verbosePrintf("Syncing label '%s' to repository: %s/%s\n", label, targetRepo.Owner, targetRepo.Name)
			updated, err := lbl.AddLabelToRepo(*targetRepo, *sourceLabel, updateIfExists)
			if err != nil {
				verbosePrintf("Error syncing label '%s' to repository '%s/%s': %v\n", label, targetRepo.Owner, targetRepo.Name, err)
				responses = append(responses, Response{
					Repository: fmt.Sprintf("%s/%s", targetRepo.Owner, targetRepo.Name),
					Label:      label,
					Success:    false,
					Message:    err.Error(),
				})
				continue
			}
			if updated {
				verbosePrintf("Label '%s' updated in repository '%s/%s'\n", label, targetRepo.Owner, targetRepo.Name)
				responses = append(responses, Response{
					Repository: fmt.Sprintf("%s/%s", targetRepo.Owner, targetRepo.Name),
					Label:      label,
					Success:    true,
					Message:    "updated",
				})
			} else {
				verbosePrintf("Label '%s' created in repository '%s/%s'\n", label, targetRepo.Owner, targetRepo.Name)
				responses = append(responses, Response{
					Repository: fmt.Sprintf("%s/%s", targetRepo.Owner, targetRepo.Name),
					Label:      label,
					Success:    true,
					Message:    "created",
				})
			}
		}
	}

	return responses, nil
}

func init() {
	rootCmd.Flags().StringVar(&fromRepo, "from", "", "Source repository (defaults to current repository)")
	rootCmd.Flags().StringSliceVar(&labels, "label", []string{}, "Labels to sync (can be specified multiple times)")
	rootCmd.Flags().StringSliceVar(&toRepos, "to", []string{}, "Target repositories (can be specified multiple times)")
	rootCmd.Flags().BoolVarP(&updateIfExists, "update", "u", false, "Update label if it already exists in target repository")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
