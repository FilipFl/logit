package configuration

import "time"

type MockConfig struct {
	config *Cfg
	err    error
}

func NewMockConfig(config *Cfg) *MockConfig {
	if config != nil {
		return &MockConfig{
			config: config,
			err:    nil,
		}
	}
	return &MockConfig{
		config: &Cfg{
			JiraOrigin:       "",
			JiraToken:        "",
			JiraTokenEnvName: "",
			Aliases:          map[string]string{},
			Snapshot:         nil,
		},
		err: nil,
	}
}

func (h *MockConfig) SetConfig(config *Cfg) {
	h.config = config
}

func (h *MockConfig) SetError(err error) {
	h.err = err
}

func (h *MockConfig) GetToken() string {
	return h.config.JiraToken
}

func (h *MockConfig) GetJiraEmail() string {
	return h.config.JiraEmail
}

func (h *MockConfig) GetJiraOrigin() string {
	return h.config.JiraOrigin
}

func (h *MockConfig) GetJiraToken() string {
	return h.config.JiraToken
}

func (h *MockConfig) GetJiraTokenEnvName() string {
	return h.config.JiraTokenEnvName
}

func (h *MockConfig) GetTrustGitBranch() bool {
	return h.config.TrustGitBranch
}

func (h *MockConfig) GetAliases() map[string]string {
	return h.config.Aliases
}

func (h *MockConfig) GetSnapshot() *time.Time {
	return h.config.Snapshot
}

func (h *MockConfig) SetJiraEmail(email string) error {
	return h.err
}

func (h *MockConfig) SetJiraOrigin(o string) error {
	return h.err
}
func (h *MockConfig) SetJiraToken(t string) error {
	return h.err
}

func (h *MockConfig) SetJiraTokenEnvName(name string) error {
	return h.err
}

func (h *MockConfig) SwapTrustGitBranch() error {
	return h.err
}

func (h *MockConfig) AddAlias(a, t string) error {
	return h.err
}

func (h *MockConfig) GetTaskFromAlias(a string) (string, error) {
	if val, exists := h.config.Aliases[a]; exists {
		return val, nil
	}
	return "", h.err
}

func (h *MockConfig) RemoveAlias(a string) error {
	return h.err
}

func (h *MockConfig) SetSnapshot(s *time.Time) error {
	return h.err
}
