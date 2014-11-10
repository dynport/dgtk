package main

type Delete struct {
	Name string `cli:"type=arg required=true"`
}

func (action *Delete) Run() error {
	logger.Printf("deleting vm %s", action.Name)
	vm, e := findFirst(action.Name)
	if e != nil {
		return e
	}
	r, err := vm.Running()
	if err != nil {
		return err
	} else if r {
		logger.Printf("vm is running, stopping")
		e = vm.Stop()
		if e != nil {
			logger.Printf("ERROR=%q", e)
		}
	}
	return vm.Delete()
}
