package rds

import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/dynport/dgtk/wunderproxy/wunderproxy/Godeps/_workspace/src/github.com/dynport/gocloud/aws"
	"github.com/dynport/dgtk/wunderproxy/wunderproxy/Godeps/_workspace/src/github.com/dynport/gocloud/aws/ec2"
)

type Client struct {
	*aws.Client
	CustomRegion string
}

func (client *Client) Endpoint() string {
	prefix := "https://rds"
	if client.Client.Region != "" {
		prefix += "." + client.Client.Region
	}
	return prefix + ".amazonaws.com"
}

func NewFromEnv() *Client {
	return &Client{Client: aws.NewFromEnv()}
}

type DescribeDBInstances struct {
	DBInstanceIdentifier string
	Filters              []*ec2.Filter
	Marker               string
	MaxRecords           int
}

func newAction(name string) url.Values {
	return url.Values{"Action": {name}, "Version": {Version}}
}

type Error struct {
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
}

func (e Error) Error() string {
	return e.Code + ": " + e.Message
}

type ErrorResponse struct {
	XMLName   xml.Name `xml:"ErrorResponse"`
	RequestID string   `xml:"RequestID"`
	Error     Error    `xml:"Error"`
}

func (client *Client) loadResource(method string, url string, r io.Reader, i interface{}) error {
	dbg.Printf("executing method=%s to url=%s", method, url)
	req, e := http.NewRequest(method, url, r)
	if e != nil {
		return e
	}
	client.SignAwsRequestV2(req, time.Now())
	rsp, e := http.DefaultClient.Do(req)
	if e != nil {
		return e
	}
	defer rsp.Body.Close()
	dbg.Printf("got status %s", rsp.Status)

	b, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		return e
	}
	dbg.Printf("got response %s", b)

	if rsp.Status[0] != '2' {
		resp := &ErrorResponse{}
		if e = xml.Unmarshal(b, resp); e != nil {
			return e
		}
		return resp.Error
	}
	return xml.Unmarshal(b, i)
}
