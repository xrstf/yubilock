package executor

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

type Executor func([]string, bool) (string, error)

func NewExecutor(username string, logger logrus.FieldLogger) (Executor, error) {
	uid := -1
	gid := -1

	if username != "" {
		user, err := user.Lookup(username)
		if err != nil {
			return nil, fmt.Errorf("could not lookup user '%s': %v", username, err)
		}

		newUid, _ := strconv.Atoi(user.Uid)
		newGid, _ := strconv.Atoi(user.Gid)

		if os.Getuid() != newUid || os.Getgid() != newGid {
			uid = newUid
			gid = newGid
			username = user.Username
		}
	}

	return func(command []string, wait bool) (string, error) {
		if len(command) == 0 {
			return "", errors.New("empty command given")
		}

		logger = logger.WithField("wait", wait)
		sysProcAttr := defineSysProcAttr(uid, gid, logger)

		if uid != -1 {
			logger.Infof("Running command (as '%s'): %v", username, command)
		} else {
			logger.Infof("Running command: %v", command)
		}

		if wait {
			return execCmd(command, sysProcAttr, logger, wait)
		}

		return execCmd(command, sysProcAttr, logger, wait)

		return execProcess(command, sysProcAttr, logger)
	}, nil
}

func defineSysProcAttr(uid int, gid int, logger logrus.FieldLogger) *syscall.SysProcAttr {
	sysProcAttr := &syscall.SysProcAttr{}

	if uid >= 0 && gid >= 0 {
		logger.Debugf("Switching credential to uid/gid %d/%d.", uid, gid)
		sysProcAttr.Credential = &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		}
	}

	return sysProcAttr
}

func execCmd(command []string, sysProcAttr *syscall.SysProcAttr, logger logrus.FieldLogger, wait bool) (string, error) {
	var buffer bytes.Buffer

	cmd := exec.Command(command[0], command[1:]...)

	cmd.Stdout = &buffer
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.SysProcAttr = sysProcAttr

	if wait {
		err := cmd.Run()
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(buffer.String()), nil
	}

	err := cmd.Start()
	if err != nil {
		return "", err
	}

	err = cmd.Process.Release()
	if err != nil {
		return "", err
	}

	return "", nil
}

func execProcess(command []string, sysProcAttr *syscall.SysProcAttr, logger logrus.FieldLogger) (string, error) {
	// sysProcAttr.Noctty = true
	// sysProcAttr.Setsid = true

	// stdin, _ := os.Open("/dev/null")
	// stdout, err := os.OpenFile("/home/xrstf/i3lock-out.log", os.O_APPEND, 0644)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to open stdout: %v", err)
	// }

	// stderr, err := os.OpenFile("/home/xrstf/i3lock-err.log", os.O_APPEND, 0644)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to open stderr: %v", err)
	// }

	var attr = os.ProcAttr{
		Env: os.Environ(),
		Sys: sysProcAttr,
		// Files: []*os.File{stdin, stdout, stderr},
		// Files: []*os.File{
		// 	os.Stdin,
		// 	os.Stdout,
		// 	os.Stderr,
		// },
	}

	process, err := os.StartProcess(command[0], command, &attr)
	if err != nil {
		return "", fmt.Errorf("failed to start process: %v", err)
	}

	err = process.Release()
	if err != nil {
		return "", err
	}

	time.Sleep(10 * time.Second)

	return "", nil
}
