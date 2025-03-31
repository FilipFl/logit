package configuration

type MockConfigurationHandler struct {
	cfg *Config
	err error
}

func NewMockConfigurationHandler() *MockConfigurationHandler {
	return &MockConfigurationHandler{
		cfg: &Config{
			JiraHost:         "",
			JiraToken:        "",
			JiraEmail:        "",
			JiraTokenEnvName: "",
			Aliases:          map[string]string{},
			Snapshot:         nil,
		},
		err: nil,
	}
}

func (h *MockConfigurationHandler) LoadConfig() *Config {
	return h.cfg
}

func (h *MockConfigurationHandler) SaveConfig(cfg *Config) error {
	return h.err
}

func (h *MockConfigurationHandler) SetConfig(cfg *Config) {
	h.cfg = cfg
	return
}

func (h *MockConfigurationHandler) SetError(err error) {
	h.err = err
	return
}

func (h *MockConfigurationHandler) GetToken() string {
	return h.cfg.JiraToken
}
