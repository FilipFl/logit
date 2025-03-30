package configuration

import (
	"encoding/json"
	"fmt"
	"os"
)

type BasicConfigurationHandler struct {
	cfg         *Config
	cfgFilePath string
}

func NewBasicConfigurationHandler() *BasicConfigurationHandler {
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Can't retrieve home directory. You're doomed. Reason: %s", err))
	}
	fullDirName := dirname + "/" + configDirectoryName
	fullConfigPath := fullDirName + "/" + configFileName
	config := &Config{Aliases: make(map[string]string)}
	basicConfig := &BasicConfigurationHandler{cfg: config, cfgFilePath: fullConfigPath}
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

func (h *BasicConfigurationHandler) LoadConfig() *Config {
	return h.cfg
}

func (h *BasicConfigurationHandler) SaveConfig(config *Config) error {
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
