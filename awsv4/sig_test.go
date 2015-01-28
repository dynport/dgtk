package awsv4

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestSign(t *testing.T) {
	tim := time.Date(2006, 1, 2, 3, 15, 4, 5, time.UTC)
	c := &Context{
		Service: "s3", Region: "eu-west-1",
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
		CurrentTime:     tim,
	}

	req, err := http.NewRequest("GET", "https://just.a/test?Method=get&Version=1", nil)
	if err != nil {
		t.Fatal("error creating request", err)
	}
	c.Sign(req)

	v := req.URL.Query()
	tests := []struct {
		Name     string
		Value    string
		Expected string
	}{
		{"Method", v.Get("Method"), "get"},
		{"Authorization", req.Header.Get("Authorization"), "AWS4-HMAC-SHA256 Credential=access/20060102/eu-west-1/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date, Signature=6e80395ae7088abbed31a6ac9dcd1d39445e75f7b920bf3cdb02a138101bf8c4"},
		{"X-Amz-Date", req.Header.Get("X-Amz-Date"), "20060102T031504Z"},
		{"X-Amz-Content-Sha256", req.Header.Get("X-Amz-Content-Sha256"), "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
	}
	for _, tst := range tests {
		if tst.Expected != tst.Value {
			t.Errorf("expected %s to be %#v, was %#v", tst.Name, tst.Expected, tst.Value)
		}
	}
}

func TestSignWithPayload(t *testing.T) {
	tim := time.Date(2006, 1, 2, 3, 15, 4, 5, time.UTC)
	c := &Context{
		Service: "s3", Region: "eu-west-1",
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
		CurrentTime:     tim,
	}

	req, err := http.NewRequest("POST", "https://just.a/test?Method=get&Version=1", nil)
	if err != nil {
		t.Fatal("error creating request", err)
	}

	req.Body = ioutil.NopCloser(strings.NewReader("this is a test"))
	c.Sign(req)

	b, err := ioutil.ReadAll(req.Body)

	v := req.URL.Query()
	tests := []struct {
		Name     string
		Value    string
		Expected string
	}{
		{"Method", v.Get("Method"), "get"},
		{"Body", string(b), "this is a test"},
		{"Authorization", req.Header.Get("Authorization"), "AWS4-HMAC-SHA256 Credential=access/20060102/eu-west-1/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date, Signature=f22bb29d143bf9b36ae9a4307a8944991740a922ed7d2ef1955900c7bc535c53"},
		{"X-Amz-Date", req.Header.Get("X-Amz-Date"), "20060102T031504Z"},
		{"X-Amz-Content-Sha256", req.Header.Get("X-Amz-Content-Sha256"), "2e99758548972a8e8822ad47fa1017ff72f06f3ff6a016851f45c398732bc50c"},
	}
	for _, tst := range tests {
		if tst.Expected != tst.Value {
			t.Errorf("expected %s to be %#v, was %#v", tst.Name, tst.Expected, tst.Value)
		}
	}
}
