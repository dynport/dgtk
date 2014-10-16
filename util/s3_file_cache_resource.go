package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/dynport/gocloud/aws/s3"
)

func OpenS3Key(bucket, key string, srcCompressed bool) (io.ReadCloser, error) {
	return (&S3Resource{Key: key, Bucket: bucket, SrcCompressed: srcCompressed}).Cacher().Open()
}

type S3Resource struct {
	Key           string
	Bucket        string
	SrcCompressed bool
}

func (s *S3Resource) Cacher() *FileCache {
	return &FileCache{Key: "s3/" + s.Bucket + "/" + s.Key, Source: S3Reader(s.Bucket, s.Key), SrcCompressed: s.SrcCompressed}
}

func S3Reader(bucket, key string) Source {
	return &s3Source{bucket: bucket, key: key}
}

type s3Source struct {
	bucket         string
	key            string
	cachedResponse *http.Response
}

func (s *s3Source) Size() (int64, error) {
	if s.cachedResponse == nil {
		return 0, fmt.Errorf("not opened yet")
	}
	return strconv.ParseInt(s.cachedResponse.Header.Get("Content-Length"), 10, 64)
}

func s3Client() *s3.Client {
	client := s3.NewFromEnv()
	client.CustomEndpointHost = "s3-eu-west-1.amazonaws.com"
	return client
}

func (s *s3Source) Open() (io.Reader, error) {
	if s.cachedResponse == nil {
		var e error
		client := s3Client()
		logger.Printf("using client %#v", client)
		logger.Printf("opening key=%s bucket=%s", s.key, s.bucket)
		s.cachedResponse, e = client.Get(s.bucket, s.key)
		if e != nil {
			var b []byte
			if s.cachedResponse != nil {
				var err error
				b, err = ioutil.ReadAll(s.cachedResponse.Body)
				if err != nil {
					return nil, err
				}
			}
			return nil, fmt.Errorf("error getting key: %s: %s", e, b)
		}
		if s.cachedResponse.Status[0] != '2' {
			b, err := ioutil.ReadAll(s.cachedResponse.Body)
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("expected status 2xx, got %s: %s", s.cachedResponse.Status, b)
		}
	}
	return s.cachedResponse.Body, nil
}
