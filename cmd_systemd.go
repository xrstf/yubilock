package main

import (
	"flag"

	"github.com/sirupsen/logrus"
	"github.com/xrstf/yubilock/executor"
)

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
