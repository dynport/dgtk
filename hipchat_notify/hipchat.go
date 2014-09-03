package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

type HipchatClient struct {
	ApiKey   string `json:"-" cli:"opt --api-key required"`
	RoomName string `json:"-" cli:"opt --room required"`
	Sender   string `json:"-" cli:"opt --sender default=hc_notify"`
}

func (c *HipchatClient) Send(n *HipchatNotification) error {
	url := "https://api.hipchat.com/v1/rooms/message?auth_token=" + c.ApiKey
	s := fmt.Sprintf("room_id=%s&message=%s&from=%s&format=%s&color=%s&message_format=%s&notify=%t", c.RoomName, n.Message, c.Sender, "json", n.Color, "text", n.Notify)
	b := bytes.NewBufferString(s)

	r, err := http.Post(url, "application/x-www-form-urlencoded", b)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	output, _ := ioutil.ReadAll(r.Body)
	if r.Status[0] != '2' {
		return fmt.Errorf("expected status 2xx, got %s: %q", r.Status, string(output))
	}
	return nil
}

type HipchatNotification struct {
	Message string `json:"message"`
	Color   string `json:"color"`
	Notify  bool   `json:"notify,omitempty"`
}
