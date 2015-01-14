package vmware

import (
	"testing"
)

func TestParseLeases(t *testing.T) {
	leases, err := ParseLeases("fixtures/vmnet-dhcpd-vmnet8.leases")
	if err != nil {
		t.Fatal("error parsing leases", err)
	}

	tests := []struct {
		Name     string
		Expected interface{}
		Value    interface{}
	}{
		{"len(leases)", 8, len(leases)},
		{"leases[0].HardwareEthernet", "00:0c:29:80:9a:df", leases[0].HardwareEthernet},
		{"leases[0].Starts", "2013-12-28T17:02", leases[0].Starts.Format("2006-01-02T15:04")},
		{"leases[0].Ends", "2013-12-28T17:32", leases[0].Ends.Format("2006-01-02T15:04")},
	}

	for _, tst := range tests {
		if tst.Expected != tst.Value {
			t.Errorf("expected %s to be %#v, was %#v", tst.Name, tst.Expected, tst.Value)
		}
	}
}
