package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type BasicConfig struct {
	cfg         *Cfg
	cfgFilePath string
}

func NewBasicConfig() *BasicConfig {
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Can't retrieve home directory. You're doomed. Reason: %s", err))
	}
	fullDirName := dirname + "/" + configDirectoryName
	fullConfigPath := fullDirName + "/" + configFileName
	config := &Cfg{Aliases: make(map[string]string)}
	basicConfig := &BasicConfig{cfg: config, cfgFilePath: fullConfigPath}
	_, err = os.Stat(fullDirName)
	if err != nil {
		err = os.Mkdir(fullDirName, 0777)
		if err != nil {
			panic(fmt.Sprintf("Error creating config directory: %s", err))
		}
	}
	file, err := os.Open(fullConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = basicConfig.SaveConfig(config)
			if err != nil {
				panic(fmt.Sprintf("the deepest panic of them all: %s", err))
			}
		} else {
			panic(fmt.Sprintf("seriously don't know what could cause this panic: %s", err))
		}
	} else {
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(config); err != nil {
			panic(fmt.Sprintf("some very serious looking error: %s", err))
		}

	}
	return basicConfig
}

func (h *BasicConfig) SaveConfig(config *Cfg) error {
	file, err := os.Create(h.cfgFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(config)
	if err != nil {
		return err
	}
	h.cfg = config
	return nil
}

func (h *BasicConfig) persistCfg() error {
	file, err := os.Create(h.cfgFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(h.cfg)
	if err != nil {
		return err
	}
	return nil
}

func (h *BasicConfig) GetToken() string {
	if h.cfg.JiraTokenEnvName != "" {
		return os.Getenv(h.cfg.JiraTokenEnvName)
	}
	return h.cfg.JiraToken
}

func (h *BasicConfig) GetJiraEmail() string {
	return h.cfg.JiraEmail
}

func (h *BasicConfig) GetJiraOrigin() string {
	return h.cfg.JiraOrigin
}

func (h *BasicConfig) GetJiraToken() string {
	return h.cfg.JiraToken
}

func (h *BasicConfig) GetJiraTokenEnvName() string {
	return h.cfg.JiraTokenEnvName
}

func (h *BasicConfig) GetTrustGitBranch() bool {
	return h.cfg.TrustGitBranch
}

func (h *BasicConfig) GetAliases() map[string]string {
	return h.cfg.Aliases
}

func (h *BasicConfig) GetSnapshot() *time.Time {
	return h.cfg.Snapshot
}

func (h *BasicConfig) SetJiraEmail(email string) error {
	h.cfg.JiraEmail = email
	return h.persistCfg()
}

func (h *BasicConfig) SetJiraOrigin(o string) error {
	h.cfg.JiraOrigin = o
	return h.persistCfg()
}
func (h *BasicConfig) SetJiraToken(t string) error {
	h.cfg.JiraToken = t
	return h.persistCfg()
}

func (h *BasicConfig) SetJiraTokenEnvName(name string) error {
	h.cfg.JiraTokenEnvName = name
	return h.persistCfg()
}

func (h *BasicConfig) SwapTrustGitBranch() error {
	h.cfg.TrustGitBranch = !h.cfg.TrustGitBranch
	return h.persistCfg()
}

func (h *BasicConfig) AddAlias(a, t string) error {
	h.cfg.Aliases[a] = t
	return h.persistCfg()
}

func (h *BasicConfig) GetTaskFromAlias(a string) (string, error) {
	if val, exists := h.cfg.Aliases[a]; exists {
		return val, nil
	}
	return "", ErrorAliasDontExists
}

func (h *BasicConfig) RemoveAlias(a string) error {
	if _, exists := h.cfg.Aliases[a]; exists {
		delete(h.cfg.Aliases, a)
		return h.persistCfg()
	}
	return ErrorAliasDontExists
}

func (h *BasicConfig) SetSnapshot(s *time.Time) error {
	h.cfg.Snapshot = s
	return h.persistCfg()
}
