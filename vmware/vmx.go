package vmware

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Vmx struct {
	MacAddress    string
	CleanShutdown bool
	SoftPowerOff  bool
	Memory        int
	Cpus          int
}

func (vmx *Vmx) Parse(path string) error {
	f, e := os.Open(path)
	if e != nil {
		return e
	}
	defer f.Close()
	scan := bufio.NewScanner(f)
	vmx.Cpus = 1
	for scan.Scan() {
		parts := strings.SplitN(scan.Text(), " = ", 2)
		if len(parts) == 2 {
			switch parts[0] {
			case "ethernet0.generatedAddress":
				vmx.MacAddress = strings.Replace(parts[1], `"`, "", -1)
			case "cleanShutdown":
				vmx.CleanShutdown = parts[1] == `"TRUE"`
			case "softPowerOff":
				vmx.SoftPowerOff = parts[1] == `"TRUE"`
			case "memsize":
				vmx.Memory, e = strconv.Atoi(strings.Replace(parts[1], `"`, "", -1))
				if e != nil {
					return e
				}
			case "numvcpus":
				vmx.Cpus, e = strconv.Atoi(strings.Replace(parts[1], `"`, "", -1))
				if e != nil {
					return e
				}
			}

		}
	}
	return scan.Err()
}
