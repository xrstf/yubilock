package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/xrstf/yubilock/executor"
	"github.com/xrstf/yubilock/yubikey"
)

func unlockCommand(config *Config, logger logrus.FieldLogger, executor executor.Executor) error {
	if !isLocked(config, logger, executor) {
		logger.Debug("No lock process detected.")
		return nil
	}

	logger.Debug("Preparing challenge-response...")

	challenge, err := yubikey.NewChallengeFromFile(config.StateFile)
	if err != nil {
		return fmt.Errorf("failed to read challenge state: %v", err)
	}

	logger.Debug("Posing challenge...")
	valid, err := challenge.Execute(executor)
	if err != nil {
		return fmt.Errorf("failed to validate response: %v", err)
	}

	if !valid {
		return fmt.Errorf("invalid response provided, rejecting YubiKey")
	}

	logger.Debug("Valid response provided.")

	command := config.UnlockCommand
	if len(command) == 0 {
		logger.Debug("No unlock command configured, exiting.")
		return nil
	}

	if _, err := executor(command, true); err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}

	logger.Info("Unlocking completed successfully.")

	return nil
}
