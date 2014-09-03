package main

type testNotification struct {
	HipchatClient

	Message string `cli:"arg required"`
}

func (act *testNotification) Run() (e error) {
	return act.Send(&HipchatNotification{Message: act.Message, Color: "gray"})
}
