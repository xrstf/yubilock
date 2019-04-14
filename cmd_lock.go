package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/xrstf/yubilock/executor"
)

func lockCommand(config *Config, logger logrus.FieldLogger, executor executor.Executor) error {
	if isLocked(config, logger, executor) {
		logger.Debug("System is already locked.")
		return nil
	}

	command := config.LockCommand
	if len(command) == 0 {
		logger.Debug("No lock command configured, exiting.")
		return nil
	}

	if _, err := executor(command, false); err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}

	logger.Info("Locking successful.")

	return nil
}
