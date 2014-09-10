package vmware

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func (vm *Vm) ModifyMemory(mem int) error {
	return vm.modify("memsize", mem)
}

func (vm *Vm) ModifyCpu(cpus int) error {
	return vm.modify("numvcpus", cpus)
}

func (vm *Vm) modify(key string, value int) error {
	out := []string{}
	f, e := os.Open(vm.Path)
	if e != nil {
		return e
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	replaced := false
	repl := fmt.Sprintf("%s = %q", key, strconv.Itoa(value))
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), key+" = ") {
			replaced = true
			out = append(out, repl)
		} else {
			out = append(out, scanner.Text())
		}
	}

	if !replaced {
		out = append(out, repl)
	}

	if e := scanner.Err(); e != nil {
		return e
	}
	return ioutil.WriteFile(vm.Path, []byte(strings.Join(out, "\n")), 0644)
}
