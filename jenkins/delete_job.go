package jenkins

import (
	"fmt"
	"io/ioutil"
)

func (c *Client) DeleteJob(name string) error {
	rsp, err := c.client().Post(c.Address+"/job/"+name+"/doDelete", "application/xml", nil)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode > 400 {
		b, _ := ioutil.ReadAll(rsp.Body)
		return fmt.Errorf("got status %s but expected < 400. body=%s", rsp.Status, string(b))
	}
	return nil
}
