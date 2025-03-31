package configuration

import "time"

type Config struct {
	JiraHost  string            `json:"jira_host"`
	JiraToken string            `json:"jira_token"`
	JiraEmail string            `json:"jira_email"`
	Aliases   map[string]string `json:"aliases"`
	Snapshot  *time.Time        `json:"snapshot"`
}

type ConfigurationHandler interface {
	LoadConfig() *Config
	SaveConfig(config *Config) error
}

const configDirectoryName = ".logit"
const configFileName = "config.json"
