package main

import (
	"errors"
	"flag"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/xrstf/yubilock/executor"
	"github.com/xrstf/yubilock/yubikey"
)

func main() {
	var err error

	flag.Parse()

	command := flag.Arg(0)
	logger := newLogger(false)

	config, err := loadConfig()
	if err != nil {
		logger.Fatalf("Failure: %v", err)
	}

	if config.Verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	cmdLogger := logger.WithField("cmd", command)

	executor, err := executor.NewExecutor(config.User, cmdLogger)
	if err != nil {
		cmdLogger.Fatalf("Failure: %v", err)
	}

	switch command {
	case "lock":
		err = lockCommand(config, cmdLogger, executor)
	case "unlock":
		err = unlockCommand(config, cmdLogger, executor)
	case "systemd-event":
		err = systemdEventCommand(config, cmdLogger, executor)
	default:
		cmdLogger.Fatalf("Usage: yubilock [lock|unlock|systemd-devent] CONFIG_FILE")
	}

	if err != nil {
		cmdLogger.Fatalf("Failure: %v", err)
	}
}

func newLogger(verbose bool) *logrus.Logger {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05 MST",
	}

	if verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	return logger
}

func loadConfig() (*Config, error) {
	filename := flag.Arg(1)
	if filename == "" {
		return nil, errors.New("no configuration file given")
	}

	config, err := LoadConfig(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %v", err)
	}

	return config, nil
}

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

func systemdEventCommand(config *Config, logger logrus.FieldLogger, executor executor.Executor) error {
	action := flag.Arg(2)

	logger.Debugf("Handling systemd event '%s'...", action)

	switch action {
	case "add":
		logger.Info("YubiKey was attached!")
		err := unlockCommand(config, logger.WithField("subcmd", "add"), executor)
		if err != nil {
			return err
		}

	case "remove":
		if tokenAttached() {
			logger.Debug("YubiKey temporarily disappeared, possibly doing a challenge-response operation.")
			return nil
		}

		logger.Info("YubiKey was removed!")
		return lockCommand(config, logger.WithField("subcmd", "remove"), executor)
	}

	return nil
}

func isLocked(config *Config, logger logrus.FieldLogger, executor executor.Executor) bool {
	if len(config.LockedCommand) == 0 {
		logger.Debugf("No lockedCommand configured, assuming lock process can handle multiple invocations. Good luck!")
		return false
	}

	_, err := executor(config.LockedCommand, true)

	return err == nil
}

func tokenAttached() bool {
	output, _ := exec.Command("lsusb").CombinedOutput()

	return strings.Contains(string(output), "Yubico.com")
}
