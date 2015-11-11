package open

import (
	"fmt"
	"os/exec"
)

func URL(s string) error {
	for _, name := range []string{"xdg-open", "open"} {
		if path, err := exec.LookPath(name); err == nil {
			return exec.Command(path, s).Run()
		}
	}
	return fmt.Errorf("no open command found")
}
