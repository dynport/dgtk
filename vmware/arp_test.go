package vmware

import "testing"

const arpData = `? (192.168.76.1) at 0:50:56:c0:0:8 on vmnet8 ifscope permanent [ethernet]
? (192.168.76.255) at ff:ff:ff:ff:ff:ff on vmnet8 ifscope [ethernet]
? (192.168.176.1) at 0:50:56:c0:0:1 on vmnet1 ifscope permanent [ethernet]
? (192.168.176.255) at ff:ff:ff:ff:ff:ff on vmnet1 ifscope [ethernet]
test.x (192.168.178.1) at c0:25:6:d9:2a:15 on en0 ifscope [ethernet]
? (192.168.178.34) at dc:86:d8:22:41:53 on en0 ifscope [ethernet]
? (192.168.178.41) at 0:c:29:a2:22:4d on en0 ifscope [ethernet]
? (192.168.178.255) at ff:ff:ff:ff:ff:ff on en0 ifscope [ethernet]
`

func TestParseArpData(t *testing.T) {
	ifs := ParseArpData(arpData)

	tests := []struct {
		Name     string
		Expected interface{}
		Value    interface{}
	}{
		{"len(ifs)", 8, len(ifs)},
		{"ifs[4].Hostname", "test.x", ifs[4].Hostname},
		{"ifs[0].IP", "192.168.76.1", ifs[0].IP.String()},
		{"ifd[0].Mac", "00:50:56:c0:00:08", ifs[0].MAC.String()},
		{"ifd[0].Interface", "vmnet8", ifs[0].Interface},
	}

	for _, tst := range tests {
		if tst.Expected != tst.Value {
			t.Errorf("expected %s to be %#v, was %#v", tst.Name, tst.Expected, tst.Value)
		}
	}

}
