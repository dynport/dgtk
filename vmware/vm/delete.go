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
	if vm.Running() {
		logger.Printf("vm is running, stopping")
		e = vm.Stop()
		if e != nil {
			return e
		}
	}
	return vm.Delete()
}
