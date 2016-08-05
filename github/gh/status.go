package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/dynport/gocli"
)

type Status struct {
	Ref string `cli:"opt --ref default=master"`
}

func (r *Status) Run() error {
	repo, err := githubRepo()
	if err != nil {
		return err
	}
	cl, err := client()
	if err != nil {
		return err
	}
	rsp, err := cl.Get("https://api.github.com/repos/" + repo + "/commits/" + r.Ref + "/status")
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		b, _ := ioutil.ReadAll(rsp.Body)
		return fmt.Errorf("got status %s but expected 2x. body=%s", rsp.Status, string(b))
	}
	var res *statusResponse
	err = json.NewDecoder(rsp.Body).Decode(&res)
	if err != nil {
		return err
	}
	color := gocli.Green
	switch res.State {
	case "success":
		color = gocli.Green
	default:
		color = gocli.Red
	}
	fmt.Println(color(res.State))
	return nil
}

type statusResponse struct {
	State string `json:"state"`
}
