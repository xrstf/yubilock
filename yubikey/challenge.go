package yubikey

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/xrstf/yubilock/executor"

	"golang.org/x/crypto/pbkdf2"
)

type Challenge struct {
	Challenge  []byte
	hash       []byte
	salt       []byte
	iterations int
	Slot       int
}

var challengeFileRegex = regexp.MustCompile(`^([a-z0-9]+):([0-9a-f]+):([0-9a-f]+):([0-9a-f]+):([0-9]+):([0-9])$`)

func NewChallengeFromFile(filename string) (*Challenge, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	match := challengeFileRegex.FindStringSubmatch(strings.TrimSpace(string(content)))
	if match == nil || match[1] != "v2" {
		return nil, errors.New("file does not seem to be a v2-encoded challenge state")
	}

	challenge, err := hex.DecodeString(match[2])
	if err != nil {
		return nil, errors.New("challenge is not a valid hex string")
	}

	hash, err := hex.DecodeString(match[3])
	if err != nil {
		return nil, errors.New("hash is not a valid hex string")
	}

	salt, err := hex.DecodeString(match[4])
	if err != nil {
		return nil, errors.New("salt is not a valid hex string")
	}

	iterations, err := strconv.Atoi(match[5])
	if err != nil {
		return nil, errors.New("iterations is not a valid number")
	}

	slot, err := strconv.Atoi(match[6])
	if err != nil {
		return nil, errors.New("slot number is not a valid number")
	}

	return &Challenge{
		Challenge:  challenge,
		hash:       hash,
		salt:       salt,
		iterations: iterations,
		Slot:       slot,
	}, nil
}

func (c *Challenge) Execute(executor executor.Executor) (bool, error) {
	response, err := ChallengeResponse(c.Challenge, c.Slot, executor)
	if err != nil {
		return false, fmt.Errorf("challenge-response failed: %v", err)
	}

	return c.validateResponse(response)
}

func (c *Challenge) validateResponse(response string) (bool, error) {
	hash, err := c.hashResponse(response)
	if err != nil {
		return false, fmt.Errorf("failed to hash response: %v", err)
	}

	return bytes.Equal(c.hash, hash), nil
}

func (c *Challenge) hashResponse(response string) ([]byte, error) {
	hash := pbkdf2.Key(
		[]byte(response),
		c.salt,
		c.iterations,
		20,
		sha1.New,
	)

	return hash, nil
}
