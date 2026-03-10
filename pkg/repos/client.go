package repos

import (
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/repository"
)

func GetRepositories(repoOrOrg string) ([]*repository.Repository, error) {
	// try to parse a repo first
	if repo, err := repository.Parse(repoOrOrg); err == nil {
		return []*repository.Repository{&repo}, nil
	}

	// if parsing fails, treat it as an org and list repos
	args := []string{"api", fmt.Sprintf("orgs/%s/repos", repoOrOrg), "--paginate"}
	stdout, stderr, err := gh.Exec(args...)
	if err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("failed to list repositories: %s", stderr.String())
		}
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	var repos []*repository.Repository
	if err := json.Unmarshal(stdout.Bytes(), &repos); err != nil {
		return nil, fmt.Errorf("failed to parse repositories response: %w", err)
	}

	return repos, nil
}
