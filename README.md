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
lockCommand: ["/bin/bash", "-c", "DISPLAY=:0 /usr/local/bin/i3lock -n"]
unlockCommand: ["pkill", "-1", "i3lock"]
lockedCommand: ["pgrep", "i3lock"]
user: you
```

udev does not allow forking and will always reap orphan processes. To work
around this, we make use of a systemd service and just trigger that service.
Create a new service in `/etc/systemd/system/yubilock@.service`:

```
[Service]
Type=forking
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

## License

MIT
