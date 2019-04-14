package yubikey

import (
	"encoding/hex"
	"fmt"

	"github.com/xrstf/yubilock/executor"
)

func ChallengeResponse(challenge []byte, slot int, runner executor.Executor) (string, error) {
	challengeHex := hex.EncodeToString(challenge)

	cmd := []string{
		"ykchalresp",
		fmt.Sprintf("-%d", slot),
		"-H",
		"-x",
		challengeHex,
	}

	return runner(cmd, true)
}
