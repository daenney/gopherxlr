package ipc

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"strconv"
	"time"
)

const (
	GoXLRSocket      = "/tmp/goxlr.socket"
	GetStatusCommand = `"GetStatus"`
)

// DaemonStatus is a partial struct representing the data
// returned by the GetStatus command.
//
// It only represents the part of the response necessary
// to construct the HTTP API address.
type DaemonStatus struct {
	Status struct {
		Config struct {
			HTTPSettings struct {
				Enabled     bool   `json:"enabled"`
				BindAddress string `json:"bind_address"`
				Port        int    `json:"port"`
			} `json:"http_settings"`
		} `json:"config"`
	} `json:"Status"`
}

// GetAddress returns a host:port for the HTTP API.
//
// If the HTTP API is not enabled, the second return
// value will be false.
func (d DaemonStatus) GetAddress() (string, bool) {
	if !d.Status.Config.HTTPSettings.Enabled {
		return "", false
	}
	return net.JoinHostPort(
			d.Status.Config.HTTPSettings.BindAddress,
			strconv.Itoa(d.Status.Config.HTTPSettings.Port)),
		true
}

// DialSocket dials the GoXLR socket.
func DialSocket(ctx context.Context) (net.Conn, error) {
	var d net.Dialer
	d.Timeout = 5 * time.Second

	conn, err := d.DialContext(ctx, "unix", GoXLRSocket)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// PackRequest will take any string payload and encoded it
// in the format the daemon expects.
//
// The msg should be valid JSON and be one of [the supported
// commands].
//
// [the supported commands]: https://github.com/GoXLR-on-Linux/goxlr-utility/blob/ab9183e027455b104e8470cd3b4ce295dc6755dc/ipc/src/lib.rs#L20
func PackRequest(msg json.RawMessage) ([]byte, error) {
	var b bytes.Buffer
	if err := binary.Write(&b, binary.BigEndian, uint32(len(msg))); err != nil {
		return nil, err
	}
	if _, err := b.Write(msg); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// UnpackResponse will read a response from the Reader.
//
// Beware that this can hang indefinitely if you pass
// a net.Conn and the command you sent is unknown to
// the daemon. It won't respond with an error message,
// it simply doesn't respond at all. Make sure to manage
// the net.Conn.ReadDeadline so that a read will
// eventually time out.
func UnpackResponse(r io.Reader) (json.RawMessage, error) {
	// payload length is encoded as a 4-byte uint32 in BigEndian
	len := make([]byte, 4)
	_, err := r.Read(len)
	if err != nil {
		return nil, err
	}

	iLen := int(binary.BigEndian.Uint32(len))
	data := make([]byte, iLen)

	i := 0
	for i < iLen {
		num, err := r.Read(data[i:])
		if err != nil {
			return nil, err
		}
		i += num
	}

	return data, nil
}

// MustGetAddress returns the address the HTTP API is
// bound on.
//
// This function will panic if any error occurs, or
// if the HTTP API is not enabled.
func MustGetAddress(ctx context.Context) string {
	conn, err := DialSocket(ctx)
	if err != nil {
		panic(err)
	}

	req, _ := PackRequest([]byte(GetStatusCommand))

	_, err = conn.Write(req)
	if err != nil {
		panic(err)
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	resp, err := UnpackResponse(conn)
	if err != nil {
		panic(err)
	}

	conn.Close()

	var status DaemonStatus
	if err := json.Unmarshal(resp, &status); err != nil {
		panic(err)
	}

	addr, enabled := status.GetAddress()
	if !enabled {
		panic("not enabled")
	}

	return addr
}
