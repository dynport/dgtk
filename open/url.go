package open

import (
	"fmt"
	"io/ioutil"
	"os/exec"
)

func URL(s string) error {
	for _, name := range []string{"xdg-open", "open"} {
		if path, err := exec.LookPath(name); err == nil {
			s := exec.Command(path, s)
			s.Stdout = ioutil.Discard
			s.Stderr = ioutil.Discard
			return s.Start()
		}
	}
	return fmt.Errorf("no open command found")
}
