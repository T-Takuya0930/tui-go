package git

import (
	"encoding/json"
	"os/exec"
)

type GHRepo struct {
	Name          string `json:"name"`
	NameWithOwner string `json:"nameWithOwner"`
	SSHCloneURL   string `json:"sshUrl"`
	URL           string `json:"url"`
}

// IsGHCLIAvailable は gh cli がシステムに存在し利用可能か確認します。
func IsGHCLIAvailable() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

// FetchGHRepos は gh repo list を実行し、リポジトリ一覧を取得します。
func FetchGHRepos() ([]GHRepo, error) {
	if !IsGHCLIAvailable() {
		return nil, nil
	}

	cmd := exec.Command("gh", "repo", "list", "--limit", "100", "--json", "name,nameWithOwner,sshUrl,url")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var repos []GHRepo
	if err := json.Unmarshal(output, &repos); err != nil {
		return nil, err
	}

	return repos, nil
}
