package dbus

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

// ToggleMediaPlayback uses the D-Bus session bus to
// trigger the PlayPause method on org.mpris.MediaPlayer2.Player.
//
// The name should be of the format org.mpris.MediaPlayer2.xxxx
// where xxx is the music player. For example, to send the command
// to Spotify: org.mpris.MediaPlayer2.spotify.
//
// For convenience, any name passed that does not begin with
// org.mpris.MediaPlayer2 will have it prepended.
//
// You can use a D-Bus explorer like D-Feet to look on your
// session bus and figure out the name. Search for MPRIS. The spec
// itself can be found here: https://specifications.freedesktop.org/mpris-spec/latest/.
func ToggleMediaPlayback(
	ctx context.Context,
	name string,
) error {
	conn, err := dbus.ConnectSessionBus(dbus.WithContext(ctx))
	if err != nil {
		return err
	}
	defer conn.Close()

	path := "/org/mpris/MediaPlayer2"
	method := "org.mpris.MediaPlayer2.Player.PlayPause"

	if !strings.HasPrefix(name, "org.mpris.MediaPlayer2") {
		name = fmt.Sprintf("org.mpris.MediaPlayer2.%s", name)
	}

	bus := conn.Object(name, dbus.ObjectPath(path))

	tctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	res := bus.CallWithContext(tctx, method, dbus.FlagNoReplyExpected)
	return res.Err
}
