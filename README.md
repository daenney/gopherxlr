# GopherXLR

> [!CAUTION]
> GopherXLR is still under development and a bit experimental. It should work and since it doesn't control the GoXLR itself should be entirely harmless to it.

A Go library and daemon to interact with the [GoXLR-Utility][gu] application, for the TC-Helicon GoXLR or GoXLR Mini, through its HTTP API.

[gu]: https://github.com/GoXLR-on-Linux/goxlr-utility

You can use tiny scripts written in [Expr][expr] to automatically perform actions when the mixer state changes. Each script is passed the previous and current state and can decide if it wants to do something.

[expr]: https://expr-lang.org/

## Example

The reason I wrote this tool is to be able to toggle playback in Spotify when I mute/unmute a certain channel on a mixer. The scripts in [examples](./examples) show how to use it. Make sure to replace the `MIXER_ID` and set the right value for `fader`.

## Functions

These are the functions you can call from your scripts:

| Function   | Arguments   | Example                | Purpose   |
| :--------- | :---------- | :--------------------- | :-------- |
| PlayPause[^1]  | player name | `PlayPause("spotify")` | Toggle music play/pause |
| RunCommand | command     | `RunCommand("notify-send", "title", "body")` | Execute arbitrary command |

[^1]: `PlayPause` uses the D-Bus user session and as such only works on Linux

## Running

A systemd unit file is included. You can install this in `$HOME/.config/systemd/user` or `/lib/systemd/user`. After a `daemon-reload` you can start it with `systemctl --user start gopherxlr.service`.

If you also `systemctl --user enable` it, it'll start automatically when you login.

The unit file has an `After=` on `app-goxlr\x2ddaemon@autostart.service`, which is automatically created by the `systemd-xdg-autostart-generator` user generator. This should ensure it's not started too early, and not started at all if goxlr-daemon isn't set to autostart.

The unit file expects to find the binary in `$HOME/.local/bin` and the scripts in `$HOME/.config/gopherxlr`. You can adjust the paths in the unit file.
