package wunderproxy

import (
	"os/exec"
	"strings"
	"testing"
)

func TestClientConnect(t *testing.T) {
	cl := &Client{}
	h, err := cl.HostInfo()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct{ Has, Want interface{} }{
		{h.Driver, "overlay"},
	}
	for i, tc := range tests {
		if tc.Has != tc.Want {
			t.Errorf("%d: want=%#v has=%#v", i+1, tc.Want, tc.Has)
		}
	}
}

func TestClientContainer(t *testing.T) {
	name := "client-test"
	id := setupContainer(t, name)
	defer exec.Command("docker", "rm", "-f", id).Run()
	cl := &Client{}
	res, err := cl.Container(id)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct{ Has, Want interface{} }{
		{res.Id != "", true},
		{res.Image != "", true},
		//{res.Name, "/client-test"},
	}
	for i, tc := range tests {
		if tc.Has != tc.Want {
			t.Errorf("%d: want=%#v has=%#v", i+1, tc.Want, tc.Has)
		}
	}
	_ = res
}

func TestClientContainers(t *testing.T) {
	name := "client-test"
	id := setupContainer(t, name)
	defer exec.Command("docker", "rm", "-f", id).Run()
	cl := &Client{}
	list, err := cl.ListContainers(true)
	if err != nil {
		t.Fatal(err)
	}

	con := func() *Container {
		for _, c := range list {
			for _, n := range c.Names {
				if n == "/"+name {
					return c
				}
			}
		}
		return nil
	}()
	tests := []struct{ Has, Want interface{} }{
		{len(list) > 0, true},
		{con != nil, true},
		{len(con.Names), 1},
		{con.Names[0], "/client-test"},
		{con.Image, "alpine"},
	}
	for i, tc := range tests {
		if tc.Has != tc.Want {
			t.Errorf("%d: want=%#v has=%#v", i+1, tc.Want, tc.Has)
		}
	}
}

func TestClientCreateContainer(t *testing.T) {
	c := &ContainerConfig{
		Cmd:   []string{"sh", "-c", "env; sleep 3600"},
		Image: "alpine",
	}
	c.Volumes = map[string]struct{}{
		"/dev/log": {},
	}
	c.ExposedPorts = map[Port]struct{}{
		"9292/tcp": {},
	}
	c.Env = []string{"HELLO=world"}
	c.HostConfig = &HostConfig{
		Binds: []string{"/dev/log:/dev/log"},
		PortBindings: map[Port][]PortBinding{
			"9292/tcp": []PortBinding{
				{},
			},
		},
	}

	cl := &Client{}
	id, err := cl.CreateContainer(c, "client-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("created container %q", id)
	defer dockerExec("rm", "-f", id)

	err = cl.StartContainer(id)
	if err != nil {
		t.Fatal(err)
	}
	port, err := cl.exposedContainerPort(id, "9292/tcp")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct{ Has, Want interface{} }{
		{port > 1, true},
	}
	for i, tc := range tests {
		if tc.Has != tc.Want {
			t.Errorf("%d: want=%#v has=%#v", i+1, tc.Want, tc.Has)
		}
	}
}

func setupContainer(t *testing.T, name string) (id string) {
	b, err := exec.Command("docker", "run", "-d", "--name", name, "alpine", "sh", "-c", "sleep 3600").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(b))
}
