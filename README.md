# yubilock

This program is meant to be executed by udev events to lock and unlock your
screen using whatever lockscreen software you are using. It makes use of
YubiKey's HMAC-SHA1 challenge-response mechanism and does not sorely rely
on USB device IDs or serial numbers.

## Installation

    $ go get github.com/xrstf/yubilock

You also need some standard YubiKey tools installed:

* `ykpamcfg` (Ubuntu: `libpam-yubico`)
* `ykchalresp` (Ubuntu: `yubikey-personalization`)

## Usage

Use the `ykpamcfg` tool to perform the initial challenge-response and record
the response in a state file in `/home/you/.yubico`:

    $ ykpamcfg -2

Afterwards, look into the directory mentioned above and find your
`challenge-...` file. You need it to configure yubilock.

Create a configuration file, e.g. in `/home/you/yubilock.yml`:

```yaml
# ykpamcfg-generated challenge file to use
stateFile: /home/you/.yubico/challenge-9736685

# command to execute when a YubiKey is removed and the
# screen is not already locked (as determined by the
# lockedCommand below)
lockCommand: ["/bin/bash", "-c", "DISPLAY=:0 /usr/local/bin/i3lock -n"]

# command to execute when a YubiKey is attached and it
# passed the challenge-response. This command is only
# executed if the lockedCommand return with code 0.
unlockCommand: ["pkill", "-1", "i3lock"]

# determines whether the screen is currently locked;
# return 0 = screen is locked
#        1 = screen is not locked
lockedCommand: ["pgrep", "i3lock"]

# in case the systemd service is not configured to run under
# your user account, you can make yubilock spawn the commands
# above explicitely under a given user.
#user: you
```

udev does not allow forking and will always reap orphan processes. To work
around this, we make use of a systemd service and just trigger that service.
Create a new service in `/etc/systemd/system/yubilock@.service`:

```
[Service]
Type=forking
User=you
ExecStart=/usr/local/bin/yubilock systemd-event /home/you/yubilock.yaml %I
```

We can now trigger the yubilock service by defining proper udev rules.
Configure the rules by creating a `/etc/udev/rules.d/90-yubikey.rules` file
with this content:

```
ACTION=="add", SUBSYSTEM=="usb", ENV{DEVTYPE}=="usb_device", ATTRS{product}=="YubiKey OTP+FIDO+CCID", RUN+="/bin/systemctl start yubilock@add.service"
ACTION=="remove", ATTRS{idVendor}=="1050", ATTRS{idProduct}=="0407", RUN+="/bin/systemctl start yubilock@remove.service"
```

That's it. Every time a YubiKey is plugged or unplugged, yubilock is being
executed and checks if the lockscreen has to be started or a challenge-reponse
run should be executed.

## FAQ

### Why is this not a long-running process?

Because I did not want to tie myself to any lifecycle management from the
init system. Just quickly handling USB events makes it easier to react to
cases where the lockscreen was started manually.

## License

MIT
