package main

import (
	"fmt"
	"github.com/dynport/dgtk/vmware"
	"github.com/dynport/gossh"
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
	"log"
	"net"
	_ "os/exec"
	"time"
)

type Test struct {
}

func (test *Test) Run() error {
	started := time.Now()
	defer func() { log.Printf("finished in %.03f", time.Since(started).Seconds()) }()
	log.Println("running test")
	vm, e := findFirst("ubuntu-13.04")
	if e != nil {
		return e
	}
	clone, e := vmware.Create(vm, "ssh")
	if e != nil {
		return e
	}
	//defer func() { log.Println("deleting vm"); clone.Delete() }()
	e = clone.Start()
	if e != nil {
		return e
	}
	//defer func() { log.Println("stopping vm"); clone.Stop() }()
	started = time.Now()
	ip := ""
	e = waitFor(100*time.Millisecond, 30*time.Second, func() bool {
		vmx, e := clone.Vmx()
		if e != nil {
			return false
		}
		leases, e := vmware.AllLeases()
		if e != nil {
			return false
		}
		lease := leases.Lookup(vmx.MacAddress)
		if lease != nil && lease.Ip != "" {
			ip = lease.Ip
			c, e := net.DialTimeout("tcp", ip+":22", 10*time.Millisecond)
			if e != nil {
				return false
			}
			defer c.Close()
			log.Printf("got ip %s", ip)
			return true
		}
		return false

	})
	if e != nil {
		return e
	}
	log.Printf("got ip %q in %.03f", ip, time.Since(started).Seconds())
	l, e := urknall.OpenStdoutLogger()
	if e != nil {
		return e
	}
	defer l.Close()
	e = doUrknallTest(ip)
	if e != nil {
		return e
	}
	if e := clone.Stop(); e != nil {
		return e
	}
	return clone.Delete()
}

func doUrknallTest(ip string) error {
	host := &urknall.Host{
		IP:       ip,
		User:     "ubuntu",
		Password: "ubuntu",
	}
	host.AddCommands("update", cmd.UpdatePackages())
	host.AddCommands("packages", cmd.InstallPackages("build-essential", "curl", "tmux", "pv", "dnsutils", "mosh", "ntp", "vim-nox"))
	host.AddCommands("error", "rgne")
	return host.Provision(nil)
}

func doTest(ip string) error {
	client := gossh.New(ip, "ubuntu")
	client.SetPassword("ubuntu")
	defer func() { log.Println("stopping client"); client.Close() }()
	client.DebugWriter = func(v ...interface{}) {
		log.Println("DEBUG: " + fmt.Sprint(v...))
	}
	client.ErrorWriter = func(v ...interface{}) {
		log.Println("ERROR: " + fmt.Sprint(v...))
	}
	_, e := client.Execute(execCmd)
	return e
}

type Writer struct {
	Prefix string
}

func (writer *Writer) Write(b []byte) (int, error) {
	log.Printf("%s", string(b))
	return len(b), nil
}

//

const execCmd = `
mkfifo /tmp/stdout
(while read line </tmp/stdout; do echo "stdout $line"; done) &
trap "rm -f /tmp/stdout" INT TERM EXIT

mkfifo /tmp/stderr
(while read line </tmp/stderr; do echo "stderr $line"; done) &
trap "rm -f /tmp/stderr" INT TERM EXIT

sudo bash <<EOF_SUDO 1>/tmp/stdout 2>/tmp/stderr

set -e

# script
bash <<EOF
set -ex -o pipefail
apt-get update 
apt-get upgrade 
apt-get install -y build-essential curl tmux pv dnsutils mosh ntp vim-nox
EOF
# end script
EOF_SUDO
`

const cmdOld = `
sudo bash <<EOF_SUDO
set -e

mkfifo /tmp/stdout
(while read line </tmp/stdout; do echo "stdout $line"; done) &
trap "rm -f /tmp/stdout" INT TERM EXIT

mkfifo /tmp/stderr
(while read line </tmp/stderr; do echo "stderr $line"; done) &
trap "rm -f /tmp/stderr" INT TERM EXIT

# script
bash <<EOF 1>/tmp/stdout 2>/tmp/stderr
set -ex -o pipefail
apt-get update 
apt-get upgrade 
apt-get install -y build-essential curl tmux pv dnsutils mosh ntp vim-nox
EOF
# end script
EOF_SUDO
`

func waitFor(sleepDuration, timeoutDuration time.Duration, f func() bool) error {
	ticker := time.NewTicker(sleepDuration)
	timeout := time.NewTimer(timeoutDuration)
	for {
		select {
		case <-ticker.C:
			fmt.Print(".")
			if f() {
				return nil
			}
		case <-timeout.C:
			return fmt.Errorf("timeout")
		}
	}
}

func init() {
	router.Register("test", &Test{})
}
