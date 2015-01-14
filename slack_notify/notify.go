package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/dynport/dgtk/slack"
)

type sendNotification struct {
	Message string   `cli:"arg required"`
	Command []string `cli:"arg required"`

	slackClient *slack.Client
}

func (act *sendNotification) Run() error {
	err := act.initSlack()
	if err != nil {
		return err
	}

	cmd := exec.Command(act.Command[0], act.Command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	switch err = cmd.Run(); err {
	case nil:
		return act.slackClient.SendSuccess("[Success] %s", act.Message)
	default:
		return act.slackClient.SendError("[Failed] %s\n%s", act.Message, err)
	}
}

func (act *sendNotification) initSlack() error {
	cfPath := os.ExpandEnv("${HOME}/.slack.conf")
	if cfPath == "" {
		return fmt.Errorf("failed to read ${HOME} from env")
	}

	fh, err := os.Open(cfPath)
	if err != nil {
		return fmt.Errorf("failed to read slack config file: %s", err)
	}
	defer fh.Close()

	err = json.NewDecoder(fh).Decode(&act.slackClient)
	if err != nil {
		return fmt.Errorf("failed to decode the slack config file: %s", err)
	}

	if act.slackClient.WebhookURL == "" {
		return fmt.Errorf("no slack webhook URL configured")
	}

	return nil
}
