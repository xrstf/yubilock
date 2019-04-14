package main

import (
	"errors"
	"flag"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/xrstf/yubilock/executor"
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
