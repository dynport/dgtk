package dockerclient

import (
	"encoding/json"
	"fmt"
	"io"
)

type JSONMessage struct {
	Stream   string        `json:"stream,omitempty"`
	Status   string        `json:"status,omitempty"`
	Progress *JSONProgress `json:"progressDetail,omitempty"`
	Id       string        `json:"id,omitempty"`
	From     string        `json:"from,omitempty"`
	Time     int64         `json:"time,omitempty"`
	Error    *JSONError    `json:"errorDetail,omitempty"`
}

func (msg *JSONMessage) Err() error {
	if msg.Error != nil {
		return fmt.Errorf(msg.Error.Message)
	}
	return nil
}

type JSONError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type JSONProgress struct {
	terminalFd uintptr
	Current    int   `json:"current,omitempty"`
	Total      int   `json:"total,omitempty"`
	Start      int64 `json:"start,omitempty"`
}

func handleJSONStream(r io.Reader, f func(msg *JSONMessage)) error {
	dec := json.NewDecoder(r)

	for {
		msg := &JSONMessage{}
		switch e := dec.Decode(&msg); e {
		case nil:
			f(msg)
		case io.EOF:
			return nil
		default:
			return e
		}
	}
}
