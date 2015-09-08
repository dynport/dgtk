package jenkins

import (
	"fmt"
	"io"
	"io/ioutil"
)

func (c *Client) UpdateJob(name string, body io.Reader) error {
	u := c.Address + "/job/" + name + "/config.xml"
	rsp, err := c.client().Post(u, "application/xml", body)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	b, _ := ioutil.ReadAll(rsp.Body)
	if rsp.Status[0] != '2' {
		return fmt.Errorf("got status %s but expected 2x. body=%s", rsp.Status, string(b))
	}
	return nil
}
