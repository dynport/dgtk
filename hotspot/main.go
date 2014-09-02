package main

import (
	"bytes"
	"fmt"

	"github.com/dynport/dgtk/cli"
)

func main() {
	router := cli.NewRouter()
	router.RegisterFunc("status", status, "Status")
	router.RegisterFunc("connect", connect, "Connect")
	router.RegisterFunc("disconnect", disconnect, "Disconnect")
	switch e := router.RunWithArgs(); e {
	case nil, cli.ErrorHelpRequested, cli.ErrorNoRoute:
		// ignore
		return
	default:
		logger.Fatal(e)
	}
}

func disconnect() error {
	b, e := newBrowser()
	if e != nil {
		return e
	}
	ok, e := online()
	if e != nil {
		return e
	}
	if !ok {
		logger.Printf("you are not online!")
	}

	e = b.Visit(disconnectUrl)
	if e != nil {
		return e
	}
	logger.Printf("you are disconnected now")
	return nil
}

func connect() error {
	b, e := newBrowser()
	if e != nil {
		return e
	}

	if ok, e := online(); e == nil && ok {
		logger.Printf(yellow("you are already connected!"))
		return nil
	}
	logger.Printf("checking for availablity")
	e = b.Visit(railControllerUrl)
	if e != nil {
		return e
	}
	body, e := b.Body()
	if e != nil {
		return e
	}
	if !bytes.Contains(body, []byte("HotSpot verfügbar")) {
		return fmt.Errorf("hotspot not available")
	}
	logger.Printf("hotspot available")

	e = b.Visit(connectUrl)
	if e != nil {
		return e
	}
	forms, e := b.Forms()
	if e != nil {
		return e
	}
	if len(forms) != 1 {
		return fmt.Errorf("expected to find 1 form, found %d", len(forms))
	}
	form := forms[0]

	login, e := getEnv(envLogin)
	if e != nil {
		return e
	}
	pwd, e := getEnv(envPwd)
	if e != nil {
		return e
	}

	e = form.FillIn("username", login)
	if e != nil {
		return e
	}

	e = form.FillIn("password", pwd)
	if e != nil {
		return e
	}

	e = b.Submit(form)
	if e != nil {
		return e
	}

	ok, e := online()
	if e != nil {
		return e
	}
	if ok {
		logger.Printf("you are online now!")
	} else {
		logger.Printf("seems that you are not online")
	}
	return nil
}

func status() error {
	if ok, e := online(); e == nil && ok {
		logger.Printf(green("you are online!"))
		return nil
	} else {
		logger.Printf(red("you are offline"))
	}
	return nil
}
