package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	User          string   `yaml:"user"`
	StateFile     string   `yaml:"stateFile"`
	LockCommand   []string `yaml:"lockCommand"`
	UnlockCommand []string `yaml:"unlockCommand"`
	LockedCommand []string `yaml:"lockedCommand"`
	Verbose       bool     `yaml:"verbose"`
}

func LoadConfig(filename string) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	config := &Config{}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		return nil, fmt.Errorf("config file is not valid YAML: %v", err)
	}

	if config.StateFile == "" {
		return nil, fmt.Errorf("no stateFile configured")
	}

	return config, nil
}
