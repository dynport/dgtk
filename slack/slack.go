package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	WebhookURL      string `json:"webhook_url,omitempty"`
	DefaultChannel  string `json:"channel,omitempty"`
	DefaultUsername string `json:"username,omitempty"`
	DefaultEmoji    string `json:"emoji,omitempty"`
}

func (c *Client) SendInfo(format string, args ...interface{}) error {
	return c.Send(&Notification{
		Text: fmt.Sprintf(format, args...),
	})
}

func (c *Client) SendSuccess(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return c.Send(&Notification{
		Attachments: []Attachment{
			{
				Fallback: "Error: " + msg,
				Text:     msg,
				Color:    "good",
			},
		},
	})
}

func (c *Client) SendWarning(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return c.Send(&Notification{
		Attachments: []Attachment{
			{
				Fallback: "Error: " + msg,
				Text:     msg,
				Color:    "warning",
			},
		},
	})
}

func (c *Client) SendError(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return c.Send(&Notification{
		Attachments: []Attachment{
			{
				Fallback: "Error: " + msg,
				Text:     msg,
				Color:    "danger",
			},
		},
	})
}

func (c *Client) Send(n *Notification) error {
	if n.Channel == "" && c.DefaultChannel != "" {
		n.Channel = c.DefaultChannel
	}

	if n.Username == "" && c.DefaultUsername != "" {
		n.Username = c.DefaultUsername
	}

	if n.Emoji == "" && c.DefaultEmoji != "" {
		n.Emoji = c.DefaultEmoji
	}

	buf, err := json.Marshal(n)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.WebhookURL, "application/json", bytes.NewReader(buf))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	output, _ := ioutil.ReadAll(resp.Body)
	if resp.Status[0] != '2' {
		return fmt.Errorf("slack notify: expected status 2xx, got %s: %q", resp.Status, string(output))
	}

	return nil
}

type Notification struct {
	Text        string       `json:"text"`
	Channel     string       `json:"channel,omitempty"`
	Username    string       `json:"username,omitempty"`
	Emoji       string       `json:"icon_emoji,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Fallback string  `json:"fallback"`
	Text     string  `json:"text,omitempty"`
	PreText  string  `json:"pretext,omitempty"`
	Color    string  `json:"color,omitempty"`
	Fields   []Field `json:"fields,omitempty"`
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short,omitempty"`
}
