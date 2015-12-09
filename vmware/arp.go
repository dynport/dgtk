package vmware

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
)

type NetworkInterface struct {
	Hostname  string
	Interface string
	IP        net.IP
	MAC       net.HardwareAddr
}

const (
	fieldName      = 0
	fieldIP        = 1
	fieldMac       = 3
	fieldInterface = 5
)

func LoadArpInterfaces() ([]*NetworkInterface, error) {
	b, err := exec.Command("arp", "-a").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error calling %q: %s (%s)", "arp -a", err, string(b))
	}
	return ParseArpData(string(b)), nil
}

func ParseArpData(s string) (ifs []*NetworkInterface) {
	lines := strings.Split(s, "\n")
	for _, l := range lines {
		inf := &NetworkInterface{}
		fields := strings.Fields(l)
		for i, v := range fields {
			switch i {
			case fieldName:
				if v != "?" {
					inf.Hostname = v
				}
			case fieldIP:
				if ip := net.ParseIP(strings.TrimSuffix(strings.TrimPrefix(v, "("), ")")); ip != nil {
					inf.IP = ip
				}
			case fieldMac:
				v := normalizeMac(v)
				if mac, err := net.ParseMAC(v); err == nil {
					inf.MAC = mac
				}
			case fieldInterface:
				inf.Interface = v
			}
		}

		if inf.MAC != nil && inf.IP != nil {
			ifs = append(ifs, inf)
		}
	}
	return ifs
}

func normalizeMac(in string) string {
	parts := strings.Split(in, ":")
	for i, p := range parts {
		if len(p) == 1 {
			parts[i] = "0" + p
		}
	}
	return strings.Join(parts, ":")
}
