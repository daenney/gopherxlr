package websocket

import "encoding/json"

type StatusResponse struct {
	ID   int `json:"id"`
	Data struct {
		Status json.RawMessage `json:"Status"`
	} `json:"data"`
}

type Status struct {
	Mixers map[string]struct {
		FaderStatus map[string]struct {
			MuteState string `json:"mute_state"`
		} `json:"fader_status"`
	} `json:"mixers"`
}

type StatusChange struct {
	Old, New Status
}
