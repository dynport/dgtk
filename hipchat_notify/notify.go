package main

import (
	"os"
	"os/exec"
)

type sendNotification struct {
	HipchatClient

	Message string   `cli:"arg required"`
	Command []string `cli:"arg required"`
}

func (act *sendNotification) Run() (e error) {
	cmd := exec.Command(act.Command[0], act.Command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	e = cmd.Run()
	switch e {
	case nil:
		e = act.Send(&HipchatNotification{Message: "[Success] " + act.Message, Color: "green"})
	default:
		e = act.Send(&HipchatNotification{Message: "[Failed] " + act.Message + "\n" + e.Error(), Color: "red"})
	}

	return e
}
