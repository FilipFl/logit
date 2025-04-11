package configuration

type MockConfigurationHandler struct {
	cfg *Config
	err error
}

func NewMockConfigurationHandler() *MockConfigurationHandler {
	return &MockConfigurationHandler{
		cfg: &Config{
			JiraOrigin:       "",
			JiraToken:        "",
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
}

func (h *MockConfigurationHandler) SetError(err error) {
	h.err = err
}

func (h *MockConfigurationHandler) GetToken() string {
	return h.cfg.JiraToken
}
