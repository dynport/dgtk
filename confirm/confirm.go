package confirm

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/dynport/gocli"
)

func ConfirmShell(actions ...*Action) error {
	if len(actions) == 0 {
		return nil
	}
	fmt.Print("\033[H\033[2J")
	for _, a := range actions {
		fmt.Println(colorizeTitle(a))
		if len(a.Diff) > 0 {
			fmt.Println(string(a.Diff))
		}
		fmt.Println(strings.Repeat("-", 100))
	}
	fmt.Printf("confirm? ")
	if readConfirm(os.Stdin) {
		for _, a := range actions {
			if err := a.Call(); err != nil {
				return err
			}
		}
	}
	return nil
}

var colorMap = map[Type]func(string) string{
	TypeCreate: gocli.Green,
	TypeDelete: gocli.Red,
	TypeUpdate: gocli.Yellow,
}

func colorizeTitle(a *Action) string {
	s := a.Title
	if f, ok := colorMap[a.Type]; ok {
		s = f(s)
	}
	return s
}

func readConfirm(in io.Reader) bool {
	t, _ := bufio.NewReader(in).ReadString('\n')
	return strings.TrimSpace(strings.ToLower(t)) == "confirm"
}

func ConfirmHTML(actions ...*Action) error {
	l := log.New(os.Stderr, "", 0)
	if len(actions) == 0 {
		l.Printf("nothing changed")
		return nil
	}

	closer, err := WebDiff(actions)
	if err != nil {
		return err
	}
	doContinue := <-closer
	if doContinue {
		for _, a := range actions {
			if err := a.Call(); err != nil {
				return err
			}
		}
	}
	return nil
}
