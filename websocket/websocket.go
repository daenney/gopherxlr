package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"golang.org/x/net/websocket"
)

func URLForAPI(hostPort string) string {
	return fmt.Sprintf("ws://%s/api/websocket", hostPort)
}

func DialAPI(ctx context.Context, url string) (*websocket.Conn, error) {
	cfg, _ := websocket.NewConfig(
		url, "http://localhost",
	)
	return cfg.DialContext(ctx)
}

func Listen(ctx context.Context, hostPort string, ch chan<- StatusChange) {
	ws, err := DialAPI(ctx, URLForAPI(hostPort))
	if err != nil {
		panic(err)
	}
	defer ws.Close()

	statusID := rand.Int32()

	err = websocket.JSON.Send(ws, map[string]any{"id": statusID, "data": "GetStatus"})
	if err != nil {
		panic(err)
	}

	var status json.RawMessage

InitLoop:
	for {
		select {
		case <-ctx.Done():
			close(ch)
			return
		default:
			var sr StatusResponse
			if err := websocket.JSON.Receive(ws, &sr); err != nil {
				// if it doesn't decode as a StatusResponse, ignore it
				continue
			}

			if sr.ID != int(statusID) {
				continue
			}

			status = sr.Data.Status
			break InitLoop
		}
	}

UpdateLoop:
	for {
		select {
		case <-ctx.Done():
			close(ch)
			break UpdateLoop
		default:
			type dataPatch struct {
				Data struct {
					// We're making an assumption here that GoXLR-Utility always
					// sends out valid JSON Patches, so we decode directly into
					// jsonpatch.Patch instead of calling jsonpatch.DecodePatch
					Patch jsonpatch.Patch `json:"Patch"`
				} `json:"data"`
			}
			var p dataPatch

			if err := websocket.JSON.Receive(ws, &p); err != nil {
				panic(err)
			}

			patches := p.Data.Patch
			var oldStatus, newStatus Status
			if err := json.Unmarshal(status, &oldStatus); err != nil {
				panic(err)
			}

			status, err = patches.Apply(status)
			if err != nil {
				panic(err)
			}

			if err := json.Unmarshal(status, &newStatus); err != nil {
				panic(err)
			}
			ch <- StatusChange{
				Old: oldStatus,
				New: newStatus,
			}
		}
	}
}
