# yubilock

This program is meant to be executed by udev events to lock and unlock your
screen using whatever lockscreen software you are using. It makes use of
YubiKey's HMAC-SHA1 challenge-response mechanism and does not sorely rely
on USB device IDs or serial numbers.

## Usage

    $ go get github.com/xrstf/yubilock

Make sure you have `libpam-yubico` (or equivalent on your distro) installed
and then use the `ykpamcfg` tool to perform the initial challenge-response
and record the response in a state fule in `/home/you/.yubico`:

    $ ykpamcfg -2

Afterwards, look into the directory mentioned above and find your
`challenge-...` file. You need it to configure yubilock.

Create a configuration file, e.g. in `/home/you/yubilock.yml`:

```yaml
stateFile: /home/you/.yubico/challenge-9736685
serialIDs: ["Yubico_YubiKey_OTP+FIDO+CCID"]
lockCommand: ["/home/you/.config/i3/lock.sh"]
unlockCommand: ["/bin/bash", "-c", "pkill -1 i3lock || true"]
lockedCommand: ["/bin/bash", "-c", "pgrep -u xrstf i3lock"]
```

Configure udev rules by creating a `/etc/udev/rules.d/90-yubikey.rules` file
with this content:

    ATTR{product}!="YubiKey OTP+FIDO+CCID", GOGO="yubikey_end"
    ACTION=="remove", RUN+="/home/you/go/bin/yubilock udev-event /home/you/yubilock.yaml"
    ACTION=="add", RUN+="/home/you/go/bin/yubilock udev-event /home/you/yubilock.yaml"
    LABEL="yubikey_end"

That's it. Every time a YubiKey is plugged or unplugged, yubilock is being
executed and checks if the lockscreen has to be started or a challenge-reponse
run should be executed.

## License

MIT
