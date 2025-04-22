package configuration

import "time"

type Cfg struct {
	JiraOrigin       string            `json:"jira_origin"`
	JiraToken        string            `json:"jira_token"`
	JiraTokenEnvName string            `json:"jira_token_env_name"`
	JiraEmail        string            `json:"jira_email"`
	Aliases          map[string]string `json:"aliases"`
	Snapshot         *time.Time        `json:"snapshot"`
	TrustGitBranch   bool              `json:"trustGitBranch"`
}

type Config interface {
	GetToken() string
	GetJiraOrigin() string
	GetJiraEmail() string
	GetJiraTokenEnvName() string
	GetJiraToken() string
	GetAliases() map[string]string
	GetTrustGitBranch() bool
	GetSnapshot() *time.Time
	SetJiraOrigin(o string) error
	SetJiraEmail(email string) error
	SetJiraTokenEnvName(name string) error
	SetJiraToken(t string) error
	AddAlias(a, t string) error
	GetTaskFromAlias(a string) (string, error)
	RemoveAlias(a string) error
	SwapTrustGitBranch() error
	SetSnapshot(s *time.Time) error
}

const configDirectoryName = ".logit"
const configFileName = "config.json"
