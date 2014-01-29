package main

import (
	"bytes"
	"fmt"
	"github.com/dynport/gocloud/aws/s3"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
)

type S3Resource struct {
	Client  *s3.Client
	Bucket  string
	Request *http.Request
}

func (resource *S3Resource) Open() (io.Reader, int64, error) {
	rsp, e := resource.Client.Get(resource.Bucket, resource.key())
	if e != nil {
		return nil, 0, e
	}
	if rsp.Status[0] != '2' {
		return nil, 0, wrongStatus("2xx", rsp.Status)
	}
	i, e := strconv.ParseInt(rsp.Header.Get("Content-Length"), 10, 64)
	if e != nil {
		return nil, 0, e
	}
	return rsp.Body, i, nil
}

func wrongStatus(expected string, found string) error {
	return fmt.Errorf("expected status of %s but got %s", expected, found)
}

func (resource *S3Resource) DockerSize() (int64, error) {
	rsp, e := resource.Client.Head(resource.Bucket, path.Dir(resource.key())+"/layer")
	if e != nil {
		return 0, e
	}
	if rsp.Status[0] != '2' {
		return 0, fmt.Errorf("expected status of 2xx but got", rsp.Status)
	}
	return strconv.ParseInt(rsp.Header.Get("Content-Length"), 10, 64)
}

func (resource *S3Resource) Exists() bool {
	rsp, e := resource.Client.Head(resource.Bucket, resource.key())
	if e != nil {
		return false
	}
	defer rsp.Body.Close()
	return rsp.Status[0] == '2'
}

func (resource *S3Resource) Tags() (map[string]string, error) {
	res, e := resource.Client.ListBucketWithOptions(resource.Bucket, &s3.ListBucketOptions{
		Prefix: resource.key(),
	})
	if e != nil {
		return nil, e
	}
	m := map[string]string{}
	for _, key := range res.Contents {
		rsp, e := resource.Client.Get(resource.Bucket, key.Key)
		if e != nil {
			return nil, e
		}
		if rsp.Status[0] != '2' {
			return nil, fmt.Errorf("expected status 2xx but got %s", rsp.Status)
		}
		b, e := ioutil.ReadAll(rsp.Body)
		if e != nil {
			return nil, e
		}
		m[path.Base(key.Key)] = strings.Replace(string(b), `"`, "", -1)
	}
	return m, nil
}

func (resource *S3Resource) Store() error {
	opts := &s3.PutOptions{
		ServerSideEncryption: true,
		ContentType:          resource.Request.Header.Get("Content-Type"),
	}
	buf := bytes.NewBuffer(make([]byte, 0, s3.MinPartSize))
	_, e := io.CopyN(buf, resource.Request.Body, s3.MinPartSize)
	if e == io.EOF {
		// less than min multipart size => direct upload
		return resource.Client.Put(resource.Bucket, resource.key(), buf.Bytes(), opts)
	} else if e != nil {
		return e
	}
	mr := io.MultiReader(buf, resource.Request.Body)

	mo := &s3.MultipartOptions{
		PartSize: 5 * 1024 * 1024,
		Callback: func(res *s3.UploadPartResult) {
			if res.Error != nil {
				logger.Print("ERROR: " + e.Error())
			} else if res.Part != nil {
				logger.Printf("uploaded: %03d (%s) %d", res.Part.PartNumber, res.Part.ETag, res.CurrentSize)
			}
		},
		PutOptions: opts,
	}
	_, e = resource.Client.PutMultipart(resource.Bucket, resource.key(), mr, mo)
	return e
}

func (resource *S3Resource) key() string {
	return strings.TrimPrefix(resource.Request.URL.String(), "/")
}
