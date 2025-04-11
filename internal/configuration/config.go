package configuration

import "time"

type Config struct {
	JiraOrigin       string            `json:"jira_origin"`
	JiraToken        string            `json:"jira_token"`
	JiraTokenEnvName string            `json:"jira_token_env_name"`
	Aliases          map[string]string `json:"aliases"`
	Snapshot         *time.Time        `json:"snapshot"`
	TrustGitBranch   bool              `json:"trustGitBranch"`
}

type ConfigurationHandler interface {
	LoadConfig() *Config
	SaveConfig(config *Config) error
	GetToken() string
}

const configDirectoryName = ".logit"
const configFileName = "config.json"
