package configuration

import "time"

type Config struct {
	JiraHost         string            `json:"jira_host"`
	JiraToken        string            `json:"jira_token"`
	JiraEmail        string            `json:"jira_email"`
	JiraTokenEnvName string            `json:"jira_token_env_name"`
	Aliases          map[string]string `json:"aliases"`
	Snapshot         *time.Time        `json:"snapshot"`
}

type ConfigurationHandler interface {
	LoadConfig() *Config
	SaveConfig(config *Config) error
	GetToken() string
}

const configDirectoryName = ".logit"
const configFileName = "config.json"
