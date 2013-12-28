package vmware

import (
	"bufio"
	"os"
	"sort"
	"strings"
	"time"
)

const leasesPath = "/private/var/db/vmware/vmnet-dhcpd-vmnet8.leases"

func AllLeases() (Leases, error) {
	return ParseLeases(leasesPath)
}

type Lease struct {
	Ip               string
	HardwareEthernet string
	Starts           time.Time
	Ends             time.Time
}

type Leases []*Lease

func (leases Leases) Lookup(addr string) *Lease {
	filtered := leases.FindHardwareAddress(addr)
	sort.Sort(filtered)
	if len(filtered) > 0 {
		return filtered[len(filtered)-1]
	}
	return nil
}

func (leases Leases) FindHardwareAddress(addr string) Leases {
	out := Leases{}
	for _, l := range leases {
		if l.HardwareEthernet == addr {
			out = append(out, l)
		}
	}
	return out
}

func (leases Leases) Len() int {
	return len(leases)
}

func (leases Leases) Swap(a, b int) {
	leases[a], leases[b] = leases[b], leases[a]
}

func (leases Leases) Less(a, b int) bool {
	return leases[a].Starts.Before(leases[b].Starts)
}

func ParseLeases(path string) (leases Leases, e error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var lease *Lease
	for scanner.Scan() {
		parts := strings.Fields(strings.TrimSpace(scanner.Text()))
		if len(parts) > 1 {
			switch parts[0] {
			case "lease":
				if lease != nil {
					leases = append(leases, lease)
				}
				lease = &Lease{Ip: parts[1]}
			case "hardware":
				if lease != nil {
					lease.HardwareEthernet = strings.TrimSuffix(parts[2], ";")
				}
			case "starts":
				if lease != nil {
					t, e := time.Parse("2006/01/02 15:04:05;", parts[2]+" "+parts[3])
					if e == nil {
						lease.Starts = t
					}
				}
			case "ends":
				if lease != nil {
					t, e := time.Parse("2006/01/02 15:04:05;", parts[2]+" "+parts[3])
					if e == nil {
						lease.Ends = t
					}
				}
			}
		}
	}
	if lease != nil {
		leases = append(leases, lease)
	}
	return leases, nil
}
